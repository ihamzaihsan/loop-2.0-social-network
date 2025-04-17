'use client'

import { useState, useEffect, useRef } from 'react'
import Sidebar from '@/components/Sidebar'
import { useRouter, useParams } from 'next/navigation'
import './groupChat.css'
import { WebSocketClient } from '../../webscoket/websocket'
import {
    fetchGroupMessages,
    sendGroupMessage,
    createGroupPost,
    createGroupComment,
    createGroupEvent,
    respondToEvent,
    fetchGroupPostComments
} from '../../chat/messageHandlers'

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
    id: number;
    sender_id: number;
    group_id: number;
    content: string;
    created_at: string;
    sender: {
        id: number;
        firstName: string;
        lastName: string;
        avatar?: string;
    };
}

interface GroupPost {
    id: number
    user_id: number
    content: string
    image?: string
    created_at: string
    first_name: string
    last_name: string
    avatar?: string
    comment_count: number
}
interface EventResponseOption {
    id: number;
    event_id: number;
    option_text: string;
    response_count: number;
}

interface GroupEvent {
    id: number
    title: string
    description: string
    event_time: string
    created_at: string
    going_count?: number
    not_going_count?: number
    user_response?: string
    ResponseOptions?: EventResponseOption[]
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

    const [wsClient, setWsClient] = useState<WebSocketClient | null>(null)
    const [currentUser, setCurrentUser] = useState<User | null>(null)
    const messagesEndRef = useRef<HTMLDivElement>(null)

    // For posts functionality
    const [newPostContent, setNewPostContent] = useState('')
    const [newPostImage, setNewPostImage] = useState<string | null>(null)
    const [selectedPost, setSelectedPost] = useState<number | null>(null)
    const [newComment, setNewComment] = useState('')

    // For events functionality
    const [showEventForm, setShowEventForm] = useState(false)
    const [eventTitle, setEventTitle] = useState('')
    const [eventDescription, setEventDescription] = useState('')
    const [eventDate, setEventDate] = useState('')
    const [eventTime, setEventTime] = useState('')
    const [eventError, setEventError] = useState<string>('');
    // Add this to your state variables
    const [userEventResponses, setUserEventResponses] = useState<{ [eventId: number]: number }>({});

    // Invite users functionality
    const [showInviteModal, setShowInviteModal] = useState(false)
    const [users, setUsers] = useState<User[]>([])
    const [selectedUsers, setSelectedUsers] = useState<number[]>([])
    const [inviteLoading, setInviteLoading] = useState(false)
    const [inviteSuccess, setInviteSuccess] = useState('')
    const [inviteError, setInviteError] = useState('')

