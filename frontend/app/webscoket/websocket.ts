import { webSocketURL } from "../../utils/api";
import { WebSocketClientInterface, MessageContent } from './types';

export class WebSocketClient implements WebSocketClientInterface {
    private static instance: WebSocketClient | null = null;
    
    static getInstance(): WebSocketClient {
        if (!WebSocketClient.instance) {
            WebSocketClient.instance = new WebSocketClient();
            // Remove credentials left by older versions; authentication uses cookies.
            localStorage.removeItem('sessionToken');
        

        }
        return WebSocketClient.instance;
    }

    
    static resetInstance(): void {
        if (WebSocketClient.instance) {
            WebSocketClient.instance.close();
            WebSocketClient.instance = null;
        }
        window.dispatchEvent(new Event('session_ended'));
    }


    private stopped = false;
    socket: WebSocket | null = null;
    messageHandlers: Map<string, (content: any) => void> = new Map();
    messageHistory: Map<number, MessageContent[]>;
    currentChatUser: number | null;
    currentUserId: number | null = null;
    onlineUsers: Map<number, boolean>;
    pingInterval: NodeJS.Timeout | null;
    reconnectTimeout: NodeJS.Timeout | null = null;

    constructor() {
        this.socket = null;
        this.messageHandlers = new Map();
        this.messageHistory = new Map();
        this.currentChatUser = null;
        this.onlineUsers = new Map();
        this.pingInterval = null;
    }

    addMessageHandler(type: string, handler: (content: any) => void): void {
        this.messageHandlers.set(type, handler);
    }
    
    connect(): void {
        // Don't reconnect if already connected or connecting
        if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
            console.log('WebSocket already connected or connecting, skipping reconnection');
            return;
        }

        this.stopped = false;
        if (this.reconnectTimeout) clearTimeout(this.reconnectTimeout);
        // Authentication uses the HTTP-only session cookie, never a URL token.
        this.socket = new WebSocket(webSocketURL());
        const connection = this.socket;

