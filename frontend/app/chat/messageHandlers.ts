import { WebSocketClientInterface } from '../webscoket/types';

interface User {
    id: number;
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
        
        if (data.success && data.users) {
            return data.users;
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

// Setup WebSocket message handler
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
