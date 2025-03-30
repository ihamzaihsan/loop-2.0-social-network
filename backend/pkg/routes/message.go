package routes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"strconv"
	"strings"
	"time"
)

type MessageResponse struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	GroupID    *int      `json:"group_id,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	IsRead     bool      `json:"is_read"`
	Sender     struct {
		ID        int     `json:"id"`
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		Avatar    *string `json:"avatar"`
	} `json:"sender"`
}

// ServeMessages handles fetching private messages between users
func ServeMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	otherUserID, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	currentUserID, err := auth.GetUserID(r)
	if err != nil || currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// First, find the chat between these two users
	var chatID int
	err = db.DBInstance.DB.QueryRow(`
		SELECT id FROM chats 
		WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)
	`, currentUserID, otherUserID, otherUserID, currentUserID).Scan(&chatID)
	
	if err == sql.ErrNoRows {
		// No chat exists yet, return empty messages
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"messages": []interface{}{},
		})
		return
	} else if err != nil {
		log.Printf("[ERROR] Failed to find chat: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get messages for this chat
	rows, err := db.DBInstance.DB.Query(`
		SELECT m.id, m.sender_id, m.content, m.created_at, 
			    u.id, u.first_name, u.last_name, u.avatar
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.chat_id = ?
		ORDER BY m.created_at ASC
		LIMIT ? OFFSET ?
	`, chatID, limit, offset)

	if err != nil {
		log.Printf("[ERROR] Database error fetching messages: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []MessageResponse

	for rows.Next() {
		var msg MessageResponse
		var isRead bool = true // Default to true since we don't track read status
		
		err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.Content, &msg.CreatedAt,
			&msg.Sender.ID, &msg.Sender.FirstName, &msg.Sender.LastName, &msg.Sender.Avatar,
		)
		if err != nil {
			log.Printf("[ERROR] Error scanning message row: %v", err)
			continue
		}
		
		// Set receiver_id based on sender
		if msg.SenderID == currentUserID {
			msg.ReceiverID = otherUserID
		} else {
			msg.ReceiverID = currentUserID
		}
		
		msg.IsRead = isRead
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}

// ServeGroupMessages handles fetching messages for a group
// func ServeGroupMessages(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	pathParts := strings.Split(r.URL.Path, "/")
// 	if len(pathParts) < 3 {
// 		http.Error(w, "Invalid request", http.StatusBadRequest)
// 		return
// 	}

// 	groupID, err := strconv.Atoi(pathParts[len(pathParts)-1])
// 	if err != nil {
// 		http.Error(w, "Invalid group ID", http.StatusBadRequest)
// 		return
// 	}

// 	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
// 	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
// 	if limit == 0 {
// 		limit = 20
// 	}

// 	currentUserID, err := auth.GetUserID(r)
// 	if err != nil || currentUserID == 0 {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Check if user is a member of the group
// 	var isMember bool
// 	err = db.DBInstance.DB.QueryRow(`
// 		SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)
// 	`, groupID, currentUserID).Scan(&isMember)

// 	if err != nil {
// 		log.Printf("[ERROR] Error checking group membership: %v", err)
// 		http.Error(w, "Database error", http.StatusInternalServerError)
// 		return
// 	}

// 	if !isMember {
// 		http.Error(w, "You are not a member of this group", http.StatusForbidden)
// 		return
// 	}

// 	// Get group messages
// 	rows, err := db.DBInstance.DB.Query(`
// 		SELECT m.id, m.sender_id, NULL as receiver_id, m.group_id, m.content, m.created_at, m.is_read,
// 			    u.id, u.first_name, u.last_name, u.avatar
// 		FROM messages m
// 		JOIN users u ON m.sender_id = u.id
// 		WHERE m.group_id = ?
// 		ORDER BY m.created_at DESC
// 		LIMIT ? OFFSET ?
// 	`, groupID, limit, offset)

// 	if err != nil {
// 		log.Printf("[ERROR] Database error fetching group messages: %v", err)
// 		http.Error(w, "Database error", http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()

// 	var messages []MessageResponse

// 	for rows.Next() {
// 		var msg MessageResponse
// 		err := rows.Scan(
// 			&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.GroupID, &msg.Content, &msg.CreatedAt, &msg.IsRead,
// 			&msg.Sender.ID, &msg.Sender.FirstName, &msg.Sender.LastName, &msg.Sender.Avatar,
// 		)
// 		if err != nil {
// 			log.Printf("[ERROR] Error scanning group message row: %v", err)
// 			continue
// 		}
// 		messages = append(messages, msg)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"success":  true,
// 		"messages": messages,
// 	})
// }

