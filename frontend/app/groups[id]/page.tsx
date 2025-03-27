'use client'

import { useEffect, useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import styles from './groupDetail.module.css'
import Sidebar from '@/components/Sidebar'

interface Group {
    id: number
    title: string
    description: string
    creator_id: number
    created_at: string
    member_count: number
    is_creator: boolean
    members: GroupMember[]
    join_requests?: GroupJoinRequest[]
}

interface GroupMember {
    id: number
    user_id: number
    first_name: string
    last_name: string
    avatar?: string
    status: string
    is_creator: boolean
}

interface GroupJoinRequest {
    id: number
    user_id: number
    first_name: string
    last_name: string
    avatar?: string
    created_at: string
}

interface GroupPost {
    id: number
    group_id: number
    user_id: number
    content: string
    image?: string
    created_at: string
    first_name: string
    last_name: string
    avatar?: string
    comment_count: number
}

interface GroupEvent {
    id: number
    group_id: number
    title: string
    description: string
    event_time: string
    created_at: string
    user_response?: string
    response_options: EventResponseOption[]
}

interface EventResponseOption {
    id: number
    event_id: number
    option_text: string
    response_count: number
}

export default function GroupDetail({ params }: { params: { id: string } }) {
    const [group, setGroup] = useState<Group | null>(null)
    const [posts, setPosts] = useState<GroupPost[]>([])
    const [events, setEvents] = useState<GroupEvent[]>([])
    const [activeTab, setActiveTab] = useState('posts')
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState('')

    // State for creating new content
    const [newPost, setNewPost] = useState('')
    const [newPostImage, setNewPostImage] = useState('')
    const [showInviteModal, setShowInviteModal] = useState(false)
    const [showCreateEventModal, setShowCreateEventModal] = useState(false)
    const [availableUsers, setAvailableUsers] = useState<{ id: number, first_name: string, last_name: string }[]>([])
    const [selectedUsers, setSelectedUsers] = useState<number[]>([])

    // State for creating an event
    const [newEvent, setNewEvent] = useState({
        title: '',
        description: '',
        event_time: '',
        options: ['Going', 'Not Going']
    })

    const router = useRouter()
    const groupId = parseInt(params.id)
    const fileInputRef = useRef<HTMLInputElement>(null)

    useEffect(() => {
        const fetchGroupDetails = async () => {
            try {
                // Get group details
                const groupResponse = await fetch(`http://localhost:8080/groups/details?id=${groupId}`, {
                    credentials: 'include'
                })

                if (!groupResponse.ok) {
                    if (groupResponse.status === 404) {
                        router.push('/groups')
                        return
                    }
                    throw new Error('Failed to fetch group details')
                }

                const groupData = await groupResponse.json()
                if (groupData.success) {
                    setGroup(groupData.data)
                }

                // Get group posts
                const postsResponse = await fetch(`http://localhost:8080/groups/posts?group_id=${groupId}`, {
                    credentials: 'include'
                })

                if (postsResponse.ok) {
                    const postsData = await postsResponse.json()
                    if (postsData.success) {
                        setPosts(postsData.posts || [])
                    }
                }

                // Get group events
                const eventsResponse = await fetch(`http://localhost:8080/groups/events?group_id=${groupId}`, {
                    credentials: 'include'
                })

                if (eventsResponse.ok) {
                    const eventsData = await eventsResponse.json()
                    if (eventsData.success) {
                        setEvents(eventsData.events || [])
                    }
                }
            } catch (err) {
                setError('Failed to load group details')
                console.error(err)
            } finally {
                setLoading(false)
            }
        }

        fetchGroupDetails()
    }, [groupId, router])

    const handleCreatePost = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!newPost.trim()) {
            return
        }

        try {
            const response = await fetch('http://localhost:8080/groups/posts/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    group_id: groupId,
                    content: newPost,
                    image: newPostImage
                })
            })

            if (!response.ok) {
                throw new Error('Failed to create post')
            }

            // Refresh posts
            const postsResponse = await fetch(`http://localhost:8080/groups/posts?group_id=${groupId}`, {
                credentials: 'include'
            })

            if (postsResponse.ok) {
                const postsData = await postsResponse.json()
                if (postsData.success) {
                    setPosts(postsData.posts || [])
                }
            }

            // Reset form
            setNewPost('')
            setNewPostImage('')
            if (fileInputRef.current) {
                fileInputRef.current.value = ''
            }
        } catch (err) {
            setError('Failed to create post')
            console.error(err)
        }
    }

    const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0]
        if (!file) return

        const reader = new FileReader()
        reader.onloadend = () => {
            setNewPostImage(reader.result as string)
        }
        reader.readAsDataURL(file)
    }

    const handleOpenInviteModal = async () => {
        try {
            // Fetch users who are not already members
            const response = await fetch('http://localhost:8080/users', {
                credentials: 'include'
            })

            if (!response.ok) {
                throw new Error('Failed to fetch users')
            }

            const data = await response.json()
            if (data.success) {
                // Filter out users who are already members
                const memberIds = group?.members.map(m => m.user_id) || []
                const filteredUsers = data.users.filter((user: any) => !memberIds.includes(user.id))
                setAvailableUsers(filteredUsers)
                setShowInviteModal(true)
            }
        } catch (err) {
            setError('Failed to load users for invitation')
            console.error(err)
        }
    }

    const handleInviteUsers = async () => {
        if (selectedUsers.length === 0) {
            setShowInviteModal(false)
            return
        }

        try {
            const response = await fetch('http://localhost:8080/groups/invite', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    group_id: groupId,
                    user_ids: selectedUsers
                })
            })

            if (!response.ok) {
                throw new Error('Failed to send invitations')
            }

            // Reset and close modal
            setSelectedUsers([])
            setShowInviteModal(false)
        } catch (err) {
            setError('Failed to send invitations')
            console.error(err)
        }
    }

    const handleJoinRequestResponse = async (userId: number, action: string) => {
        try {
            const response = await fetch('http://localhost:8080/groups/membership/handle', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    group_id: groupId,
                    user_id: userId,
                    action: action,
                    request_type: 'request'
                })
            })

            if (!response.ok) {
                throw new Error(`Failed to ${action} request`)
            }

            // Refresh group details to update members and requests
            const groupResponse = await fetch(`http://localhost:8080/groups/details?id=${groupId}`, {
                credentials: 'include'
            })

            if (groupResponse.ok) {
                const groupData = await groupResponse.json()
                if (groupData.success) {
                    setGroup(groupData.data)
                }
            }
        } catch (err) {
            setError(`Failed to ${action} request`)
            console.error(err)
        }
    }

    const handleCreateEvent = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!newEvent.title.trim() || !newEvent.event_time || newEvent.options.length < 2) {
            return
        }

        try {
            const response = await fetch('http://localhost:8080/groups/events/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    group_id: groupId,
                    title: newEvent.title,
                    description: newEvent.description,
                    event_time: new Date(newEvent.event_time).toISOString(),
                    options: newEvent.options
                })
            })

            if (!response.ok) {
                throw new Error('Failed to create event')
            }

            // Refresh events
            const eventsResponse = await fetch(`http://localhost:8080/groups/events?group_id=${groupId}`, {
                credentials: 'include'
            })

            if (eventsResponse.ok) {
                const eventsData = await eventsResponse.json()
                if (eventsData.success) {
                    setEvents(eventsData.events || [])
                }
            }

            // Reset form and close modal
            setNewEvent({
                title: '',
                description: '',
                event_time: '',
                options: ['Going', 'Not Going']
            })
            setShowCreateEventModal(false)
        } catch (err) {
            setError('Failed to create event')
            console.error(err)
        }
    }

    const handleEventResponse = async (eventId: number, optionId: number) => {
        try {
            const response = await fetch('http://localhost:8080/groups/events/respond', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    event_id: eventId,
                    option_id: optionId
                })
            })

            if (!response.ok) {
                throw new Error('Failed to respond to event')
            }

            // Refresh events
            const eventsResponse = await fetch(`http://localhost:8080/groups/events?group_id=${groupId}`, {
                credentials: 'include'
            })

            if (eventsResponse.ok) {
                const eventsData = await eventsResponse.json()
                if (eventsData.success) {
                    setEvents(eventsData.events || [])
                }
            }
        } catch (err) {
            setError('Failed to respond to event')
            console.error(err)
        }
    }

    const formatDate = (dateString: string) => {
        const date = new Date(dateString)
        return date.toLocaleDateString()
    }

    const formatEventDate = (dateString: string) => {
        const date = new Date(dateString)
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }

    if (loading) {
        return (
            <div className={styles.container}>
                <Sidebar activePage="groups" />
                <main className={styles.content}>
                    <div className={styles.loading}>Loading group details...</div>
                </main>
            </div>
        )
    }

    if (!group) {
        return (
            <div className={styles.container}>
                <Sidebar activePage="groups" />
                <main className={styles.content}>
                    <div className={styles.error}>Group not found or you don't have access.</div>
                    <button
                        className={styles.backButton}
                        onClick={() => router.push('/groups')}
                    >
                        Back to Groups
                    </button>
                </main>
            </div>
        )
    }

    return (
        <div className={styles.container}>
            <Sidebar activePage="groups" />

            <main className={styles.content}>
                {error && <div className={styles.error}>{error}</div>}

                <div className={styles.header}>
                    <button
                        className={styles.backButton}
                        onClick={() => router.push('/groups')}
                    >
                        ← Back to Groups
                    </button>
                </div>

                <div className={styles.groupHeader}>
                    <div className={styles.groupInfo}>
                        <h1 className={styles.groupTitle}>{group.title}</h1>
                        {group.description && (
                            <p className={styles.groupDescription}>{group.description}</p>
                        )}
                        <div className={styles.groupMeta}>
                            <span>{group.member_count} members</span>
                            <span>Created: {formatDate(group.created_at)}</span>
                            {group.is_creator && <span className={styles.creatorBadge}>Creator</span>}
                        </div>
                    </div>

                    <div className={styles.groupActions}>
                        {group.is_creator && (
                            <button
                                className={styles.inviteButton}
                                onClick={handleOpenInviteModal}
                            >
                                Invite Members
                            </button>
                        )}
                    </div>
                </div>

                {/* Join Requests Section - Only visible to group creator */}
                {group.is_creator && group.join_requests && group.join_requests.length > 0 && (
                    <div className={styles.joinRequestsSection}>
                        <h2>Join Requests</h2>

                        <div className={styles.joinRequestsList}>
                            {group.join_requests.map(request => (
                                <div key={request.id} className={styles.requestCard}>
                                    <div className={styles.requestUser}>
                                        {request.avatar ? (
                                            <img src={request.avatar} alt={`${request.first_name} ${request.last_name}`} className={styles.userAvatar} />
                                        ) : (
                                            <div className={styles.userInitials}>
                                                {request.first_name[0]}{request.last_name[0]}
                                            </div>
                                        )}
                                        <div>
                                            <h3>{request.first_name} {request.last_name}</h3>
                                            <p className={styles.requestTime}>Requested: {formatDate(request.created_at)}</p>
                                        </div>
                                    </div>

                                    <div className={styles.requestActions}>
                                        <button
                                            className={styles.acceptButton}
                                            onClick={() => handleJoinRequestResponse(request.user_id, 'accept')}
                                        >
                                            Accept
                                        </button>
                                        <button
                                            className={styles.rejectButton}
                                            onClick={() => handleJoinRequestResponse(request.user_id, 'reject')}
                                        >
                                            Decline
                                        </button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {/* Tabs Navigation */}
                <div className={styles.tabsContainer}>
                    <div className={styles.tabs}>
                        <button
                            className={`${styles.tabButton} ${activeTab === 'posts' ? styles.activeTab : ''}`}
                            onClick={() => setActiveTab('posts')}
                        >
                            Posts
                        </button>
                        <button
                            className={`${styles.tabButton} ${activeTab === 'events' ? styles.activeTab : ''}`}
                            onClick={() => setActiveTab('events')}
                        >
                            Events
                        </button>
                        <button
                            className={`${styles.tabButton} ${activeTab === 'members' ? styles.activeTab : ''}`}
                            onClick={() => setActiveTab('members')}
                        >
                            Members
                        </button>
                    </div>
                </div>

                {/* Posts Tab */}
                {activeTab === 'posts' && (
                    <div className={styles.postsSection}>
                        <div className={styles.createPostForm}>
                            <form onSubmit={handleCreatePost}>
                                <textarea
                                    placeholder="Write something to the group..."
                                    value={newPost}
                                    onChange={(e) => setNewPost(e.target.value)}
                                    className={styles.postInput}
                                />

                                {newPostImage && (
                                    <div className={styles.imagePreview}>
                                        <img src={newPostImage} alt="Preview" />
                                        <button
                                            type="button"
                                            className={styles.removeImageButton}
                                            onClick={() => {
                                                setNewPostImage('')
                                                if (fileInputRef.current) fileInputRef.current.value = ''
                                            }}
                                        >
                                            ×
                                        </button>
                                    </div>
                                )}

                                <div className={styles.postActions}>
                                    <div>
                                        <input
                                            type="file"
                                            accept="image/*"
                                            onChange={handleImageUpload}
                                            ref={fileInputRef}
                                            style={{ display: 'none' }}
                                            id="image-upload"
                                        />
                                        <label htmlFor="image-upload" className={styles.uploadButton}>
                                            Add Image
                                        </label>
                                    </div>

                                    <button
                                        type="submit"
                                        className={styles.postButton}
                                        disabled={!newPost.trim()}
                                    >
                                        Post
                                    </button>
                                </div>
                            </form>
                        </div>

                        {posts.length > 0 ? (
                            <div className={styles.postsList}>
                                {posts.map(post => (
                                    <div key={post.id} className={styles.postCard}>
                                        <div className={styles.postHeader}>
                                            <div className={styles.postAuthor}>
                                                {post.avatar ? (
                                                    <img src={post.avatar} alt={`${post.first_name} ${post.last_name}`} className={styles.authorAvatar} />
                                                ) : (
                                                    <div className={styles.authorInitials}>
                                                        {post.first_name[0]}{post.last_name[0]}
                                                    </div>
                                                )}
                                                <div>
                                                    <h3>{post.first_name} {post.last_name}</h3>
                                                    <p className={styles.postTime}>{formatDate(post.created_at)}</p>
                                                </div>
                                            </div>
                                        </div>

                                        <div className={styles.postContent}>
                                            <p>{post.content}</p>
                                            {post.image && (
                                                <img src={post.image} alt="Post attachment" className={styles.postImage} />
                                            )}
                                        </div>

                                        <div className={styles.postFooter}>
                                            <button
                                                className={styles.commentButton}
                                                onClick={() => router.push(`/groups/${groupId}/posts/${post.id}`)}
                                            >
                                                {post.comment_count > 0 ? `${post.comment_count} Comments` : 'Comment'}
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div className={styles.emptyState}>
                                <p>No posts in this group yet. Be the first to post!</p>
                            </div>
                        )}
                    </div>
                )}

                {/* Events Tab */}
                {activeTab === 'events' && (
                    <div className={styles.eventsSection}>
                        <div className={styles.eventsSectionHeader}>
                            <h2>Upcoming Events</h2>
                            <button
                                className={styles.createEventButton}
                                onClick={() => setShowCreateEventModal(true)}
                            >
                                Create Event
                            </button>
                        </div>

                        {events.length > 0 ? (
                            <div className={styles.eventsList}>
                                {events.map(event => (
                                    <div key={event.id} className={styles.eventCard}>
                                        <div className={styles.eventHeader}>
                                            <h3 className={styles.eventTitle}>{event.title}</h3>
                                            <div className={styles.eventTime}>
                                                {formatEventDate(event.event_time)}
                                            </div>
                                        </div>

                                        {event.description && (
                                            <div className={styles.eventDescription}>
                                                <p>{event.description}</p>
                                            </div>
                                        )}

                                        <div className={styles.eventResponseSection}>
                                            {event.user_response ? (
                                                <div className={styles.userResponse}>
                                                    Your response: <span className={styles.responseHighlight}>{event.user_response}</span>
                                                </div>
                                            ) : (
                                                <div className={styles.responseOptions}>
                                                    {event.response_options.map(option => (
                                                        <button
                                                            key={option.id}
                                                            className={styles.responseOption}
                                                            onClick={() => handleEventResponse(event.id, option.id)}
                                                        >
                                                            {option.option_text}
                                                        </button>
                                                    ))}
                                                </div>
                                            )}
                                        </div>

                                        <div className={styles.eventFooter}>
                                            <button
                                                className={styles.viewEventButton}
                                                onClick={() => router.push(`/groups/${groupId}/events/${event.id}`)}
                                            >
                                                View Details
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div className={styles.emptyState}>
                                <p>No events scheduled yet. Create one to get started!</p>
                            </div>
                        )}
                    </div>
                )}

                {/* Members Tab */}
                {activeTab === 'members' && (
                    <div className={styles.membersSection}>
                        <h2>Group Members ({group.members.length})</h2>

                        <div className={styles.membersList}>
                            {group.members.map(member => (
                                <div key={member.id} className={styles.memberCard}>
                                    <div className={styles.memberInfo}>
                                        {member.avatar ? (
                                            <img src={member.avatar} alt={`${member.first_name} ${member.last_name}`} className={styles.memberAvatar} />
                                        ) : (
                                            <div className={styles.memberInitials}>
                                                {member.first_name[0]}{member.last_name[0]}
                                            </div>
                                        )}
                                        <div>
                                            <h3>{member.first_name} {member.last_name}</h3>
                                            {member.is_creator && <span className={styles.creatorTag}>Creator</span>}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </main>

            {/* Invite Members Modal */}
            {showInviteModal && (
                <div className={styles.modalOverlay}>
                    <div className={styles.modal}>
                        <h2>Invite Members</h2>

                        {availableUsers.length > 0 ? (
                            <>
                                <div className={styles.usersList}>
                                    {availableUsers.map(user => (
                                        <div key={user.id} className={styles.userItem}>
                                            <label className={styles.checkboxLabel}>
                                                <input
                                                    type="checkbox"
                                                    checked={selectedUsers.includes(user.id)}
                                                    onChange={() => {
                                                        if (selectedUsers.includes(user.id)) {
                                                            setSelectedUsers(selectedUsers.filter(id => id !== user.id))
                                                        } else {
                                                            setSelectedUsers([...selectedUsers, user.id])
                                                        }
                                                    }}
                                                />
                                                {user.first_name} {user.last_name}
                                            </label>
                                        </div>
                                    ))}
                                </div>

                                <div className={styles.modalActions}>
                                    <button
                                        className={styles.cancelButton}
                                        onClick={() => {
                                            setSelectedUsers([])
                                            setShowInviteModal(false)
                                        }}
                                    >
                                        Cancel
                                    </button>
                                    <button
                                        className={styles.inviteButton}
                                        onClick={handleInviteUsers}
                                        disabled={selectedUsers.length === 0}
                                    >
                                        Send Invitations
                                    </button>
                                </div>
                            </>
                        ) : (
                            <div className={styles.emptyState}>
                                <p>No users available to invite.</p>
                                <button
                                    className={styles.cancelButton}
                                    onClick={() => setShowInviteModal(false)}
                                >
                                    Close
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            )}

            {/* Create Event Modal */}
            {showCreateEventModal && (
                <div className={styles.modalOverlay}>
                    <div className={styles.modal}>
                        <h2>Create New Event</h2>

                        <form onSubmit={handleCreateEvent}>
                            <div className={styles.formGroup}>
                                <label htmlFor="eventTitle">Event Title *</label>
                                <input
                                    type="text"
                                    id="eventTitle"
                                    value={newEvent.title}
                                    onChange={(e) => setNewEvent({ ...newEvent, title: e.target.value })}
                                    required
                                />
                            </div>

                            <div className={styles.formGroup}>
                                <label htmlFor="eventDescription">Description</label>
                                <textarea
                                    id="eventDescription"
                                    value={newEvent.description}
                                    onChange={(e) => setNewEvent({ ...newEvent, description: e.target.value })}
                                    rows={3}
                                />
                            </div>

                            <div className={styles.formGroup}>
                                <label htmlFor="eventTime">Date and Time *</label>
                                <input
                                    type="datetime-local"
                                    id="eventTime"
                                    value={newEvent.event_time}
                                    onChange={(e) => setNewEvent({ ...newEvent, event_time: e.target.value })}
                                    required
                                />
                            </div>

                            <div className={styles.formGroup}>
                                <label>Response Options *</label>
                                <p className={styles.optionsHelp}>Add at least two options (e.g., "Going", "Not Going")</p>

                                {newEvent.options.map((option, index) => (
                                    <div key={index} className={styles.optionItem}>
                                        <input
                                            type="text"
                                            value={option}
                                            onChange={(e) => {
                                                const updatedOptions = [...newEvent.options]
                                                updatedOptions[index] = e.target.value
                                                setNewEvent({ ...newEvent, options: updatedOptions })
                                            }}
                                            placeholder={`Option ${index + 1}`}
                                            required
                                        />
                                        {newEvent.options.length > 2 && (
                                            <button
                                                type="button"
                                                className={styles.removeOptionButton}
                                                onClick={() => {
                                                    const updatedOptions = newEvent.options.filter((_, i) => i !== index)
                                                    setNewEvent({ ...newEvent, options: updatedOptions })
                                                }}
                                            >
                                                ×
                                            </button>
                                        )}
                                    </div>
                                ))}

                                <button
                                    type="button"
                                    className={styles.addOptionButton}
                                    onClick={() => setNewEvent({
                                        ...newEvent,
                                        options: [...newEvent.options, '']
                                    })}
                                >
                                    + Add Option
                                </button>
                            </div>

                            <div className={styles.modalActions}>
                                <button
                                    type="button"
                                    className={styles.cancelButton}
                                    onClick={() => setShowCreateEventModal(false)}
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    className={styles.createButton}
                                    disabled={!newEvent.title.trim() || !newEvent.event_time || newEvent.options.some(opt => !opt.trim())}
                                >
                                    Create Event
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    )
}


