'use client'

import { useRouter } from 'next/navigation'
import { useState, useEffect, useRef } from 'react'
import Sidebar from '../../components/Sidebar'
import './chat.css'
import { WebSocketClient } from '../webscoket/websocket'
import { fetchFollowedUsers, fetchChatContacts, fetchMessages, sendMessage } from './messageHandlers'
interface User {
  id: number
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
  const router = useRouter()
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

      useEffect(() => {
        const client = WebSocketClient.getInstance();
    
        if (currentUser) {
          client.addMessageHandler('private_message', (content) => {
            console.log('Received private message:', content);
        
            // Only process messages if they're from the currently selected contact
            if (selectedContact && content.sender_id === selectedContact.id) {
              const newMessage = {
                id: content.id || 0,
                sender_id: content.sender_id || 0,
                receiver_id: currentUser.id,
                content: content.content || "",
                type: content.type || "text",
                created_at: content.created_at || new Date().toISOString(),
                is_read: false,
                sender: {
                  id: content.sender_id || 0,
                  first_name: typeof content.sender === 'string' ? content.sender.split(' ')[0] : "",
                  last_name: typeof content.sender === 'string' ? (content.sender.split(' ')[1] || "") : "",
                  avatar: undefined
                }
              };
          
              setMessages(prev => [...prev, newMessage]);
            }
        
            // Update the contact's last message in the contacts list
            setContacts(prev => {
              const updatedContacts = [...prev];
              const contactIndex = updatedContacts.findIndex(c => c.id === content.sender_id);
          
              if (contactIndex >= 0) {
                updatedContacts[contactIndex] = {
                  ...updatedContacts[contactIndex],
                  lastMessage: content.type === 'image' || content.content.match(/\.(jpeg|jpg|gif|png)$/i) 
                    ? '📷 Image' 
                    : content.content || "",
                  lastMessageTime: content.created_at || new Date().toISOString(),
                  unreadCount: (updatedContacts[contactIndex].unreadCount || 0) + 1
                };
              } else if (content.sender_id) {
                // If this is a new contact, we need to fetch their info and add them
                fetch(`http://localhost:8080/users?id=${content.sender_id}`, {
                  credentials: 'include'
                })
                .then(res => res.json())
                .then(data => {
                  if (data.success && data.user) {
                    const newContact = {
                      id: content.sender_id,
                      firstName: data.user.firstName || data.user.first_name,
                      lastName: data.user.lastName || data.user.last_name,
                      nickname: data.user.nickname,
                      avatar: data.user.avatar,
                      lastMessage: content.type === 'image' ? '📷 Image' : content.content || "",
                      lastMessageTime: content.created_at || new Date().toISOString(),
                      unreadCount: 1
                    };
                    setContacts(prev => [...prev, newContact]);
                  }
                })
                .catch(err => console.error('Error fetching new contact info:', err));
              }
          
              return updatedContacts;
            });
          });
      
          setWsClient(client);
        }
    
        // No cleanup needed as we want to keep the connection alive
      }, [currentUser, selectedContact]);

  useEffect(() => {
    const fetchUserData = async () => {
      try {
        const response = await fetch('http://localhost:8080/profile', {
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
    setSelectedContact(contact);
    const contactMessages = await fetchMessages(contact.id);
    setMessages(contactMessages);
    setShowUsersList(false);
    
    // Mark messages as read (this would be implemented in the backend)
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
          const response = await fetch('http://localhost:8080/messages', {
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
    console.log(users);
    setFollowedUsers(users);
    setShowUsersList(true);
  };

  // Start a new chat with a user
  const handleStartChat = (user: User) => {
    const contact: ChatContact = {
      id: user.id,
      firstName: user.firstName,
      lastName: user.lastName,
      nickname: user.nickname,
      avatar: user.avatar
    };
    
    setSelectedContact(contact);
    setMessages([]);
    setShowUsersList(false);
    
    // Add this user to contacts if not already there
    if (!contacts.some(c => c.id === user.id)) {
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
        const response = await fetch('http://localhost:8080/chat/upload-image', {
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
              await fetch('http://localhost:8080/messages', {
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
    <div className="chat-page">
      <Sidebar activePage="chat" />
      
      <div className="chat-container">
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
                    <li key={user.id} className="user-item" onClick={() => handleStartChat(user)}>
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
                      onClick={() => handleSelectContact(contact)}
                    >
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
                      {contact.unreadCount && contact.unreadCount > 0 && (
                        <div className="unread-badge">{contact.unreadCount}</div>
                      )}
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
        // If the content is an image URL
        <img src={`http://localhost:8080/${message.content}`} alt="User uploaded content" />
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
                  placeholder="Type a message..."
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleSendMessage()}
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
    </div>
  );
};
