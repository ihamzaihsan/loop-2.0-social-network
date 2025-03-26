import { WebSocketClientInterface, MessageContent, MessageHandler } from './types';

let globalWsClient: WebSocketClient | null = null;

export class WebSocketClient implements WebSocketClientInterface {
    private static instance: WebSocketClient | null = null;
    
    static getInstance(): WebSocketClient {
        if (!WebSocketClient.instance) {
            WebSocketClient.instance = new WebSocketClient();
        
            // Automatically connect if we have a session token
            const sessionToken = localStorage.getItem('sessionToken');
            if (sessionToken) {
                WebSocketClient.instance.connect();
            }
        }
        return WebSocketClient.instance;
    }

    
    static resetInstance(): void {
        if (WebSocketClient.instance) {
            if (WebSocketClient.instance.socket) {
                WebSocketClient.instance.socket.close(1000, "Reset instance");
            }
            WebSocketClient.instance = null;
        }
    }


    socket: WebSocket | null = null;
    messageHandlers: Map<string, (content: any) => void> = new Map();
    messageHistory: Map<number, MessageContent[]>;
    currentChatUser: number | null;
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

    console.log('Attempting WebSocket connection...');
    const sessionToken = localStorage.getItem('sessionToken') || document.cookie.replace(/(?:(?:^|.*;\s*)session_token\s*\=\s*([^;]*).*$)|^.*$/, "$1");

    if (!sessionToken) {
        console.error('No session token found, skipping WebSocket connection');
        return;
    }

    // Include session token in the URL for authentication
    this.socket = new WebSocket(`ws://localhost:8080/ws?token=${encodeURIComponent(sessionToken)}`);

    this.socket.onopen = () => {
        console.log('WebSocket connected successfully');
    
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
    
        if (message.type === 'pong') {
            console.log('Ping-pong successful, connection is working properly');
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
        this.clearPingInterval();
    
        // Only attempt to reconnect if the close wasn't intentional (code 1000)
        if (event.code !== 1000) {
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
    if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout);
        this.reconnectTimeout = null;
    }
    
    if (this.socket) {
        this.socket.close(1000, "Intentional close");
        this.socket = null;
    }
    }
}
