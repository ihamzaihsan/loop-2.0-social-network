package services

import ws "socialNetwork/pkg/websocket"

// Services and routes use the same active WebSocket registry.
type Message = ws.Message

func SendToUser(userID int, message Message) bool { return ws.SendToUser(userID, message) }
