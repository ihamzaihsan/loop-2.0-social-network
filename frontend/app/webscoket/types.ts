export interface MessageContent {
    id?: number;
    sender_id?: number;
    receiver_id?: number;
    group_id?: number;  // Add group_id for group messages
    content?: string;
    created_at?: string;
    sender?: string | {
        id: number;
        firstName: string;
        lastName: string;
        avatar?: string;
    };
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
    sendGroupMessage(groupId: number, content: string): boolean;
    sendGroupPost(groupId: number, content: string, image?: string): boolean;
    sendGroupEvent(groupId: number, title: string, description: string, eventTime: string, options?: string[]): boolean;
    sendEventResponse(eventId: number, optionId: number): boolean;
    setTypingStatus(receiverId: number, isTyping: boolean): void;
}