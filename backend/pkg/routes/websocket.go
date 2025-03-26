package routes

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type SafeConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

var clients = make(map[int]*SafeConn)
var clientsMutex sync.RWMutex

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] WebSocket connection attempt from %s", r.RemoteAddr)

	// Get the user ID from the session
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("[ERROR] Failed to get user ID from session: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if userID == 0 {
		log.Printf("[ERROR] Invalid user ID in session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("[INFO] User ID %d authenticated for WebSocket connection", userID)

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR] WebSocket upgrade failed: %v", err)
		return
	}

	log.Printf("[INFO] WebSocket connection upgraded successfully for user %d", userID)

	// Get user info for the connected user
	user, err := query.GetUserInfo(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve user info for ID %d: %v", userID, err)
		conn.Close()
		return
	}

	log.Printf("[INFO] Retrieved user info for %s %s (ID: %d)", user.FirstName, user.LastName, userID)

	// Create a thread-safe connection wrapper
	safeConn := &SafeConn{
		conn: conn,
	}

	// Register the client
	clientsMutex.Lock()
	clients[userID] = safeConn
	clientsMutex.Unlock()

	log.Printf("[INFO] Registered client for user %d", userID)

	// Clean up when the connection closes
	defer func() {
		log.Printf("[INFO] Cleaning up WebSocket connection for user %d", userID)
		clientsMutex.Lock()
		delete(clients, userID)
		clientsMutex.Unlock()
		conn.Close()
		log.Printf("[INFO] WebSocket connection closed for user %d (%s %s)",
			userID, user.FirstName, user.LastName)
	}()

	log.Printf("[INFO] New WebSocket connection established for user %d (%s %s)",
		userID, user.FirstName, user.LastName)

	// Message handling loop
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("[INFO] WebSocket read error for user %d: %v", userID, err)
			break
		}

		log.Printf("[INFO] Received message of type '%s' from user %d", msg.Type, userID)

		switch msg.Type {
		case "private_message":
			log.Printf("[INFO] Handling private message from user %d", userID)

			// Extract message content
			contentMap, ok := msg.Content.(map[string]interface{})
			if !ok {
				log.Printf("[ERROR] Invalid message content format")
				continue
			}

			// Call the function to handle the private message
			handlePrivateMessage(userID, contentMap)
		case "typing_status":
			log.Printf("[INFO] Handling typing status from user %d", userID)
			// TODO: Implement typing status handling
		case "ping":
			log.Printf("[INFO] Handling ping from user %d", userID)
			err = safeConn.WriteJSON(Message{
				Type: "pong",
				Content: map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
				},
			})
			if err != nil {
				log.Printf("[ERROR] Failed to send pong to user %d (%s %s): %v",
					userID, user.FirstName, user.LastName, err)
			} else {
				log.Printf("[INFO] Sent pong to user %d", userID)
			}
		default:
			log.Printf("[WARN] Unknown message type from user %d (%s %s): %s",
				userID, user.FirstName, user.LastName, msg.Type)
		}
	}
}

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

// Add this function to handle private messages
func handlePrivateMessage(userID int, content map[string]interface{}) {
	// Extract receiver_id
	receiverIDFloat, ok := content["receiver_id"].(float64)
	if !ok {
		log.Printf("[ERROR] Invalid receiver_id format")
		return
	}
	receiverID := int(receiverIDFloat)

	// Extract message content
	messageContent, ok := content["content"].(string)
	if !ok {
		log.Printf("[ERROR] Invalid content format")
		return
	}

	log.Printf("[INFO] Processing message from user %d to user %d: %s", userID, receiverID, messageContent)

	// Find or create chat between the two users
	var chatID int
	err := db.DBInstance.DB.QueryRow(`
        SELECT id FROM chats 
        WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)
    `, userID, receiverID, receiverID, userID).Scan(&chatID)

	if err == sql.ErrNoRows {
		// Create a new chat
		result, err := db.DBInstance.DB.Exec(`
            INSERT INTO chats (user1_id, user2_id, created_at) 
            VALUES (?, ?, ?)
        `, userID, receiverID, time.Now())

		if err != nil {
			log.Printf("[ERROR] Failed to create chat: %v", err)
			return
		}

		chatIDInt64, err := result.LastInsertId()
		if err != nil {
			log.Printf("[ERROR] Failed to get last insert ID: %v", err)
			return
		}

		chatID = int(chatIDInt64)
		log.Printf("[INFO] Created new chat with ID %d between users %d and %d", chatID, userID, receiverID)
	} else if err != nil {
		log.Printf("[ERROR] Failed to check for existing chat: %v", err)
		return
	} else {
		log.Printf("[INFO] Found existing chat with ID %d between users %d and %d", chatID, userID, receiverID)
	}

	// Insert the message
	result, err := db.DBInstance.DB.Exec(
		"INSERT INTO messages (chat_id, sender_id, content, created_at) VALUES (?, ?, ?, ?)",
		chatID, userID, messageContent, time.Now(),
	)
	if err != nil {
		log.Printf("[ERROR] Failed to store message: %v", err)
		return
	}

	messageID, _ := result.LastInsertId()
	log.Printf("[INFO] Stored message with ID %d", messageID)

	// Get sender info for the small notification
	sender, err := query.GetUserInfo(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get sender info: %v", err)
		return
	}

	// Send notification to receiver if they're online
	clientsMutex.RLock()
	if recipientConn, ok := clients[receiverID]; ok {
		err = recipientConn.WriteJSON(Message{
			Type: "private_message",
			Content: map[string]interface{}{
				"id":          messageID,
				"sender_id":   userID,
				"receiver_id": receiverID,
				"sender":      sender.FirstName + " " + sender.LastName,
				"content":     messageContent,
				"created_at":  time.Now(),
			},
		})

		if err != nil {
			log.Printf("[ERROR] Failed to deliver message to user %d: %v", receiverID, err)
		} else {
			log.Printf("[INFO] Successfully delivered message to user %d", receiverID)
		}
	} else {
		log.Printf("[INFO] User %d is offline, message will be delivered when they connect", receiverID)
	}
	clientsMutex.RUnlock()
}
