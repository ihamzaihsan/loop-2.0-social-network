package routes

import (
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

	userId, err := strconv.Atoi(pathParts[len(pathParts)-1])
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

	// Get private messages between users
	rows, err := db.DBInstance.DB.Query(`
        SELECT m.id, m.sender_id, m.receiver_id, m.content, m.created_at, m.is_read,
               u.id, u.first_name, u.last_name, u.avatar
        FROM messages m
        JOIN users u ON m.sender_id = u.id
        WHERE (m.sender_id = ? AND m.receiver_id = ? AND m.group_id IS NULL) 
        OR (m.sender_id = ? AND m.receiver_id = ? AND m.group_id IS NULL)
        ORDER BY m.created_at DESC
        LIMIT ? OFFSET ?
    `, currentUserID, userId, userId, currentUserID, limit, offset)

	if err != nil {
		log.Printf("[ERROR] Database error fetching messages: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []MessageResponse

	for rows.Next() {
		var msg MessageResponse
		err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.CreatedAt, &msg.IsRead,
			&msg.Sender.ID, &msg.Sender.FirstName, &msg.Sender.LastName, &msg.Sender.Avatar,
		)
		if err != nil {
			log.Printf("[ERROR] Error scanning message row: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// Mark messages as read
	_, err = db.DBInstance.DB.Exec(`
		UPDATE messages 
		SET is_read = true 
		WHERE sender_id = ? AND receiver_id = ? AND is_read = false
	`, userId, currentUserID)

	if err != nil {
		log.Printf("[ERROR] Failed to mark messages as read: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}

// ServeGroupMessages handles fetching messages for a group
func ServeGroupMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	currentUserID, err := auth.GetUserID(r)
	if err != nil || currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the group
	var isMember bool
	err = db.DBInstance.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)
	`, groupID, currentUserID).Scan(&isMember)

	if err != nil {
		log.Printf("[ERROR] Error checking group membership: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !isMember {
		http.Error(w, "You are not a member of this group", http.StatusForbidden)
		return
	}

	// Get group messages
	rows, err := db.DBInstance.DB.Query(`
		SELECT m.id, m.sender_id, NULL as receiver_id, m.group_id, m.content, m.created_at, m.is_read,
			   u.id, u.first_name, u.last_name, u.avatar
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.group_id = ?
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, groupID, limit, offset)

	if err != nil {
		log.Printf("[ERROR] Database error fetching group messages: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []MessageResponse

	for rows.Next() {
		var msg MessageResponse
		err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.GroupID, &msg.Content, &msg.CreatedAt, &msg.IsRead,
			&msg.Sender.ID, &msg.Sender.FirstName, &msg.Sender.LastName, &msg.Sender.Avatar,
		)
		if err != nil {
			log.Printf("[ERROR] Error scanning group message row: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// Mark group messages as read for this user
	_, err = db.DBInstance.DB.Exec(`
		INSERT OR IGNORE INTO message_read_status (message_id, user_id, is_read)
		SELECT id, ?, true FROM messages WHERE group_id = ? AND sender_id != ?
	`, currentUserID, groupID, currentUserID)

	if err != nil {
		log.Printf("[ERROR] Failed to mark group messages as read: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}

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
		// Private message
		result, err := db.DBInstance.DB.Exec(
			"INSERT INTO messages (sender_id, receiver_id, content, created_at, is_read) VALUES (?, ?, ?, ?, ?)",
			currentUserID, *request.ReceiverID, request.Content, time.Now(), false,
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
					"id":         messageID,
					"sender_id":  currentUserID,
					"sender":     sender.FirstName + " " + sender.LastName,
					"content":    request.Content,
					"created_at": time.Now(),
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
