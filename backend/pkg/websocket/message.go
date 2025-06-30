package websocket

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"sync"
	"time"

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

// BroadcastToGroupMembers broadcasts a message to all members of a group except the sender
func BroadcastToGroupMembers(groupID, senderID int, message Message) {
	// Get all members of the group
	rows, err := db.DBInstance.DB.Query(`
        SELECT user_id FROM group_members
        WHERE group_id = ?
    `, groupID)

	if err != nil {
		log.Printf("[ERROR] Failed to get group members for broadcast: %v", err)
		return
	}
	defer rows.Close()

	var memberIDs []int
	for rows.Next() {
		var memberID int
		if err := rows.Scan(&memberID); err != nil {
			log.Printf("[ERROR] Failed to scan member ID: %v", err)
			continue
		}
		memberIDs = append(memberIDs, memberID)
	}

	log.Printf("[INFO] Broadcasting %s message to %d group members (excluding sender %d)", message.Type, len(memberIDs), senderID)

	// Send the message to all online members except the sender
	successCount := 0
	for _, memberID := range memberIDs {
		if memberID != senderID || senderID == 0 {
			if SendToUser(memberID, message) {
				successCount++
			}
		}
	}

	log.Printf("[INFO] Successfully broadcast %s message to %d/%d group members", message.Type, successCount, len(memberIDs))
}

// BroadcastGroupPost creates and broadcasts a complete group post to all group members
func BroadcastGroupPost(groupID, userID int, postID int64, content, image string) {
	// Get user info for the post
	var firstName, lastName string
	var avatar sql.NullString
	err := db.DBInstance.DB.QueryRow(`
        SELECT first_name, last_name, avatar
        FROM users
        WHERE id = ?
    `, userID).Scan(&firstName, &lastName, &avatar)

	if err != nil {
		log.Printf("[ERROR] Failed to get user info for group post broadcast: %v", err)
		return
	}

	// Create complete post object for broadcasting
	post := map[string]interface{}{
		"id":            postID,
		"group_id":      groupID,
		"user_id":       userID,
		"content":       content,
		"image":         image,
		"created_at":    time.Now(),
		"first_name":    firstName,
		"last_name":     lastName,
		"comment_count": 0,
	}

	if avatar.Valid {
		post["avatar"] = avatar.String
	}

	// Broadcast to all group members
	BroadcastToGroupMembers(groupID, userID, Message{
		Type:    "group_post",
		Content: post,
	})
}
