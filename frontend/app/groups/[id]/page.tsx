'use client'

import { useState, useEffect, useRef } from 'react'
import Sidebar from '@/components/Sidebar'
import { useRouter, useParams } from 'next/navigation'
import './groupChat.css'

interface Group {
    id: number
    title: string
    description: string
    creator_id: number
    created_at: string
    member_count: number
    is_creator: boolean
}

interface GroupMember {
    id: number
    user_id: number
    first_name: string
    last_name: string
    avatar?: string
    role: string
}

interface GroupMessage {
    id: number
    sender_id: number
    content: string
    created_at: string
    sender: {
        id: number
        firstName: string
        lastName: string
        avatar?: string
    }
}

interface GroupPost {
    id: number
    user_id: number
    content: string
    image?: string
    created_at: string
    firstName: string
    lastName: string
    avatar?: string
    comment_count: number
}

interface GroupEvent {
    id: number
    title: string
    description: string
    event_time: string
    created_at: string
}

interface User {
    id: number
    firstName: string
    lastName: string
    avatar?: string
}

export default function GroupChatPage() {
    const router = useRouter()
    const params = useParams()
    const groupId = params?.id ? parseInt(params.id as string) : 0

    const [loading, setLoading] = useState(true)
    const [group, setGroup] = useState<Group | null>(null)
    const [members, setMembers] = useState<GroupMember[]>([])
    const [messages, setMessages] = useState<GroupMessage[]>([])
    const [posts, setPosts] = useState<GroupPost[]>([])
    const [events, setEvents] = useState<GroupEvent[]>([])
    const [newMessage, setNewMessage] = useState('')
    const [error, setError] = useState('')
    const [activeTab, setActiveTab] = useState('chat') // 'chat', 'posts', or 'events'

    // Invite users functionality
    const [showInviteModal, setShowInviteModal] = useState(false)
    const [users, setUsers] = useState<User[]>([])
    const [selectedUsers, setSelectedUsers] = useState<number[]>([])
    const [inviteLoading, setInviteLoading] = useState(false)
    const [inviteSuccess, setInviteSuccess] = useState('')
    const [inviteError, setInviteError] = useState('')

    useEffect(() => {
        const fetchGroupDetails = async () => {
            try {
                // Fetch group details
                const response = await fetch(`http://localhost:8080/groups/details?id=${groupId}`, {
                    method: 'GET',
                    credentials: 'include'
                })

                if (response.ok) {
                    const data = await response.json()
                    if (data.success && data.data) {
                        setGroup(data.data.group)
                        setMembers(data.data.members || [])
                    } else {
                        setError('Failed to load group details')
                    }
                } else {
                    setError('Failed to load group')
                    router.push('/groups')
                }
            } catch (error) {
                console.error('Error fetching group details:', error)
                setError('An error occurred while loading the group')
            } finally {
                setLoading(false)
            }
        }

        if (groupId) {
            fetchGroupDetails()
        }
    }, [groupId, router])

    // Fetch all users for invitation
    const fetchUsers = async () => {
        try {
            const response = await fetch('http://localhost:8080/users', {
                method: 'GET',
                credentials: 'include'
            })

            if (response.ok) {
                const data = await response.json()
                if (data.users) {
                    // Filter out users who are already members
                    const memberIds = members.map(member => member.user_id)
                    const filteredUsers = data.users.filter((user: User) =>
                        !memberIds.includes(user.id)
                    )
                    setUsers(filteredUsers)
                }
            } else {
                console.error('Failed to fetch users')
            }
        } catch (error) {
            console.error('Error fetching users:', error)
        }
    }

    // Handle inviting users
    const handleInviteUsers = async () => {
        if (selectedUsers.length === 0) {
            setInviteError('Please select at least one user to invite')
            return
        }

        setInviteLoading(true)
        setInviteError('')
        setInviteSuccess('')

        try {
            const response = await fetch('http://localhost:8080/groups/invite', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    group_id: groupId,
                    user_ids: selectedUsers
                }),
            })

            if (response.ok) {
                setInviteSuccess('Invitations sent successfully!')
                setSelectedUsers([])
                setTimeout(() => {
                    setShowInviteModal(false)
                }, 1500)
            } else {
                const errorData = await response.json()
                setInviteError(errorData.error || 'Failed to send invitations')
            }
        } catch (error) {
            console.error('Error sending invitations:', error)
            setInviteError('An error occurred while sending invitations')
        } finally {
            setInviteLoading(false)
        }
    }

    // Toggle user selection
    const toggleUserSelection = (userId: number) => {
        setSelectedUsers(prev =>
            prev.includes(userId)
                ? prev.filter(id => id !== userId)
                : [...prev, userId]
        )
    }

    const handleSendMessage = (e: React.FormEvent) => {
        e.preventDefault()
        console.log('Message would be sent:', newMessage)
        setNewMessage('')
    }

    if (loading) {
        return (
            <div className="groups-page">
                <Sidebar activePage="groups" />
                <div className="group-chat-container">
                    <div className="loading-message">Loading group...</div>
                </div>
            </div>
        )
    }

    if (error) {
        return (
            <div className="groups-page">
                <Sidebar activePage="groups" />
                <div className="group-chat-container">
                    <div className="error-message">{error}</div>
                </div>
            </div>
        )
    }

    return (
        <div className="groups-page">
            <Sidebar activePage="groups" />

            <div className="group-chat-container">
                <div className="group-chat-sidebar">
                    <div className="group-chat-sidebar-header">
                        <h2>{group?.title}</h2>
                        <button
                            className="back-button"
                            onClick={() => router.push('/groups')}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <path d="M19 12H5M12 19l-7-7 7-7" />
                            </svg>
                            Back
                        </button>
                    </div>

                    <div className="group-info">
                        <p className="group-description">{group?.description}</p>
                        <div className="group-meta">
                            <span className="member-count">{group?.member_count || 0} members</span>
                            <span className="created-date">Created {new Date(group?.created_at || '').toLocaleDateString()}</span>
                        </div>
                    </div>

                    <div className="group-tabs">
                        <button
                            className={`group-tab ${activeTab === 'chat' ? 'active' : ''}`}
                            onClick={() => setActiveTab('chat')}
                        >
                            Chat
                        </button>
                        <button
                            className={`group-tab ${activeTab === 'posts' ? 'active' : ''}`}
                            onClick={() => setActiveTab('posts')}
                        >
                            Posts
                        </button>
                        <button
                            className={`group-tab ${activeTab === 'events' ? 'active' : ''}`}
                            onClick={() => setActiveTab('events')}
                        >
                            Events
                        </button>
                    </div>

                    <div className="group-members-container">
                        <h3>Members</h3>
                        {members.length > 0 ? (
                            <ul className="group-members-list">
                                {members.map(member => (
                                    <li key={member.id} className="member-item">
                                        <div className="member-avatar">
                                            {member.avatar ? (
                                                <img src={member.avatar} alt={`${member.first_name}'s avatar`} />
                                            ) : (
                                                <div className="avatar-placeholder">
                                                    {member.first_name.charAt(0)}
                                                </div>
                                            )}
                                        </div>
                                        <div className="member-info">
                                            <span className="member-name">{member.first_name} {member.last_name}</span>
                                            <span className="member-role">{member.role}</span>
                                        </div>
                                    </li>
                                ))}
                            </ul>
                        ) : (
                            <p className="empty-list-message">No members found</p>
                        )}
                    </div>
                </div>

                <div className="group-chat-main">
                    <div className="group-chat-header">
                        <div className="group-chat-info">
                            <h2>{group?.title}</h2>
                            <span className="member-count">{group?.member_count || 0} members</span>
                        </div>
                        <div className="group-actions">
                            <button
                                className="invite-button"
                                onClick={() => {
                                    fetchUsers()
                                    setShowInviteModal(true)
                                }}
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                    <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                                    <circle cx="8.5" cy="7" r="4"></circle>
                                    <line x1="20" y1="8" x2="20" y2="14"></line>
                                    <line x1="23" y1="11" x2="17" y2="11"></line>
                                </svg>
                                Invite Users
                            </button>
                            <button className="group-action-button">
                                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                    <circle cx="12" cy="12" r="1"></circle>
                                    <circle cx="19" cy="12" r="1"></circle>
                                    <circle cx="5" cy="12" r="1"></circle>
                                </svg>
                            </button>
                        </div>
                    </div>

                    {activeTab === 'chat' && (
                        <>
                            <div className="group-chat-messages">
                                {messages.length > 0 ? (
                                    <div className="messages-list">
                                        {messages.map(message => (
                                            <div
                                                key={message.id}
                                                className={`message ${message.sender_id === 1 ? 'sent' : 'received'}`}
                                            >
                                                {message.sender_id !== 1 && (
                                                    <div className="message-sender">
                                                        <div className="sender-avatar">
                                                            {message.sender.avatar ? (
                                                                <img src={message.sender.avatar} alt={`${message.sender.firstName}'s avatar`} />
                                                            ) : (
                                                                <div className="avatar-placeholder">
                                                                    {message.sender.firstName.charAt(0)}
                                                                </div>
                                                            )}
                                                        </div>
                                                        <span className="sender-name">{message.sender.firstName} {message.sender.lastName}</span>
                                                    </div>
                                                )}
                                                <div className="message-content">{message.content}</div>
                                                <div className="message-time">
                                                    {new Date(message.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="empty-messages">
                                        <p>No messages in this group yet. Be the first to say hello!</p>
                                    </div>
                                )}
                            </div>

                            <form className="group-message-input-container" onSubmit={handleSendMessage}>
                                <input
                                    type="text"
                                    className="message-input"
                                    placeholder="Type a message..."
                                    value={newMessage}
                                    onChange={(e) => setNewMessage(e.target.value)}
                                />
                                <button
                                    type="submit"
                                    className="send-button"
                                    disabled={!newMessage.trim()}
                                >
                                    Send
                                </button>
                            </form>
                        </>
                    )}

                    {activeTab === 'posts' && (
                        <div className="group-posts">
                            <div className="create-post">
                                <textarea
                                    className="post-input"
                                    placeholder="Write a post..."
                                    rows={3}
                                ></textarea>
                                <button className="post-button">Post</button>
                            </div>

                            {posts.length > 0 ? (
                                <div className="posts-list">
                                    {posts.map(post => (
                                        <div key={post.id} className="post-item">
                                            <div className="post-header">
                                                <div className="post-author">
                                                    <div className="author-avatar">
                                                        {post.avatar ? (
                                                            <img src={post.avatar} alt={`${post.firstName}'s avatar`} />
                                                        ) : (
                                                            <div className="avatar-placeholder">
                                                                {post.firstName.charAt(0)}
                                                            </div>
                                                        )}
                                                    </div>
                                                    <div className="author-info">
                                                        <span className="author-name">{post.firstName} {post.lastName}</span>
                                                        <span className="post-time">{new Date(post.created_at).toLocaleString()}</span>
                                                    </div>
                                                </div>
                                            </div>
                                            <div className="post-content">{post.content}</div>
                                            {post.image && (
                                                <div className="post-image">
                                                    <img src={post.image} alt="Post attachment" />
                                                </div>
                                            )}
                                            <div className="post-footer">
                                                <button className="comment-button">
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                                                    </svg>
                                                    {post.comment_count} Comments
                                                </button>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <div className="empty-posts">
                                    <p>No posts in this group yet. Create the first post!</p>
                                </div>
                            )}
                        </div>
                    )}

                    {activeTab === 'events' && (
                        <div className="group-events">
                            <div className="create-event">
                                <button className="create-event-button">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                                        <line x1="16" y1="2" x2="16" y2="6"></line>
                                        <line x1="8" y1="2" x2="8" y2="6"></line>
                                        <line x1="3" y1="10" x2="21" y2="10"></line>
                                        <line x1="12" y1="14" x2="12" y2="18"></line>
                                        <line x1="10" y1="16" x2="14" y2="16"></line>
                                    </svg>
                                    Create Event
                                </button>
                            </div>

                            {events.length > 0 ? (
                                <div className="events-list">
                                    {events.map(event => (
                                        <div key={event.id} className="event-item">
                                            <div className="event-date">
                                                <div className="event-month">
                                                    {new Date(event.event_time).toLocaleString('default', { month: 'short' })}
                                                </div>
                                                <div className="event-day">
                                                    {new Date(event.event_time).getDate()}
                                                </div>
                                            </div>
                                            <div className="event-details">
                                                <h3 className="event-title">{event.title}</h3>
                                                <p className="event-description">{event.description}</p>
                                                <div className="event-time">
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                        <circle cx="12" cy="12" r="10"></circle>
                                                        <polyline points="12 6 12 12 16 14"></polyline>
                                                    </svg>
                                                    {new Date(event.event_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                                </div>
                                            </div>
                                            <div className="event-actions">
                                                <button className="event-action-button">
                                                    Going
                                                </button>
                                                <button className="event-action-button">
                                                    Maybe
                                                </button>
                                                <button className="event-action-button">
                                                    Not Going
                                                </button>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <div className="empty-events">
                                    <p>No events scheduled in this group yet.</p>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            {/* Invite Users Modal */}
            {showInviteModal && (
                <div className="modal-overlay">
                    <div className="invite-modal">
                        <div className="modal-header">
                            <h3>Invite Users to {group?.title}</h3>
                            <button
                                className="close-button"
                                onClick={() => setShowInviteModal(false)}
                            >
                                &times;
                            </button>
                        </div>

                        {inviteError && <div className="error-message">{inviteError}</div>}
                        {inviteSuccess && <div className="success-message">{inviteSuccess}</div>}

                        <div className="users-list-container">
                            {users.length > 0 ? (
                                <ul className="users-list">
                                    {users.map(user => (
                                        <li
                                            key={user.id}
                                            className={`user-item ${selectedUsers.includes(user.id) ? 'selected' : ''}`}
                                            onClick={() => toggleUserSelection(user.id)}
                                        >
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
                                                <span className="user-name">{user.firstName} {user.lastName}</span>
                                            </div>
                                            <div className="checkbox">
                                                {selectedUsers.includes(user.id) && (
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                        <polyline points="20 6 9 17 4 12"></polyline>
                                                    </svg>
                                                )}
                                            </div>
                                        </li>
                                    ))}
                                </ul>
                            ) : (
                                <p className="empty-list-message">No users available to invite</p>
                            )}
                        </div>

                        <div className="modal-footer">
                            <button
                                className="cancel-button"
                                onClick={() => setShowInviteModal(false)}
                            >
                                Cancel
                            </button>
                            <button
                                className="invite-button"
                                onClick={handleInviteUsers}
                                disabled={selectedUsers.length === 0 || inviteLoading}
                            >
                                {inviteLoading ? 'Sending...' : 'Send Invitations'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}