    // Add these to your existing state variables
const fileInputRef = useRef<HTMLInputElement>(null);

// Add this function to handle image uploads
const handleGroupImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
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
          // Send the image URL as a message
          if (groupId && currentUser) {
            if (wsClient && wsClient.socket && wsClient.socket.readyState === WebSocket.OPEN) {
              wsClient.sendGroupMessage(groupId, data.imageUrl);
            } else {
              // Fallback to HTTP
              await sendGroupMessage(groupId, data.imageUrl);
            }
            
            // Add the message to the UI
            const newMessageObj: GroupMessage = {
              id: Date.now(),
              sender_id: currentUser.id,
              group_id: groupId,
              content: data.imageUrl,
              created_at: new Date().toISOString(),
              sender: {
                id: currentUser.id,
                firstName: currentUser.firstName,
                lastName: currentUser.lastName,
                avatar: currentUser.avatar
              }
            };
            
            setMessages(prev => [...prev, newMessageObj]);
          }
        } else {
          console.error('Failed to upload image');
        }
      } catch (error) {
        console.error('Error uploading image:', error);
      }
    }
  };
  
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
                        avatar: data.user.avatar
                    });
                }
            } catch (error: any) {
                console.error('Error fetching user data:', error);
            }
        };

        fetchUserData();
    }, [router]);

    // Fetch group details
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

    // Fetch group messages, posts, and events
    useEffect(() => {
        const fetchGroupData = async () => {
            if (!groupId) return;

            try {
                // Fetch group messages
                const messagesResponse = await fetch(`http://localhost:8080/groups/messages?id=${groupId}`, {
                    method: 'GET',
                    credentials: 'include'
                });

                if (messagesResponse.ok) {
                    const messagesData = await messagesResponse.json();
                    if (messagesData.success && messagesData.messages) {
                        setMessages(messagesData.messages);
                    }
                }

                // Fetch group posts
                const postsResponse = await fetch(`http://localhost:8080/groups/posts?group_id=${groupId}`, {
                    method: 'GET',
                    credentials: 'include'
                });

                if (postsResponse.ok) {
                    const postsData = await postsResponse.json();
                    if (postsData.success && postsData.posts) {
                        setPosts(postsData.posts);
                    }
                }

                // Fetch group events
                const eventsResponse = await fetch(`http://localhost:8080/groups/events?group_id=${groupId}`, {
                    method: 'GET',
                    credentials: 'include'
                });

                if (eventsResponse.ok) {
                    const eventsData = await eventsResponse.json();
                    if (eventsData.success && eventsData.events) {
                        setEvents(eventsData.events);
                    }
                }
            } catch (error) {
                console.error('Error fetching group data:', error);
            }
        };

        if (groupId) {
            fetchGroupData();
        }
    }, [groupId]);

    // Setup WebSocket connection
    useEffect(() => {
        const client = WebSocketClient.getInstance();

        if (client && currentUser) {
            // Add message handler for group messages
            client.addMessageHandler('group_message', (content) => {
                console.log('Received group message:', content);

                // Only process messages for the current group
                if (content.group_id === groupId) {
                    const newMessage: GroupMessage = {
                        id: content.id || 0,
                        sender_id: content.sender_id || 0,
                        group_id: content.group_id || 0,
                        content: content.content || "",
                        created_at: content.created_at || new Date().toISOString(),
                        sender: {
                            id: content.sender_id || 0,
                            firstName: content.sender?.firstName || "",
                            lastName: content.sender?.lastName || "",
                            avatar: content.sender?.avatar
                        }
                    };

                    // Don't add messages from the current user (they're added directly when sent)
                    if (newMessage.sender_id !== currentUser.id) {
                        setMessages(prev => [...prev, newMessage]);
                    }
                }
            });

            // Add message handler for group posts
            client.addMessageHandler('group_post', (content) => {
                if (content.group_id === groupId && content.user_id !== currentUser.id) {
                    setPosts(prev => [content, ...prev]);
                }
            });

            client.addMessageHandler('group_comment', (content) => {
                console.log('Received group comment:', content);
                
                // Update the post's comment count
                setPosts(prev =>
                    prev.map(post =>
                        post.id === content.post_id
                            ? { ...post, comment_count: (post.comment_count || 0) + 1 }
                            : post
                    )
                );
                
                // Add the comment to the comments list if we're viewing that post
                if (content.post_id) {
                    setPostComments(prev => {
                        // If we already have comments for this post, add the new one
                        if (prev[content.post_id]) {
                            // Check if this comment is already in the list to avoid duplicates
                            const isDuplicate = prev[content.post_id].some(comment => 
                                comment.id === content.id
                            );
                            
                            if (!isDuplicate) {
                                return {
                                    ...prev,
                                    [content.post_id]: [...prev[content.post_id], content]
                                };
                            }
                        }
                        return prev;
                    });
                }
            });
            

            // Add message handler for group events
            client.addMessageHandler('group_event', (content) => {
                if (content.group_id === groupId && content.creator_id !== currentUser.id) {
                    setEvents(prev => [...prev, content]);
                }
            });

            client.addMessageHandler('event_response', (content) => {
                console.log('Received event response:', content);

                if (content.event_id) {
                    // Update the events array with new counts
                    setEvents(prev =>
                        prev.map(event =>
                            event.id === content.event_id
                                ? {
                                    ...event,
                                    going_count: content.going_count,
                                    not_going_count: content.not_going_count,
                                    // Update user_response if this is the current user's response
                                    user_response: content.user_id === currentUser.id
                                        ? content.response
                                        : event.user_response
                                }
                                : event
                        )
                    );

                    // Also update userEventResponses state if needed
                    if (content.user_id === currentUser.id) {
                        setUserEventResponses(prev => ({
                            ...prev,
                            [content.event_id]: content.option_id
                        }));
                    }
                }
            });



            setWsClient(client);
        }

        // No cleanup needed as we want to keep the connection alive
    }, [currentUser, groupId]);

    // Scroll to bottom of messages when new messages arrive
    useEffect(() => {
        if (messagesEndRef.current && activeTab === 'chat') {
            messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [messages, activeTab]);

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

    
    // Add this useEffect to load comments when a post is selected
    useEffect(() => {
        if (selectedPost) {
            loadCommentsForPost(selectedPost);
        }
    }, [selectedPost]);


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

    // Handle sending a group message
    const handleSendMessage = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!newMessage.trim() || !groupId || !currentUser) return;

        try {
            // Try to send via WebSocket first
            let sentViaWebSocket = false;
            if (wsClient && wsClient.socket && wsClient.socket.readyState === WebSocket.OPEN) {
                console.log('Attempting to send group message via WebSocket');
                sentViaWebSocket = wsClient.sendGroupMessage(groupId, newMessage);
                console.log('WebSocket send result:', sentViaWebSocket);
            } else {
                console.log('WebSocket not available, using HTTP');
            }

            // If WebSocket failed or not available, use HTTP
            if (!sentViaWebSocket) {
                console.log('Sending group message via HTTP');
                await sendGroupMessage(groupId, newMessage);
            }

            // Add the message to the UI regardless of how it was sent
            const newMessageObj: GroupMessage = {
                id: Date.now(), // Temporary ID until we get the real one
                sender_id: currentUser.id,
                group_id: groupId,
                content: newMessage,
                created_at: new Date().toISOString(),
                sender: {
                    id: currentUser.id,
                    firstName: currentUser.firstName,
                    lastName: currentUser.lastName,
                    avatar: currentUser.avatar
                }
            };

            // Add the new message to the messages list
            setMessages(prev => [...prev, newMessageObj]);

            // Clear the input field
            setNewMessage('');
        } catch (error) {
            console.error('Error sending group message:', error);
        }
    };

    // Handle creating a post
    const handleCreatePost = async () => {
        if (!newPostContent.trim() || !groupId) return;

        try {
            console.log(currentUser);
            const newPost = await createGroupPost(groupId, newPostContent, newPostImage || undefined, currentUser);
            if (newPost) {
                // Only add to posts if it's not already there (might be added by WebSocket)
                setPosts(prev => {
                    // Check if this post is already in the list (by content and timestamp)
                    const isDuplicate = prev.some(p =>
                        p.content === newPostContent &&
                        (new Date().getTime() - new Date(p.created_at).getTime()) < 5000
                    );

                    if (!isDuplicate) {
                        return [newPost, ...prev];
                    }
                    return prev;
                });

                setNewPostContent('');
                setNewPostImage(null);
            }
        } catch (error) {
            console.error('Error creating post:', error);
        }
    };

    const [postComments, setPostComments] = useState<{ [postId: number]: any[] }>({});
    const [loadingComments, setLoadingComments] = useState<{ [postId: number]: boolean }>({});

    // Function to load comments for a post
    const loadCommentsForPost = async (postId: number) => {
        // Only load if not already loading
        if (loadingComments[postId]) return;

        setLoadingComments(prev => ({ ...prev, [postId]: true }));

        try {
            const comments = await fetchGroupPostComments(postId);
            setPostComments(prev => ({ ...prev, [postId]: comments }));
        } catch (error) {
            console.error('Error loading comments:', error);
        } finally {
            setLoadingComments(prev => ({ ...prev, [postId]: false }));
        }
    };

