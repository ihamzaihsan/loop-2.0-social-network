'use client'
import { API } from "../../utils/api";

import { useRealtimeRefresh } from '@/app/webscoket/useRealtimeRefresh'

import { useRouter, useSearchParams } from 'next/navigation'
import { useState, useEffect, useRef, Suspense } from 'react'
import Sidebar from '../../components/Sidebar'
import './chat.css'
import { WebSocketClient } from '../webscoket/websocket'
import { fetchFollowedUsers, fetchChatContacts, fetchMessages } from './messageHandlers'

interface User {
  id: number
  followedID?: number  // Add this optional property
  firstName: string
  lastName: string
  nickname?: string
  avatar?: string
}

interface Message {
  id: number
  sender_id: number
  receiver_id: number
  content: string
  created_at: string
  is_read: boolean
  type?: string
  sender: {
    id: number
    first_name: string
    last_name: string
    avatar?: string
  }
}

interface ChatContact {
  id: number
  firstName: string
  lastName: string
  nickname?: string
  avatar?: string
  lastMessage?: string
  lastMessageTime?: string
  unreadCount?: number
}

export default function Chat() {
  return <Suspense fallback={<div className="chat-page">Loading chat…</div>}><ChatContent /></Suspense>
}

function ChatContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const requestedContact = Number(searchParams.get('user'))
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [currentUser, setCurrentUser] = useState<User | null>(null)
  const [contacts, setContacts] = useState<ChatContact[]>([])
  const [selectedContact, setSelectedContact] = useState<ChatContact | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [newMessage, setNewMessage] = useState('')
  const [showUsersList, setShowUsersList] = useState(false)
  const [followedUsers, setFollowedUsers] = useState<User[]>([])
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const [wsClient, setWsClient] = useState<WebSocketClient | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const chatViewportRef = useRef<HTMLDivElement>(null)

  useEffect(() => { setWsClient(WebSocketClient.getInstance()) }, [])

  useEffect(() => {
    const viewport = window.visualViewport
    // Follow the visible viewport when a mobile keyboard reduces the chat area.
    const resize = () => {
      const height = viewport?.height ?? window.innerHeight
      chatViewportRef.current?.style.setProperty('--chat-viewport-height', `${height}px`)
    }
    resize()
    viewport?.addEventListener('resize', resize)
    window.addEventListener('resize', resize)
    return () => {
      viewport?.removeEventListener('resize', resize)
      window.removeEventListener('resize', resize)
    }
  }, [loading])
  
  // Add state for image modal
  const [showImageModal, setShowImageModal] = useState(false)
  const [selectedImage, setSelectedImage] = useState<string>('')

  // Add function to handle image click
  const handleImageClick = (imageSrc: string) => {
    setSelectedImage(imageSrc)
    setShowImageModal(true)
  }

  // Add function to close modal
  const closeImageModal = () => {
    setShowImageModal(false)
    setSelectedImage('')
  }

  // Add keyboard event listener for ESC key
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && showImageModal) {
        closeImageModal()
      }
    }

    if (showImageModal) {
      document.addEventListener('keydown', handleKeyDown)
      // Prevent body scroll when modal is open
      document.body.style.overflow = 'hidden'
    }

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      document.body.style.overflow = 'unset'
    }
  }, [showImageModal])

  const [onlineUsers, setOnlineUsers] = useState<number[]>([])
  const [typingUser, setTypingUser] = useState<number | null>(null)
  const typingTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)
  const selectedId = useRef<number | null>(null)
  useEffect(() => {
    const client = WebSocketClient.getInstance()
    selectedId.current = selectedContact?.id ?? null
    client.currentChatUser = selectedId.current
    return () => { client.currentChatUser = null }
  }, [selectedContact?.id])

  useEffect(() => {
    if (!currentUser || !Number.isInteger(requestedContact) || requestedContact <= 0 || requestedContact === currentUser.id || selectedId.current === requestedContact) return
    let cancelled = false
    const selectRequestedContact = async () => {
      try {
        const response = await fetch(`${API}/user/info?id=${requestedContact}`, { credentials: 'include' })
        if (!response.ok || cancelled) return
        const data = await response.json()
        if (cancelled || !data.user) return
        selectedId.current = requestedContact
        setSelectedContact(data.user)
        setShowUsersList(false)
        setNewMessage('')
      } catch (error) { console.error('Failed to open conversation:', error) }
    }
    selectRequestedContact()
    return () => { cancelled = true }
  }, [currentUser, requestedContact])

  const refreshChat = async () => {
    const contactId = selectedContact?.id
    if (contactId) {
      const history = await fetchMessages(contactId)
      if (selectedId.current === contactId) {
        setMessages(history)
        const throughId = Math.max(0, ...history.filter(message => message.sender_id === contactId && !message.is_read).map(message => message.id))
        if (throughId && document.visibilityState === 'visible') {
          await fetch(`${API}/chat/read`, {
            method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ contact_id: contactId, through_id: throughId })
          })
        }
      }
    }
    const latest = await fetchChatContacts()
    setContacts(latest)
    if (showUsersList) setFollowedUsers(await fetchFollowedUsers())
  }
  useRealtimeRefresh(['chat', 'profiles', 'social'], refreshChat, !!currentUser)
  useEffect(() => { if (currentUser) refreshChat() }, [currentUser, selectedContact?.id])

  useEffect(() => {
    const client = WebSocketClient.getInstance()
    const presence = () => setOnlineUsers([...client.onlineUsers].filter(([, online]) => online).map(([id]) => id))
    let expires: ReturnType<typeof setTimeout> | undefined
    const typing = (event: Event) => {
      const content = (event as CustomEvent).detail
      if (content.sender_id !== selectedContact?.id) return
      if (expires) clearTimeout(expires)
      setTypingUser(content.is_typing ? content.sender_id : null)
      if (content.is_typing) expires = setTimeout(() => setTypingUser(null), 3000)
    }
    presence()
    setTypingUser(null)
    window.addEventListener('presence_update', presence)
    window.addEventListener('typing_status', typing)
    return () => {
      window.removeEventListener('presence_update', presence)
      window.removeEventListener('typing_status', typing)
      if (expires) clearTimeout(expires)
    }
  }, [selectedContact?.id])

  useEffect(() => {
    if (!selectedContact) return
    const client = WebSocketClient.getInstance()
    return () => {
      if (typingTimeout.current) clearTimeout(typingTimeout.current)
      client.setTypingStatus(selectedContact.id, false)
    }
  }, [selectedContact?.id])

  const changeMessage = (value: string) => {
    setNewMessage(value)
    if (!selectedContact) return
    const client = WebSocketClient.getInstance()
    if (typingTimeout.current) clearTimeout(typingTimeout.current)
    client.setTypingStatus(selectedContact.id, !!value.trim())
    typingTimeout.current = setTimeout(() => client.setTypingStatus(selectedContact.id, false), 1500)
  }

      // Add a separate effect to handle message updates when selectedContact changes
      useEffect(() => {
        if (selectedContact) {
            // When a contact is selected, add any pending messages for this contact
            setMessages(prev => {
                // Filter messages to only show messages between current user and selected contact
                return prev.filter(msg => 
                    (msg.sender_id === currentUser?.id && msg.receiver_id === selectedContact.id) ||
                    (msg.sender_id === selectedContact.id && msg.receiver_id === currentUser?.id)
                );
            });
        }
      }, [selectedContact, currentUser]);
  useEffect(() => {
    const fetchUserData = async () => {
      try {
        const response = await fetch(`${API}/profile`, {
          method: 'GET',
          credentials: 'include'
        });

        if (!response.ok) {
          if (response.status === 401) {
            router.push('/');
            return;
          }
          throw new Error('Failed to fetch user data');
        }

        const data = await response.json();
        
        if (data.user) {
          setCurrentUser({
            id: data.user.id,
            firstName: data.user.firstName,
            lastName: data.user.lastName,
            nickname: data.user.nickname,
            avatar: data.user.avatar
          });
          
          // Fetch chat contacts
          const chatContacts = await fetchChatContacts();
          console.log('Fetched chat contacts:', chatContacts); // Add this debug line
          setContacts(chatContacts);
        }
      } catch (error: any) {
        console.error('Error fetching user data:', error);
        setError(error.message);
      } finally {
        setLoading(false);
      }
    };

    fetchUserData();
  }, [router]);


  // Handle selecting a contact
  const handleSelectContact = async (contact: ChatContact) => {
    if (selectedContact?.id !== contact.id) setNewMessage('');
    selectedId.current = contact.id;
    setSelectedContact(contact);
    
    // Fetch messages for this contact
    const contactMessages = await fetchMessages(contact.id);
    if (selectedId.current === contact.id) setMessages(contactMessages);
    setShowUsersList(false);
    
    // Mark messages as read
    if (contact.unreadCount && contact.unreadCount > 0) {
        // Update the contact to show no unread messages
        setContacts(prev => 
            prev.map(c => 
                c.id === contact.id ? { ...c, unreadCount: 0 } : c
            )
        );
    }
  };
    // Handle sending a message
    const handleSendMessage = async () => {
      if (!newMessage.trim() || !selectedContact || !currentUser) return;

      // Validate message length
      if (newMessage.length > 100) {
        console.error('Message exceeds maximum length of 100 characters');
        return;
      }

      WebSocketClient.getInstance().setTypingStatus(selectedContact.id, false);
      try {
        // Try to send via WebSocket first
        let sentViaWebSocket = false;
        if (wsClient && wsClient.socket && wsClient.socket.readyState === WebSocket.OPEN) {
          console.log('Attempting to send message via WebSocket');
          sentViaWebSocket = wsClient.sendMessage(selectedContact.id, newMessage);
          console.log('WebSocket send result:', sentViaWebSocket);
        } else {
          console.log('WebSocket not available, using HTTP');
        }
        
        // If WebSocket failed or not available, use HTTP
        if (!sentViaWebSocket) {
          console.log('Sending message via HTTP');
          const response = await fetch(`${API}/messages`, {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              receiver_id: selectedContact.id,
              content: newMessage,
              type: 'text'
            }),
          });

          if (!response.ok) {
            throw new Error('Failed to send message via HTTP');
          }
          
          const data = await response.json();
          console.log('HTTP response:', data);
        }
      
        // Add the message to the UI regardless of how it was sent
        const newMessageObj = {
          id: Date.now(), // Temporary ID until we get the real one
          sender_id: currentUser.id,
          receiver_id: selectedContact.id,
          content: newMessage,
          type: 'text',
          created_at: new Date().toISOString(),
          is_read: false,
          sender: {
            id: currentUser.id,
            first_name: currentUser.firstName,
            last_name: currentUser.lastName,
            avatar: currentUser.avatar
          }
        };
      
        // Add the new message to the messages list
        setMessages(prev => [...prev, newMessageObj]);
      
        // Update the contact's last message
        setContacts(prev => {
          const updatedContacts = [...prev];
          const contactIndex = updatedContacts.findIndex(c => c.id === selectedContact.id);
        
          if (contactIndex >= 0) {
            updatedContacts[contactIndex] = {
              ...updatedContacts[contactIndex],
              lastMessage: newMessage.match(/\.(jpeg|jpg|gif|png)$/i) ? '📷 Image' : newMessage,
              lastMessageTime: new Date().toISOString()
            };
          } else {
            // If this is a new contact, add them to the list
            updatedContacts.push({
              ...selectedContact,
              lastMessage: newMessage,
              lastMessageTime: new Date().toISOString()
            });
          }
        
          return updatedContacts;
        });
      
        // Clear the input field
        setNewMessage('');
      } catch (error) {
        console.error('Error sending message:', error);
      }
    };
  // Show the users list
  const handleShowUsersList = async () => {
    const users = await fetchFollowedUsers();
    console.log('All followed users:', users);
    console.log('Current contacts:', contacts);
    
    // Filter out users we already have conversations with
    const existingContactIds = contacts.map(contact => contact.id);
    console.log('Existing contact IDs:', existingContactIds);
    
    // The issue might be that we need to use followedID instead of id for comparison
    // Let's check both id and followedID to be safe
    const availableUsers = users.filter(user => {
        const userId = user.followedID || user.id; // Use followedID if available, otherwise use id
        const isExistingContact = existingContactIds.includes(userId);
        console.log(`User ${user.firstName} ${user.lastName} (ID: ${userId}) - Existing contact: ${isExistingContact}`);
        return !isExistingContact;
    });
    
    console.log('Available users for new chat:', availableUsers);
    setFollowedUsers(availableUsers);
    setShowUsersList(true);
  };

  // Start a new chat with a user
  const handleStartChat = (user: User) => {
    // Use followedID if available, otherwise use id
    const userId = user.followedID || user.id;
    
    const contact: ChatContact = {
        id: userId,
        firstName: user.firstName,
        lastName: user.lastName,
        nickname: user.nickname,
        avatar: user.avatar
    };
    
    console.log('Starting chat with user:', contact);
    
    setNewMessage('');
    selectedId.current = contact.id;
    setSelectedContact(contact);
    setMessages([]);
    setShowUsersList(false);
    
    // Add this user to contacts if not already there
    if (!contacts.some(c => c.id === userId)) {
        setContacts(prev => [...prev, contact]);
    }
  };

  // Scroll to bottom of messages
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      const formData = new FormData();
      formData.append('image', file);
  
      try {
        const response = await fetch(`${API}/chat/upload-image`, {
          method: 'POST',
          credentials: 'include',
          body: formData,
        });
  
        const data = await response.json();
        if (data.success && data.imageUrl) {
          // Send the image URL as a message with type 'image'
          if (selectedContact && currentUser) {
            if (wsClient && wsClient.socket && wsClient.socket.readyState === WebSocket.OPEN) {
              wsClient.sendMessage(selectedContact.id, data.imageUrl);
            } else {
              // Fallback to HTTP
              await fetch(`${API}/messages`, {
                method: 'POST',
                credentials: 'include',
                headers: {
                  'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                  receiver_id: selectedContact.id,
                  content: data.imageUrl,
                  type: 'image'
                }),
              });
            }
            
            // Add the message to the UI
            const newMessageObj = {
              id: Date.now(),
              sender_id: currentUser.id,
              receiver_id: selectedContact.id,
              content: data.imageUrl,
              created_at: new Date().toISOString(),
              is_read: false,
              type: 'image',
              sender: {
                id : currentUser.id,
                first_name: currentUser.firstName,
                last_name: currentUser.lastName,
                avatar: currentUser.avatar
              }
            };
            
            setMessages(prev => [...prev, newMessageObj]);
            
            // Update the contact's last message
            setContacts(prev => {
              const updatedContacts = [...prev];
              const contactIndex = updatedContacts.findIndex(c => c.id === selectedContact.id);
            
              if (contactIndex >= 0) {
                updatedContacts[contactIndex] = {
                  ...updatedContacts[contactIndex],
                  lastMessage: '📷 Image',
                  lastMessageTime: new Date().toISOString()
                };
              }
            
              return updatedContacts;
            });
          }
        } else {
          console.error('Failed to upload image');
        }
      } catch (error) {
        console.error('Error uploading image:', error);
      }
    }
  };
  

  if (loading) return <div className="chat-page">Loading...</div>;
  if (error) return <div className="chat-page">Error: {error}</div>;

  return (
    <div ref={chatViewportRef} className="chat-page chat-workspace">
      <Sidebar activePage="chat" />
      
      <div className={`chat-container ${selectedContact ? 'has-conversation' : ''}`}>
        <div className="chat-sidebar">
          <div className="chat-sidebar-header">
            <h2>Messages</h2>
            <button className="new-chat-button" onClick={handleShowUsersList}>
              New Chat
            </button>
          </div>
          
          {showUsersList ? (
            <div className="users-list-container">
              <h3>Start a new conversation</h3>
              {followedUsers.length > 0 ? (
                <ul className="users-list">
                  {followedUsers.map(user => (
                    <li key={user.id} className="user-item">
                      <button type="button" className="contact-select" onClick={() => handleStartChat(user)}>
                        <div className="user-avatar">
                          {user.avatar ? (
                            <img src={user.avatar} alt={`${user.firstName}'s avatar`} />
                          ) : (
                            <div className="avatar-placeholder">
                              {user.firstName.charAt(0)}
                            </div>
                          )}
                        </div>
                        <div className="user-info">
                          <span className="user-name">
                            {user.nickname || `${user.firstName} ${user.lastName}`}
                          </span>
                        </div>
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="empty-list-message">You're not following anyone yet.</p>
              )}
            </div>
          ) : (
            <div className="contacts-list-container">
              {contacts.length > 0 ? (
                <ul className="contacts-list">
                  {contacts.map(contact => (
                    <li 
                      key={contact.id} 
                      className={`contact-item ${selectedContact?.id === contact.id ? 'active' : ''}`}
                    >
                      <button type="button" className="contact-select" aria-pressed={selectedContact?.id === contact.id} onClick={() => handleSelectContact(contact)}>
                        <div className="contact-avatar">
                          {contact.avatar ? (
                            <img src={contact.avatar} alt={`${contact.firstName}'s avatar`} />
                          ) : (
                            <div className="avatar-placeholder">
                              {contact.firstName.charAt(0)}
                            </div>
                          )}
                        </div>
                        <div className="contact-info">
                          <span className="contact-name">
                            {contact.nickname || `${contact.firstName} ${contact.lastName}`}
                          </span>
                          {contact.lastMessage && (
                            <p className="last-message">
                              {contact.lastMessage.match(/\.(jpeg|jpg|gif|png)$/i)
                                ? '📷 Image'
                                : contact.lastMessage}
                            </p>
                          )}
                        </div>
                        {(contact.unreadCount ?? 0) > 0 && (
                          <div className="unread-badge">{contact.unreadCount}</div>
                        )}
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="empty-list-message">No conversations yet. Start a new chat!</p>
              )}
            </div>
          )}
        </div>
        
        <div className="chat-main">
          {selectedContact ? (
            <>
              <div className="chat-header">
                <button type="button" className="chat-back-button" aria-label="Back to conversations" onClick={() => setSelectedContact(null)}>
                  <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><path d="m15 18-6-6 6-6" /></svg>
                </button>
                <div className="chat-contact-info">
                  <div className="contact-avatar">
                    {selectedContact.avatar ? (
                      <img src={selectedContact.avatar} alt={`${selectedContact.firstName}'s avatar`} />
                    ) : (
                      <div className="avatar-placeholder">
                        {selectedContact.firstName.charAt(0)}
                      </div>
                    )}
                  </div>
                  <span className="contact-name">
                    {selectedContact.nickname || `${selectedContact.firstName} ${selectedContact.lastName}`}
                  </span>
                  <span className="chat-presence" aria-live="polite">
                    {typingUser === selectedContact.id ? 'Typing…' : onlineUsers.includes(selectedContact.id) ? 'Online' : 'Offline'}
                  </span>
                </div>
              </div>
              
              <div className="messages-container">
                {messages.length > 0 ? (
                  <div className="messages-list">
                   {messages.map(message => (
                     <div
                       key={message.id}
                       className={`message ${message.sender_id === currentUser?.id ? 'sent' : 'received'}`}
                     >
                       <div className="message-content">
                         {typeof message.content === 'string' && message.content.match(/\.(jpeg|jpg|gif|png)$/i) ? (
                           // If the content is an image URL - make it clickable
                           <img
                             src={`${API}${message.content.replace(/\\/g, '/')}`}
                             alt="User uploaded content"
                             onClick={() => handleImageClick(`${API}${message.content.replace(/\\/g, '/')}`)}
                             style={{ cursor: 'pointer' }}
                           />
                         ) : typeof message.content === 'number' ? (
                           // If the content is a number
                           <span>{message.content}</span>
                         ) : typeof message.content === 'string' ? (
                           // If the content is a word or text
                           <span>{message.content}</span>
                         ) : (
                           // Fallback for unsupported content types
                           <span>Unsupported content</span>
                         )}
                       </div>
                       <div className="message-time">
                         {new Date(message.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                         {message.sender_id === currentUser?.id && <span className="message-receipt">{message.is_read ? 'Read' : 'Sent'}</span>}
                       </div>
                     </div>
                   ))}

                    <div ref={messagesEndRef} />
                  </div>
                ) : (
                  <div className="empty-messages">
                    <p>No messages yet. Start the conversation!</p>
                  </div>
                )}
              </div>
              
              <div className="message-input-container">
                <button
                  type="button"
                  className="image-button"
                  aria-label="Attach an image"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                    <circle cx="8.5" cy="8.5" r="1.5"></circle>
                    <polyline points="21 15 16 10 5 21"></polyline>
                  </svg>
                </button>
                <input
                  type="text"
                  className="message-input"
                  aria-label="Message"
                  placeholder="Type a message..."
                  value={newMessage}
                  onChange={(e) => changeMessage(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleSendMessage()}
                  maxLength={100}
                />
                <button 
                  className="send-button"
                  onClick={handleSendMessage}
                  disabled={!newMessage.trim()}
                >
                  Send
                </button>
                <input
                  type="file"
                  accept="image/*"
                  style={{ display: 'none' }}
                  ref={fileInputRef}
                  onChange={handleImageUpload}
                />
              </div>
            </>
          ) : (
            <div className="empty-chat">
              <div className="empty-chat-message">
                <h3>Select a conversation or start a new one</h3>
                <p>Choose a contact from the sidebar or start a new conversation.</p>
                <button className="new-chat-button-large" onClick={handleShowUsersList}>
                  Start a New Chat
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Image Modal */}
      {showImageModal && (
        <div className="image-modal" onClick={closeImageModal}>
          <div className="image-modal-content" onClick={(e) => e.stopPropagation()}>
            <button className="image-modal-close" aria-label="Close image" onClick={closeImageModal}>
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
            <img 
              src={selectedImage} 
              alt="Full size image" 
              className="modal-image"
            />
            <div className="image-modal-info">
              <p>Click outside or press ESC to close</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
