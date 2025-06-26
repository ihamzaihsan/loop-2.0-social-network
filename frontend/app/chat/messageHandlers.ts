import { WebSocketClientInterface } from '../webscoket/types';
import { WebSocketClient } from '../webscoket/websocket';

interface User {
    id: number;
    followedID?: number;  // Add this optional property
    firstName: string;
    lastName: string;
    nickname?: string;
    avatar?: string;
}

interface Message {
    id: number;
    sender_id: number;
    receiver_id: number;
    content: string;
    created_at: string;
    is_read: boolean;
    sender: {
        id: number;
        first_name: string;
        last_name: string;
        avatar?: string;
    };
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

interface ChatContact {
    id: number;
    firstName: string;
    lastName: string;
    nickname?: string;
    avatar?: string;
    lastMessage?: string;
    lastMessageTime?: string;
    unreadCount?: number;
}

// Fetch chat contacts (users with message history)
export const fetchChatContacts = async (): Promise<ChatContact[]> => {
    try {
        const response = await fetch('http://localhost:8080/chat/contacts', {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) {
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
export const fetchMessages = async (contactId: number): Promise<Message[]> => {
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

// Fetch group messages
export const fetchGroupMessages = async (groupId: number): Promise<GroupMessage[]> => {
    try {
        const response = await fetch(`http://localhost:8080/groups/messages?id=${groupId}`, {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Failed to fetch group messages');
        }

        const data = await response.json();
        if (data.success && data.messages) {
            return data.messages;
        }
        return [];
    } catch (error) {
        console.error('Error fetching group messages:', error);
        return [];
    }
};

// Fetch group posts
export const fetchGroupPosts = async (groupId: number): Promise<any[]> => {
    try {
        const response = await fetch(`http://localhost:8080/groups/posts?group_id=${groupId}`, {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Failed to fetch group posts');
        }

        const data = await response.json();
        if (data.success && data.posts) {
            return data.posts;
        }
        return [];
    } catch (error) {
        console.error('Error fetching group posts:', error);
        return [];
    }
};

// Fetch group events
export const fetchGroupEvents = async (groupId: number): Promise<any[]> => {
    try {
        const response = await fetch(`http://localhost:8080/groups/events?group_id=${groupId}`, {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Failed to fetch group events');
        }

        const data = await response.json();
        if (data.success && data.events) {
            return data.events;
        }
        return [];
    } catch (error) {
        console.error('Error fetching group events:', error);
        return [];
    }
};

// Fetch followed users
export const fetchFollowedUsers = async (): Promise<User[]> => {
    try {
        // Use the new endpoint specifically for fetching followed users
        const response = await fetch('http://localhost:8080/chat/followed', {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Failed to fetch followed users');
        }

        const data = await response.json();
        console.log("Followed users response:", data);
        
        // Fix: Check for data.following instead of data.users
        if (data.success && data.following) {
            return data.following;
        }
        return [];
    } catch (error) {
        console.error('Error fetching followed users:', error);
        return [];
    }
};

// Send a message
export const sendMessage = async (
    receiverId: number,
    content: string
): Promise<Message | null> => {
    try {
        const response = await fetch('http://localhost:8080/messages', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                receiver_id: receiverId,
                content: content
            }),
        });

        if (!response.ok) {
            throw new Error('Failed to send message');
        }

        const data = await response.json();
        if (data.success && data.message) {
            return data.message;
        }
        return null;
    } catch (error) {
        console.error('Error sending message:', error);
        return null;
    }
};
    
    export const setupMessageHandler = (
        wsClient: WebSocketClientInterface,
        currentUser: User | null,
        selectedContact: ChatContact | null,
        setMessages: React.Dispatch<React.SetStateAction<Message[]>>,
        setContacts: React.Dispatch<React.SetStateAction<ChatContact[]>>
    ) => {
        wsClient.addMessageHandler('private_message', (content) => {
            if (selectedContact && content.sender_id === selectedContact.id) {
                // Make sure all required properties have fallback values
                const newMessage: Message = {
                    id: content.id ?? 0, // Fallback to 0 if undefined
                    sender_id: content.sender_id ?? 0, // Fallback to 0 if undefined
                    receiver_id: currentUser?.id ?? 0,
                    content: content.content ?? "", // Fallback to empty string if undefined
                    created_at: content.created_at ?? new Date().toISOString(), // Fallback to current time if undefined
                    is_read: false,
                    sender: {
                        id: content.sender_id ?? 0, // Fallback to 0 if undefined
                        first_name: typeof content.sender === 'string' ? content.sender.split(' ')[0] : "",
                        last_name: typeof content.sender === 'string' ? (content.sender.split(' ')[1] || "") : "",
                        avatar: undefined
                    }
                };
                
                setMessages(prev => [...prev, newMessage]);
            }
            
            // Update contacts list with new message
            setContacts(prev => {
                const updatedContacts = [...prev];
                const contactIndex = updatedContacts.findIndex(c => c.id === content.sender_id);
                
                if (contactIndex >= 0) {
                    updatedContacts[contactIndex] = {
                        ...updatedContacts[contactIndex],
                        lastMessage: content.content ?? "",
                        lastMessageTime: content.created_at ?? new Date().toISOString(),
                        unreadCount: (updatedContacts[contactIndex].unreadCount || 0) + 1
                    };
                }
                
                return updatedContacts;
            });
        });
    };

    // Setup WebSocket message handler for group messages
    export const setupGroupMessageHandler = (
        wsClient: WebSocketClientInterface,
        currentUser: User | null,
        currentGroupId: number | null,
        setGroupMessages: React.Dispatch<React.SetStateAction<GroupMessage[]>>
    ) => {
        wsClient.addMessageHandler('group_message', (content) => {
            if (currentGroupId && content.group_id === currentGroupId) {
                // Make sure all required properties have fallback values
                const newMessage: GroupMessage = {
                    id: content.id ?? 0,
                    sender_id: content.sender_id ?? 0,
                    group_id: content.group_id ?? 0,
                    content: content.content ?? "",
                    created_at: content.created_at ?? new Date().toISOString(),
                    sender: {
                        id: content.sender_id ?? 0,
                        firstName: typeof content.sender === 'string' ? content.sender.split(' ')[0] : "",
                        lastName: typeof content.sender === 'string' ? (content.sender.split(' ')[1] || "") : "",
                        avatar: undefined
                    }
                };
                
                setGroupMessages(prev => [...prev, newMessage]);
            }
        });
    };

    // Setup WebSocket message handler for group posts
    export const setupGroupPostHandler = (
        wsClient: WebSocketClientInterface,
        currentGroupId: number | null,
        setPosts: React.Dispatch<React.SetStateAction<any[]>>
    ) => {
        wsClient.addMessageHandler('group_post', (content) => {
            if (currentGroupId && content.group_id === currentGroupId) {
                setPosts(prev => [content, ...prev]);
            }
        });
    };

    // Setup WebSocket message handler for group post comments
    export const setupGroupCommentHandler = (
        wsClient: WebSocketClientInterface,
        currentPostId: number | null,
        setComments: React.Dispatch<React.SetStateAction<any[]>>
    ) => {
        wsClient.addMessageHandler('group_comment', (content) => {
            if (currentPostId && content.post_id === currentPostId) {
                setComments(prev => [...prev, content]);
            }
        });
    };

    // Setup WebSocket message handler for group events
    export const setupGroupEventHandler = (
        wsClient: WebSocketClientInterface,
        currentGroupId: number | null,
        setEvents: React.Dispatch<React.SetStateAction<any[]>>
    ) => {
        wsClient.addMessageHandler('group_event', (content) => {
            if (currentGroupId && content.group_id === currentGroupId) {
                setEvents(prev => [...prev, content]);
            }
        });
    };

    // Setup WebSocket message handler for event responses
    export const setupEventResponseHandler = (
        wsClient: WebSocketClientInterface,
        setEvents: React.Dispatch<React.SetStateAction<any[]>>
    ) => {
        wsClient.addMessageHandler('event_response', (content) => {
            setEvents(prev => 
                prev.map(event => 
                    event.id === content.event_id 
                        ? {
                            ...event,
                            going_count: content.going_count,
                            not_going_count: content.not_going_count,
                            user_response: content.user_id === content.current_user_id ? content.response : event.user_response
                        } 
                        : event
                )
            );
        });
    };

    export const sendGroupMessage = async (
        groupId: number,
        content: string
        ): Promise<any | null> => {
        try {
            const response = await fetch('http://localhost:8080/messages', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                group_id: groupId,
                content: content
            }),
            });
        
            if (!response.ok) {
            throw new Error('Failed to send group message');
            }
        
            const data = await response.json();
            if (data.success && data.message) {
            return data.message;
            }
            return null;
        } catch (error) {
            console.error('Error sending group message:', error);
            return null;
        }
        };
        
        // Create a group post with file upload
        export const createGroupPostWithFile = async (
        groupId: number,
        content: string,
        imageFile?: File,
        currentUser?: User | null
        ): Promise<any | null> => {
            try {
                console.log('Creating group post with file upload');

                const formData = new FormData()
                formData.append('group_id', groupId.toString())
                formData.append('content', content)
                if (imageFile) {
                    formData.append('image', imageFile)
                }

                const response = await fetch('http://localhost:8080/groups/posts/create', {
                    method: 'POST',
                    credentials: 'include',
                    body: formData,
                });

                if (!response.ok) {
                    const errorText = await response.text()
                    throw new Error(`Failed to create group post: ${errorText}`)
                }

                const data = await response.json();
                if (data.success) {
                    return {
                        id: data.post_id,
                        content: content,
                        image: imageFile ? URL.createObjectURL(imageFile) : null,
                        created_at: new Date().toISOString(),
                        first_name: currentUser?.firstName || 'Unknown',
                        last_name: currentUser?.lastName || '',
                        avatar: currentUser?.avatar || null,
                        comment_count: 0
                    };
                }
                return null;
            } catch (error) {
                console.error('Error creating group post with file:', error);
                throw error;
            }
        };

        // Create a group post (legacy function for URL-based images)
        export const createGroupPost = async (
        groupId: number,
        content: string,
        image?: string,
        currentUser?: User | null
        ): Promise<any | null> => {
            try {
                // Try to send via WebSocket first
                const wsClient = WebSocketClient.getInstance();
                let sentViaWebSocket = false;

                if (wsClient && wsClient.socket && wsClient.socket.readyState === WebSocket.OPEN) {
                    console.log('Attempting to send group post via WebSocket');
                    sentViaWebSocket = wsClient.sendGroupPost(groupId, content, image);
                    console.log('WebSocket send result:', sentViaWebSocket);
                } else {
                    console.log('WebSocket not available, using HTTP');
                }

                // If WebSocket failed or not available, use HTTP
                if (!sentViaWebSocket) {
                    console.log('Sending group post via HTTP');
                    const response = await fetch('http://localhost:8080/groups/posts/create', {
                        method: 'POST',
                        credentials: 'include',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            group_id: groupId,
                            content: content,
                            image: image
                        }),
                    });

                    if (!response.ok) {
                        throw new Error('Failed to create group post');
                    }
                    
                    const data = await response.json();
                    if (data.success && data.post_id) {
                        // Return a constructed post object since the API might not return the full post
                        return {
                            id: data.post_id,
                            user_id: currentUser?.id,
                            content: content,
                            image: image,
                            created_at: new Date().toISOString(),
                            comment_count: 0,
                            first_name: currentUser?.firstName || "",
                            last_name: currentUser?.lastName || "",
                            avatar: currentUser?.avatar
                        };
                    }
                } else {
                    // If sent via WebSocket, return a temporary object with user info
                    return {
                        id: Date.now(), // Temporary ID
                        user_id: currentUser?.id,
                        content: content,
                        image: image,
                        created_at: new Date().toISOString(),
                        comment_count: 0,
                        first_name: currentUser?.firstName || "",
                        last_name: currentUser?.lastName || "",
                        avatar: currentUser?.avatar
                    };
                }
                
                return null;
            } catch (error) {
                console.error('Error creating group post:', error);
                return null;
            }
        };
        
        // Create a comment on a group post
        export const createGroupComment = async (
        postId: number,
        content: string
        ): Promise<any | null> => {
        try {
            const response = await fetch('http://localhost:8080/groups/posts/comments/create', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                post_id: postId,
                content: content
            }),
            });
        