// Handle creating a comment
const handleCreateComment = async (postId: number) => {
    if (!newComment.trim()) return;

    try {
        // Try to send via WebSocket first
        let sentViaWebSocket = false;
        if (wsClient && wsClient.socket && wsClient.socket.readyState === WebSocket.OPEN) {
            console.log('Attempting to send group comment via WebSocket');
            sentViaWebSocket = wsClient.sendGroupComment(postId, newComment);
            console.log('WebSocket send result:', sentViaWebSocket);
            
            // If sent via WebSocket, update the local state immediately
            if (sentViaWebSocket && currentUser) {
                // Create a temporary comment object
                const tempComment = {
                    id: Date.now(), // Temporary ID
                    post_id: postId,
                    user_id: currentUser.id,
                    content: newComment,
                    created_at: new Date().toISOString(),
                    first_name: currentUser.firstName,
                    last_name: currentUser.lastName,
                    avatar: currentUser.avatar
                };
                
                // Update posts comment count
                setPosts(prev =>
                    prev.map(post =>
                        post.id === postId
                            ? { ...post, comment_count: (post.comment_count || 0) + 1 }
                            : post
                    )
                );
                
                // Add to comments
                setPostComments(prev => ({
                    ...prev,
                    [postId]: [...(prev[postId] || []), tempComment]
                }));
            }
        } else {
            // If WebSocket failed or not available, use HTTP
            console.log('Sending group comment via HTTP');
            await createGroupComment(postId, newComment);
            
            // Refresh comments for this post if using HTTP
            await loadCommentsForPost(postId);
        }

        // Clear the comment input
        setNewComment('');
    } catch (error) {
        console.error('Error creating comment:', error);
    }
};


    // Handle creating an event
    const handleCreateEvent = async () => {
        try {
            // First, validate that date and time are not empty
            if (!eventDate || !eventTime) {
                setEventError("Please select both date and time");
                return;
            }

            // Create a date string in the format that JavaScript can parse
            const dateTimeString = `${eventDate}T${eventTime}`;

            // Validate the date before converting to ISO string
            const eventDateTime = new Date(dateTimeString);

            if (isNaN(eventDateTime.getTime())) {
                setEventError("Invalid date or time format");
                return;
            }

            // Now it's safe to convert to ISO string
            const isoDateTime = eventDateTime.toISOString();

            // Define response options for the event
            const eventOptions = ["Going", "Maybe", "Not Going"];

            // Create the event
            const createdEvent = await createGroupEvent(
                groupId,
                eventTitle,
                eventDescription,
                isoDateTime,
                eventOptions
            );

            if (createdEvent) {
                // Only add to events if it's not already there (might be added by WebSocket)
                setEvents(prev => {
                    // Check if this event is already in the list (by title and timestamp)
                    const isDuplicate = prev.some(e =>
                        e.title === eventTitle &&
                        (new Date().getTime() - new Date(e.created_at).getTime()) < 5000
                    );

                    if (!isDuplicate) {
                        return [createdEvent, ...prev];
                    }
                    return prev;
                });

                // Reset the form
                setEventTitle('');
                setEventDescription('');
                setEventDate('');
                setEventTime('');
                setShowEventForm(false);
            } else {
                setEventError("Failed to create event. Please try again.");
            }
        } catch (error) {
            console.error("Error creating event:", error);
            setEventError("Failed to create event. Please try again.");
        }
    };

    const fetchGroupEvents = async () => {
        try {
            const response = await fetch(`http://localhost:8080/groups/events?group_id=${groupId}`, {
                method: 'GET',
                credentials: 'include'
            });

            if (response.ok) {
                const data = await response.json();
                if (data.success && data.events) {
                    setEvents(data.events);

                    // Initialize user responses from the fetched data
                    const userResponses: { [eventId: number]: number } = {};
                    data.events.forEach((event: any) => {
                        if (event.user_response_id) {
                            userResponses[event.id] = event.user_response_id;
                        }
                    });
                    setUserEventResponses(userResponses);
                }
            }
        } catch (error) {
            console.error('Error fetching events:', error);
        }
    };

    // Call this function when the tab changes to events
    useEffect(() => {
        if (groupId && activeTab === 'events') {
            fetchGroupEvents();
        }
    }, [groupId, activeTab]);




    // Handle responding to an event
    // Handle responding to an event
    const handleEventResponse = async (eventId: number, optionId: number) => {
        try {
            console.log(`Responding to event ${eventId} with option ${optionId}`);

            // Optimistically update the UI
            const optionText = optionId === 1 ? 'Going' : 'Not Going';

            setEvents(prev =>
                prev.map(event => {
                    if (event.id === eventId) {
                        // Calculate new counts
                        let goingCount = event.going_count || 0;
                        let notGoingCount = event.not_going_count || 0;

                        // If user already responded, adjust the old count down
                        if (event.user_response === 'Going') {
                            goingCount--;
                        } else if (event.user_response === 'Not Going') {
                            notGoingCount--;
                        }

                        // Adjust the new count up
                        if (optionText === 'Going') {
                            goingCount++;
                        } else if (optionText === 'Not Going') {
                            notGoingCount++;
                        }

                        return {
                            ...event,
                            going_count: goingCount,
                            not_going_count: notGoingCount,
                            user_response: optionText
                        };
                    }
                    return event;
                })
            );

            // Update the user responses state
            setUserEventResponses(prev => ({
                ...prev,
                [eventId]: optionId
            }));

            // Send the response to the server
            const success = await respondToEvent(eventId, optionId);

            if (!success) {
                console.error("Failed to record response");
                // Revert the optimistic update if the server request fails
                fetchGroupEvents();
            }
        } catch (error) {
            console.error("Error responding to event:", error);
            // Revert the optimistic update if there's an error
            fetchGroupEvents();
        }
    };




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
                                                className={`message ${message.sender_id === currentUser?.id ? 'sent' : 'received'}`}
                                            >
                                                {message.sender_id !== currentUser?.id && (
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

                                        <div className="message-content">
                                                 {message.content.match(/\.(jpeg|jpg|gif|png)$/i) ? (
                                                  // If the content is an image URL
                                             <img src={`http://localhost:8080/${message.content}`} alt="User uploaded content" />
                                             ) : (
  
                                                      message.content
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
  <input
    type="file"
    accept="image/*"
    style={{ display: 'none' }}
    ref={fileInputRef}
    onChange={handleGroupImageUpload}
  />
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
                                    value={newPostContent}
                                    onChange={(e) => setNewPostContent(e.target.value)}
                                ></textarea>
                                <div className="post-image-upload">
                                    <input
                                        type="text"
                                        placeholder="Image URL (optional)"
                                        value={newPostImage || ''}
                                        onChange={(e) => setNewPostImage(e.target.value || null)}
                                    />
                                </div>
                                <button
                                    className="post-button"
                                    onClick={handleCreatePost}
                                    disabled={!newPostContent.trim()}
                                >
                                    Post
                                </button>
                            </div>

                            {posts.length > 0 ? (
                                <div className="posts-list">
                                    {posts.map(post => (
                                        <div key={post.id} className="post-item">
                                            <div className="post-header">
                                                <div className="post-author">
                                                    <div className="author-avatar">
                                                        {post.avatar ? (
                                                            <img src={post.avatar} alt={`${post.first_name}'s avatar`} />
                                                        ) : (
                                                            <div className="avatar-placeholder">
                                                                {post.first_name ? post.first_name.charAt(0) : 'U'}
                                                            </div>
                                                        )}
                                                    </div>
                                                    <div className="author-info">
                                                        <span className="author-name">{post.first_name || 'Unknown'} {post.last_name || ''}</span>
                                                        <span className="post-time">{post.created_at ? new Date(post.created_at).toLocaleString() : 'Unknown date'}</span>
                                                    </div>
                                                </div>
                                            </div>
                                            <div className="post-content">{post.content || ''}</div>
                                            {post.image && (
                                                <div className="post-image">
                                                    <img src={post.image} alt="Post attachment" />
                                                </div>
                                            )}
                                            <div className="post-footer">
                                                <button
                                                    className="comment-button"
                                                    onClick={() => setSelectedPost(selectedPost === post.id ? null : post.id)}
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                                        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                                                    </svg>
                                                    {post.comment_count || 0} Comments
                                                </button>
                                            </div>

                                            {/* Comment section */}
                                            {selectedPost === post.id && (
                                                <div className="comments-section">
                                                    {/* Comment form */}
                                                    <div className="comment-form">
                                                        <textarea
                                                            className="comment-input"
                                                            placeholder="Write a comment..."
                                                            value={newComment}
                                                            onChange={(e) => setNewComment(e.target.value)}
                                                        ></textarea>
                                                        <button
                                                            className="comment-submit-button"
                                                            onClick={() => handleCreateComment(post.id)}
                                                            disabled={!newComment.trim()}
                                                        >
                                                            Comment
                                                        </button>
                                                    </div>
                                                    
                                                    {/* Comments list */}
                                                    <div className="comments-list">
                                                        {loadingComments[post.id] ? (
                                                            <div className="loading-comments">Loading comments...</div>
                                                        ) : postComments[post.id]?.length > 0 ? (
                                                            postComments[post.id].map(comment => (
                                                                <div key={comment.id} className="comment-item">
                                                                    <div className="comment-author">
                                                                        <div className="author-avatar">
                                                                            {comment.avatar ? (
                                                                                <img src={comment.avatar} alt={`${comment.first_name}'s avatar`} />
                                                                            ) : (
                                                                                <div className="avatar-placeholder">
                                                                                    {comment.first_name ? comment.first_name.charAt(0) : 'U'}
                                                                                </div>
                                                                            )}
                                                                        </div>
                                                                        <span className="author-name">{comment.first_name} {comment.last_name}</span>
                                                                    </div>
                                                                    <div className="comment-content">{comment.content}</div>
                                                                    <div className="comment-time">
                                                                        {new Date(comment.created_at).toLocaleString()}
                                                                    </div>
                                                                </div>
                                                            ))
                                                        ) : (
                                                            <div className="no-comments">No comments yet. Be the first to comment!</div>
                                                        )}
                                                    </div>
                                                </div>
                                            )}
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
                                <button
                                    className="create-event-button"
                                    onClick={() => setShowEventForm(!showEventForm)}
                                >
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

                            {/* Event creation form */}
                            {showEventForm && (
                                <div className="event-form">
                                    <h3>Create New Event</h3>

                                    {eventError && (
                                        <div className="error-message">
                                            {eventError}
                                        </div>
                                    )}

                                    <div className="form-group">
                                        <label>Event Title</label>
                                        <input
                                            type="text"
                                            value={eventTitle}
                                            onChange={(e) => setEventTitle(e.target.value)}
                                            placeholder="Enter event title"
                                        />
                                    </div>

                                    <div className="form-group">
                                        <label>Description</label>
                                        <textarea
                                            value={eventDescription}
                                            onChange={(e) => setEventDescription(e.target.value)}
                                            placeholder="Describe your event"
                                            rows={3}
                                        ></textarea>
                                    </div>

                                    <div className="form-group">
                                        <label>Date</label>
                                        <input
                                            type="date"
                                            value={eventDate}
                                            onChange={(e) => setEventDate(e.target.value)}
                                        />
                                    </div>

                                    <div className="form-group">
                                        <label>Time</label>
                                        <input
                                            type="time"
                                            value={eventTime}
                                            onChange={(e) => setEventTime(e.target.value)}
                                        />
                                    </div>

                                    <div className="form-actions">
                                        <button
                                            className="cancel-button"
                                            onClick={() => {
                                                setShowEventForm(false);
                                                setEventError(''); // Clear error when canceling
                                            }}
                                        >
                                            Cancel
                                        </button>
                                        <button
                                            className="create-button"
                                            onClick={handleCreateEvent}
                                            disabled={!eventTitle.trim() || !eventDescription.trim() || !eventDate || !eventTime}
                                        >
                                            Create Event
                                        </button>
                                    </div>
                                </div>

                            )}

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
                                                <div className="event-stats">
                                                    <span className="going-count">{event.going_count || 0} going</span>
                                                    <span className="not-going-count">{event.not_going_count || 0} not going</span>
                                                </div>

                                            </div>
                                            <div className="event-actions">
                                                <button
                                                    className={`event-action-button ${event.user_response === 'Going' ? 'active' : ''}`}
                                                    onClick={() => handleEventResponse(event.id, 1)}
                                                >
                                                    Going
                                                </button>
                                                <button
                                                    className={`event-action-button ${event.user_response === 'Not Going' ? 'active' : ''}`}
                                                    onClick={() => handleEventResponse(event.id, 2)}
                                                >
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
                                ×
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