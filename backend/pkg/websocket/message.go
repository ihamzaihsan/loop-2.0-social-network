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

// Keep every connection for a user so one tab cannot replace another.
var clients = make(map[int]map[*gorilla.Conn]*SafeConn)
var clientsMutex sync.RWMutex

func (sc *SafeConn) WriteJSON(v interface{}) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return sc.conn.WriteJSON(v)
}

func SendToUser(userID int, message Message) bool {
	clientsMutex.RLock()
	connections := make([]*SafeConn, 0, len(clients[userID]))
	for _, client := range clients[userID] {
		connections = append(connections, client)
	}
	clientsMutex.RUnlock()
	sent := false
	for _, client := range connections {
		if err := client.WriteJSON(message); err == nil {
			sent = true
		}
	}
	return sent
}

func BroadcastMessage(message Message) {
	clientsMutex.RLock()
	ids := make([]int, 0, len(clients))
	for id := range clients {
		ids = append(ids, id)
	}
	clientsMutex.RUnlock()
	for _, id := range ids {
		SendToUser(id, message)
	}
}

func RegisterClient(userID int, conn *gorilla.Conn) {
	clientsMutex.Lock()
	wasOffline := len(clients[userID]) == 0
	if clients[userID] == nil {
		clients[userID] = make(map[*gorilla.Conn]*SafeConn)
	}
	clients[userID][conn] = &SafeConn{conn: conn}
	ids := make([]int, 0, len(clients))
	for id := range clients {
		ids = append(ids, id)
	}
	client := clients[userID][conn]
	clientsMutex.Unlock()
	client.WriteJSON(Message{Type: "presence_snapshot", Content: map[string]interface{}{"user_ids": ids, "user_id": userID}})
	if wasOffline {
		BroadcastMessage(Message{Type: "presence_changed", Content: map[string]interface{}{"user_id": userID, "online": true}})
	}
}

func UnregisterClient(userID int, conn *gorilla.Conn) {
	clientsMutex.Lock()
	delete(clients[userID], conn)
	isOffline := len(clients[userID]) == 0
	if isOffline {
		delete(clients, userID)
	}
	clientsMutex.Unlock()
	if isOffline {
		BroadcastMessage(Message{Type: "presence_changed", Content: map[string]interface{}{"user_id": userID, "online": false}})
	}
}

func PublishChange(resource string) {
	BroadcastMessage(Message{Type: "resource_changed", Content: map[string]string{"resource": resource}})
}

func DisconnectUser(userID int) {
	clientsMutex.RLock()
	connections := make([]*gorilla.Conn, 0, len(clients[userID]))
	for conn := range clients[userID] {
		connections = append(connections, conn)
	}
	clientsMutex.RUnlock()
	for _, conn := range connections {
		conn.Close()
	}
}

// GetClient is retained for handlers that send a response to a connected user.
func GetClient(userID int) (*SafeConn, bool) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	for _, client := range clients[userID] {
		return client, true
	}
	return nil, false
}

// BroadcastToGroupMembers broadcasts a message to all members of a group except the sender
func BroadcastToGroupMembers(groupID, senderID int, message Message) {
	PublishChange("groups")
	// Get all members of the group
	rows, err := db.DBInstance.DB.Query(`
        SELECT user_id FROM group_members
        WHERE group_id = ? AND status = 'active'
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
