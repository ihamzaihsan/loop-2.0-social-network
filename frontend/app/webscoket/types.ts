export interface MessageContent {
    id?: number;
    sender_id?: number;
    receiver_id?: number;
    content?: string;
    created_at?: string;
    sender?: string;
    timestamp?: string;
    is_typing?: boolean;
    post_id?: number;
  [key: string]: any; // Allow for additional properties
}

export type MessageHandler = (content: MessageContent) => void;

export interface WebSocketClientInterface {
    socket: WebSocket | null;
    messageHandlers: Map<string, MessageHandler>;
    messageHistory: Map<number, MessageContent[]>;
    currentChatUser: number | null;
    onlineUsers: Map<number, boolean>;
    pingInterval: NodeJS.Timeout | null;
    
    addMessageHandler(type: string, handler: MessageHandler): void;
    connect(): void;
    startPingInterval(): void;
    clearPingInterval(): void;
    sendMessage(receiverId: number, content: string): boolean;
    setTypingStatus(receiverId: number, isTyping: boolean): void;
}
