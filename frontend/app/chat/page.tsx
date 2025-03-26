'use client'

import { useRouter } from 'next/navigation'
import { useState, useEffect, useRef } from 'react'
import Sidebar from '../../components/Sidebar'
import './chat.css'
import { WebSocketClient } from '../webscoket/websocket'

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

  // Get or create WebSocket connection
  useEffect(() => {
    // Get the singleton instance of WebSocketClient
    const client = WebSocketClient.getInstance();
    
    // Only set up message handlers if we have user data
    if (currentUser) {
        // Add message handler for private messages
        client.addMessageHandler('private_message', (content) => {
            // Your existing message handling code
        });
        
        // Don't call connect() here - it's already connected from Home page
        // or will be connected by getInstance if needed
        
        setWsClient(client);
    }
    
    // No need to clean up the connection when leaving the chat page
    // as we want to keep it alive for the entire session
  }, [currentUser, selectedContact]);

  // Fetch current user data
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

  // Fetch chat contacts (users with message history)
  const fetchChatContacts = async () => {
    try {
      // This endpoint doesn't exist yet, but we're assuming it will be implemented
      const response = await fetch('http://localhost:8080/chat/contacts', {
        method: 'GET',
        credentials: 'include'
      });

      if (!response.ok) {
        // If the endpoint doesn't exist yet, we'll just use an empty array
        console.warn('Chat contacts endpoint not implemented yet');
        return [];
      }

      const data = await response.json();
      if (data.success && data.contacts) {
        return data.contacts;
      }
      return [];
    } catch (error) {
      console.error('Error fetching chat contacts:', error);
      return [];
    }
  };

  // Fetch messages for a selected contact
  const fetchMessages = async (contactId: number) => {
    try {
      const response = await fetch(`http://localhost:8080/messages/${contactId}`, {
        method: 'GET',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to fetch messages');
      }

      const data = await response.json();
      if (data.success && data.messages) {
        return data.messages;
      }
      return [];
    } catch (error) {
      console.error('Error fetching messages:', error);
      return [];
    }
  };

  // Fetch followed users
  const fetchFollowedUsers = async () => {
    try {
      // We'll use the profile endpoint to get followed users
      const response = await fetch('http://localhost:8080/profile', {
        method: 'GET',
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error('Failed to fetch followed users');
      }

      const data = await response.json();
      if (data.following) {
        return data.following;
      }
      return [];
    } catch (error) {
      console.error('Error fetching followed users:', error);
      return [];
    }
  };

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
      const response = await fetch('http://localhost:8080/messages', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          receiver_id: selectedContact.id,
          content: newMessage
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to send message');
      }

      const data = await response.json();
      if (data.success && data.message) {
        // Add the new message to the messages list
        setMessages(prev => [...prev, data.message]);
        
        // Update the contact's last message
        setContacts(prev => {
          const updatedContacts = [...prev];
          const contactIndex = updatedContacts.findIndex(c => c.id === selectedContact.id);
          
          if (contactIndex >= 0) {
            updatedContacts[contactIndex] = {
              ...updatedContacts[contactIndex],
              lastMessage: newMessage,
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
      }
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
                          <p className="last-message">{contact.lastMessage}</p>
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
                        <div className="message-content">{message.content}</div>
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
}