// SendMessage handles sending a new message (both private and group)
func SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUserID, err := auth.GetUserID(r)
	if err != nil || currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		ReceiverID *int   `json:"receiver_id,omitempty"`
		GroupID    *int   `json:"group_id,omitempty"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Content == "" {
		http.Error(w, "Message content cannot be empty", http.StatusBadRequest)
		return
	}

	if request.ReceiverID == nil && request.GroupID == nil {
		http.Error(w, "Either receiver_id or group_id must be provided", http.StatusBadRequest)
		return
	}

	if request.ReceiverID != nil && request.GroupID != nil {
		http.Error(w, "Cannot specify both receiver_id and group_id", http.StatusBadRequest)
		return
	}

	var messageID int64

	// Get sender info for the response
	sender, err := query.GetUserInfo(currentUserID)
	if err != nil {
		log.Printf("[ERROR] Failed to get sender info: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if request.ReceiverID != nil {
		// Private message - first get or create a chat between the two users
		var chatID int

		// Check if a chat already exists between these users
		err = db.DBInstance.DB.QueryRow(`
			SELECT id FROM chats 
			WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)
		`, currentUserID, *request.ReceiverID, *request.ReceiverID, currentUserID).Scan(&chatID)

		if err == sql.ErrNoRows {
			// Create a new chat
			result, err := db.DBInstance.DB.Exec(`
				INSERT INTO chats (user1_id, user2_id, created_at) 
				VALUES (?, ?, ?)
			`, currentUserID, *request.ReceiverID, time.Now())

			if err != nil {
				log.Printf("[ERROR] Failed to create chat: %v", err)
				http.Error(w, "Failed to send message", http.StatusInternalServerError)
				return
			}

			chatIDInt64, err := result.LastInsertId()
			if err != nil {
				log.Printf("[ERROR] Failed to get last insert ID: %v", err)
				http.Error(w, "Failed to send message", http.StatusInternalServerError)
				return
			}

			chatID = int(chatIDInt64)
		} else if err != nil {
			log.Printf("[ERROR] Failed to check for existing chat: %v", err)
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		// Now insert the message with the chat_id
		result, err := db.DBInstance.DB.Exec(
			"INSERT INTO messages (chat_id, sender_id, content, created_at) VALUES (?, ?, ?, ?)",
			chatID, currentUserID, request.Content, time.Now(),
		)
		if err != nil {
			log.Printf("[ERROR] Failed to store private message: %v", err)
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		messageID, _ = result.LastInsertId()

		// Send WebSocket notification if receiver is online
		clientsMutex.RLock()
		if recipientConn, ok := clients[*request.ReceiverID]; ok {
			err = recipientConn.WriteJSON(Message{
				Type: "private_message",
				Content: map[string]interface{}{
					"id":          messageID,
					"sender_id":   currentUserID,
					"receiver_id": *request.ReceiverID,
					"sender":      sender.FirstName + " " + sender.LastName,
					"content":     request.Content,
					"created_at":  time.Now(),
				},
			})

			if err != nil {
				log.Printf("[ERROR] Failed to deliver private message to user %d: %v", *request.ReceiverID, err)
			}
		}
		clientsMutex.RUnlock()
	} else {
		// Group message
		// Check if user is a member of the group
		var isMember bool
		err = db.DBInstance.DB.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)
		`, *request.GroupID, currentUserID).Scan(&isMember)

		if err != nil {
			log.Printf("[ERROR] Error checking group membership: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if !isMember {
			http.Error(w, "You are not a member of this group", http.StatusForbidden)
			return
		}

		result, err := db.DBInstance.DB.Exec(
			"INSERT INTO messages (sender_id, group_id, content, created_at, is_read) VALUES (?, ?, ?, ?, ?)",
			currentUserID, *request.GroupID, request.Content, time.Now(), false,
		)
		if err != nil {
			log.Printf("[ERROR] Failed to store group message: %v", err)
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		messageID, _ = result.LastInsertId()

		// Get all group members
		rows, err := db.DBInstance.DB.Query("SELECT user_id FROM group_members WHERE group_id = ?", *request.GroupID)
		if err != nil {
			log.Printf("[ERROR] Failed to get group members: %v", err)
		} else {
			defer rows.Close()

			// Send WebSocket notification to all online group members
			var memberIDs []int
			for rows.Next() {
				var memberID int
				if err := rows.Scan(&memberID); err == nil && memberID != currentUserID {
					memberIDs = append(memberIDs, memberID)
				}
			}

			clientsMutex.RLock()
			for _, memberID := range memberIDs {
				if memberConn, ok := clients[memberID]; ok {
					err = memberConn.WriteJSON(Message{
						Type: "group_message",
						Content: map[string]interface{}{
							"id":         messageID,
							"group_id":   *request.GroupID,
							"sender_id":  currentUserID,
							"sender":     sender.FirstName + " " + sender.LastName,
							"content":    request.Content,
							"created_at": time.Now(),
						},
					})

					if err != nil {
						log.Printf("[ERROR] Failed to deliver group message to user %d: %v", memberID, err)
					}
				}
			}
			clientsMutex.RUnlock()
		}
	}

	// Return the sent message
	response := MessageResponse{
		ID:        int(messageID),
		SenderID:  currentUserID,
		Content:   request.Content,
		CreatedAt: time.Now(),
		IsRead:    false,
		Sender: struct {
			ID        int     `json:"id"`
			FirstName string  `json:"first_name"`
			LastName  string  `json:"last_name"`
			Avatar    *string `json:"avatar"`
		}{
			ID:        sender.ID,
			FirstName: sender.FirstName,
			LastName:  sender.LastName,
			Avatar:    sender.Avatar,
		},
	}

	if request.ReceiverID != nil {
		response.ReceiverID = *request.ReceiverID
	} else {
		response.GroupID = request.GroupID
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": response,
	})
}