            if (!response.ok) {
            throw new Error('Failed to create comment');
            }
        
            const data = await response.json();
            if (data.success) {
            return {
                id: data.comment_id,
                content: content,
                created_at: new Date().toISOString()
            };
            }
            return null;
        } catch (error) {
            console.error('Error creating comment:', error);
            return null;
        }
        };
        
        // Create a group event
        export const createGroupEvent = async (
            groupId: number,
            title: string,
            description: string,
            eventTime: string,
            options: string[] = ["Going", "Not Going"] // Default options if none provided
        ): Promise<any | null> => {
            try {
                // Try to send via WebSocket first
                const wsClient = WebSocketClient.getInstance();
                let sentViaWebSocket = false;
                
                if (wsClient && wsClient.socket && wsClient.socket.readyState === WebSocket.OPEN) {
                    console.log('Attempting to send group event via WebSocket');
                    sentViaWebSocket = wsClient.sendGroupEvent(groupId, title, description, eventTime, options);
                    console.log('WebSocket send result:', sentViaWebSocket);
                } else {
                    console.log('WebSocket not available, using HTTP');
                }
                
                // If WebSocket failed or not available, use HTTP
                if (!sentViaWebSocket) {
                    console.log('Sending group event via HTTP');
                    const response = await fetch('http://localhost:8080/groups/events/create', {
                        method: 'POST',
                        credentials: 'include',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            group_id: groupId,
                            title: title,
                            description: description,
                            event_time: eventTime,
                            options: options
                        }),
                    });
                    
                    if (!response.ok) {
                        const errorData = await response.json().catch(() => ({}));
                        console.error("Server error response:", errorData);
                        throw new Error('Failed to create event');
                    }
                    
                    const data = await response.json();
                    if (data.success && data.event_id) {
                        // Return a constructed event object
                        return {
                            id: data.event_id,
                            title: title,
                            description: description,
                            event_time: eventTime,
                            created_at: new Date().toISOString(),
                            options: options
                        };
                    }
                } else {
                    // If sent via WebSocket, return a temporary object
                    // The real object will come through the WebSocket handler
                    return {
                        id: Date.now(), // Temporary ID
                        title: title,
                        description: description,
                        event_time: eventTime,
                        created_at: new Date().toISOString(),
                        options: options
                    };
                }
                
                return null;
            } catch (error) {
                console.error('Error creating event:', error);
                return null;
            }
        };
        
        // Respond to a group event
        export const respondToEvent = async (
            eventId: number,
            optionId: number
        ): Promise<boolean> => {
            try {
                console.log(`Sending event response: event_id=${eventId}, option_id=${optionId}`);
                
                // Try to send via WebSocket first
                const wsClient = WebSocketClient.getInstance();
                let sentViaWebSocket = false;
                
                if (wsClient && wsClient.socket && wsClient.socket.readyState === WebSocket.OPEN) {
                    console.log('Attempting to send event response via WebSocket');
                    sentViaWebSocket = wsClient.sendEventResponse(eventId, optionId);
                    console.log('WebSocket send result:', sentViaWebSocket);
                } else {
                    console.log('WebSocket not available, using HTTP');
                }
                
                // If WebSocket failed or not available, use HTTP
                if (!sentViaWebSocket) {
                    console.log('Sending event response via HTTP');
                    const response = await fetch('http://localhost:8080/groups/events/respond', {
                        method: 'POST',
                        credentials: 'include',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            event_id: eventId,
                            option_id: optionId
                        }),
                    });
                    
                    if (!response.ok) {
                        const errorText = await response.text();
                        console.error(`Failed to respond to event: ${response.status} ${response.statusText}`, errorText);
                        throw new Error('Failed to respond to event');
                    }
                    
                    const data = await response.json();
                    console.log("Response data:", data);
                    return data.success === true;
                }
                
                return sentViaWebSocket;
            } catch (error) {
                console.error('Error responding to event:', error);
                return false;
            }
        };

        // Add this function to fetch comments for a group post
        export const fetchGroupPostComments = async (postId: number): Promise<any[]> => {
            try {
                const response = await fetch(`http://localhost:8080/groups/posts/comments?post_id=${postId}`, {
                    method: 'GET',
                    credentials: 'include'
                });

                if (!response.ok) {
                    throw new Error('Failed to fetch post comments');
                }

                const data = await response.json();
                if (data.success && data.comments) {
                    return data.comments;
                }
                return [];
            } catch (error) {
                console.error('Error fetching post comments:', error);
                return [];
            }
        };
        
        