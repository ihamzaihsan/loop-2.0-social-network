package services

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// SafeConn is a thread-safe WebSocket connection wrapper
type SafeConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// Global client management
var clients = make(map[int]*SafeConn)
var clientsMutex sync.RWMutex

// WriteJSON sends a JSON message through the WebSocket connection
func (sc *SafeConn) WriteJSON(v interface{}) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteJSON(v)
}

// SendToUser sends a message to a specific user if they are connected
func SendToUser(userID int, message Message) bool {
	clientsMutex.RLock()
	client, exists := clients[userID]
	clientsMutex.RUnlock()

	if !exists {
		return false
	}

	if err := client.WriteJSON(message); err != nil {
		log.Printf("[ERROR] Failed to send message to user %d: %v", userID, err)
		return false
	}
	return true
}