        this.socket.onopen = () => {
            this.clearPingInterval();
            this.startPingInterval();
            window.dispatchEvent(new CustomEvent('connection_status', { detail: { connected: true } }));
            window.dispatchEvent(new Event('realtime_reconnected'));
        
            // Send a ping to test the connection
            if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify({
                    type: 'ping',
                    content: { timestamp: new Date().toISOString() }
                }));
            }
        };

        this.socket.onmessage = (event: MessageEvent) => {
            try {
                const message = JSON.parse(event.data);
                console.log('Received WebSocket message:', message);
            
                // Handle session invalidation message
                if (message.type === 'session_invalidated') {
                    console.log('Session invalidated from another device');
                    
                    // Clear local storage and cookies
                    localStorage.removeItem('sessionToken');
                    
                    // Close the WebSocket connection
                    if (this.socket) {
                        this.socket.close(1000, "Session invalidated");
                    }
                    
                    // Reset the WebSocket instance
                    WebSocketClient.resetInstance();
                    
                    // Redirect to login page with reason parameter
                    window.location.href = new URL('/login?reason=session_expired', window.location.origin).toString();
                    return;
                }
            
                if (message.type === 'pong') {
                    console.log('Ping-pong successful, connection is working properly');
                }

                if (message.type === 'presence_snapshot') {
                    this.currentUserId = message.content.user_id;
                    this.onlineUsers = new Map((message.content.user_ids as number[]).map(id => [id, true]));
                    window.dispatchEvent(new Event('presence_update'));
                }
                if (message.type === 'presence_changed') {
                    this.onlineUsers.set(message.content.user_id, message.content.online);
                    window.dispatchEvent(new Event('presence_update'));
                }
                if (message.type === 'typing_status') {
                    window.dispatchEvent(new CustomEvent('typing_status', { detail: message.content }));
                }
                if (message.type === 'private_message') {
                    window.dispatchEvent(new CustomEvent('private_message', { detail: message.content }));
                }
                let resource = message.type === 'resource_changed' ? message.content.resource : null;
                if (['private_message', 'messages_read'].includes(message.type)) resource = 'chat';
                if (['notification', 'notification_changed', 'follow_request_handled'].includes(message.type)) resource = 'notifications';
                if (resource) window.dispatchEvent(new CustomEvent('realtime_changed', { detail: { resource } }));

                // Handle notification messages for all notification types
                if (message.type === 'notification') {
                    // Trigger notification event
                    const event = new CustomEvent('notification', { detail: message.content });
                    window.dispatchEvent(event);
                    

                }

                // Handle follow status updates
                if (message.type === 'follow_status_update') {
                    // Trigger follow status update event
                    const event = new CustomEvent('follow_status_update', { detail: message.content });
                    window.dispatchEvent(event);
                }

                if (['social_graph_updated', 'follow_request_handled', 'notification_changed'].includes(message.type)) {
                    window.dispatchEvent(new CustomEvent(message.type, { detail: message.content }));
                }
                if (['social_graph_updated', 'follow_request_handled', 'follow_status_update'].includes(message.type)
                    || (message.type === 'notification' && message.content?.type?.startsWith('follow_'))) {
                    window.dispatchEvent(new Event('social_update'));
                }
            
                const handler = this.messageHandlers.get(message.type);
                if (handler) {
                    handler(message.content);
                }
            } catch (error) {
                console.error('Error handling WebSocket message:', error);
            }
        };

        this.socket.onclose = (event: CloseEvent) => {
            console.warn('WebSocket connection closed:', event.reason || 'Unknown reason');
            if (this.socket !== connection) return;
            this.clearPingInterval();
            this.onlineUsers.clear();
            window.dispatchEvent(new Event('presence_update'));
            window.dispatchEvent(new CustomEvent('connection_status', { detail: { connected: false } }));
        
            // Only attempt to reconnect if the close wasn't intentional (code 1000)
            if (!this.stopped && event.code !== 1000) {
                console.log('Attempting to reconnect in 5 seconds...');
                this.reconnectTimeout = setTimeout(() => this.connect(), 5000);
            }
        };

        this.socket.onerror = (error: Event) => {
            console.error('WebSocket error:', error);
        };
    }
    
    startPingInterval(): void {
        this.pingInterval = setInterval(() => {
            if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify({
                    type: 'ping',
                    content: { timestamp: new Date().toISOString() }
                }));
            }
        }, 30000); // Send ping every 30 seconds
    }
    
    clearPingInterval(): void {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
    }
    
    sendMessageWithType(type: string, content: any): boolean {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('WebSocket is not connected');
            return false;
        }
        
        try {
            this.socket.send(JSON.stringify({
                type: type,
                content: content
            }));
            return true;
        } catch (error) {
            console.error('Error sending message via WebSocket:', error);
            return false;
        }
    }
    
    sendMessage(receiverId: number, content: string): boolean {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('WebSocket is not connected, readyState:', this.socket?.readyState);
            return false;
        }
    
        try {
            const message = {
                type: 'private_message',
                content: {
                    receiver_id: receiverId,
                    content: content
                }
            };
            
            console.log('Sending WebSocket message:', JSON.stringify(message));
            this.socket.send(JSON.stringify(message));
            
            console.log('Message sent successfully via WebSocket');
            return true;
        } catch (error) {
            console.error('Error sending message via WebSocket:', error);
            return false;
        }
    }
    
    sendGroupMessage(groupId: number, content: string): boolean {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('WebSocket is not connected, readyState:', this.socket?.readyState);
            return false;
        }
    
        try {
            const message = {
                type: 'group_message',
                content: {
                    group_id: groupId,
                    content: content
                }
            };
            
            console.log('Sending group message via WebSocket:', JSON.stringify(message));
            this.socket.send(JSON.stringify(message));
            
            console.log('Group message sent successfully via WebSocket');
            return true;
        } catch (error) {
            console.error('Error sending group message via WebSocket:', error);
            return false;
        }
    }
    
    setTypingStatus(receiverId: number, isTyping: boolean): void {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            return;
        }
        
        try {
            this.socket.send(JSON.stringify({
                type: 'typing_status',
                content: {
                    receiver_id: receiverId,
                    is_typing: isTyping
                }
            }));
        } catch (error) {
            console.error('Error sending typing status:', error);
        }
    }
    
    close(): void {
        this.stopped = true;
        this.clearPingInterval();
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }
        
        if (this.socket) {
            this.socket.close(1000, "Intentional close");
            this.socket = null;
        }
    }

    sendGroupPost(groupId: number, content: string, image?: string): boolean {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('WebSocket is not connected, readyState:', this.socket?.readyState);
            return false;
        }

        try {
            const message = {
                type: 'group_post',
                content: {
                    group_id: groupId,
                    content: content,
                    image: image
                }
            };
            
            console.log('Sending group post via WebSocket:', JSON.stringify(message));
            this.socket.send(JSON.stringify(message));
            
            console.log('Group post sent successfully via WebSocket');
            return true;
        } catch (error) {
            console.error('Error sending group post via WebSocket:', error);
            return false;
        }
    }

    sendGroupEvent(groupId: number, title: string, description: string, eventTime: string, options?: string[]): boolean {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('WebSocket is not connected, readyState:', this.socket?.readyState);
            return false;
        }

        try {
            const message = {
                type: 'group_event',
                content: {
                    group_id: groupId,
                    title: title,
                    description: description,
                    event_time: eventTime,
                    options: options || ["Going", "Not Going"]
                }
            };
            
            console.log('Sending group event via WebSocket:', JSON.stringify(message));
            this.socket.send(JSON.stringify(message));
            
            console.log('Group event sent successfully via WebSocket');
            return true;
        } catch (error) {
            console.error('Error sending group event via WebSocket:', error);
            return false;
        }
    }

    sendEventResponse(eventId: number, optionId: number): boolean {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('WebSocket is not connected, readyState:', this.socket?.readyState);
            return false;
        }

        try {
            const message = {
                type: 'event_response',
                content: {
                    event_id: eventId,
                    option_id: optionId
                }
            };
            
            console.log('Sending event response via WebSocket:', JSON.stringify(message));
            this.socket.send(JSON.stringify(message));
            
            console.log('Event response sent successfully via WebSocket');
            return true;
        } catch (error) {
            console.error('Error sending event response via WebSocket:', error);
            return false;
        }
    }
    sendGroupComment(postId: number, content: string): boolean {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('WebSocket is not connected, readyState:', this.socket?.readyState);
            return false;
        }
    
        try {
            const message = {
                type: 'group_comment',
                content: {
                    post_id: postId,
                    content: content
                }
            };
            
            console.log('Sending group comment via WebSocket:', JSON.stringify(message));
            this.socket.send(JSON.stringify(message));
            
            console.log('Group comment sent successfully via WebSocket');
            return true;
        } catch (error) {
            console.error('Error sending group comment via WebSocket:', error);
            return false;
        }
    }
    
    sendGroupMembershipUpdate(groupId: number, action: string): boolean {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('WebSocket is not connected, readyState:', this.socket?.readyState);
            return false;
        }

        try {
            const message = {
                type: 'group_membership_update',
                content: {
                    group_id: groupId,
                    action: action
                }
            };
            
            console.log('Sending group membership update via WebSocket:', JSON.stringify(message));
            this.socket.send(JSON.stringify(message));
            
            console.log('Group membership update sent successfully via WebSocket');
            return true;
        } catch (error) {
            console.error('Error sending group membership update via WebSocket:', error);
            return false;
        }
    }
}
