package websocket

import (
	"log"
	"sync"

	gorilla "github.com/gorilla/websocket"
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// SafeConn is a thread-safe WebSocket connection wrapper
type SafeConn struct {
	conn *gorilla.Conn
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

// BroadcastMessage sends a message to all connected clients
func BroadcastMessage(message Message) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	for userID, client := range clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("[ERROR] Failed to broadcast message to user %d: %v", userID, err)
		}
	}
}

// RegisterClient adds a client to the global client map
func RegisterClient(userID int, conn *gorilla.Conn) {
	safeConn := &SafeConn{
		conn: conn,
	}

	clientsMutex.Lock()
	clients[userID] = safeConn
	clientsMutex.Unlock()
}

// UnregisterClient removes a client from the global client map
func UnregisterClient(userID int) {
	clientsMutex.Lock()
	delete(clients, userID)
	clientsMutex.Unlock()
}

// GetClient retrieves a client's connection
func GetClient(userID int) (*SafeConn, bool) {
	clientsMutex.RLock()
	client, exists := clients[userID]
	clientsMutex.RUnlock()
	return client, exists
}
