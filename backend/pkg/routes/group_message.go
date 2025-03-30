package routes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
	"strconv"
	"time"
)

// GroupMessageResponse represents a message in a group
type GroupMessageResponse struct {
	ID        int       `json:"id"`
	SenderID  int       `json:"sender_id"`
	GroupID   int       `json:"group_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Sender    struct {
		ID        int     `json:"id"`
		FirstName string  `json:"firstName"`
		LastName  string  `json:"lastName"`
		Avatar    *string `json:"avatar,omitempty"`
	} `json:"sender"`
}

// ServeGroupMessages handles fetching messages for a group
func ServeGroupMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get group ID from query parameters
	groupIDStr := r.URL.Query().Get("id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get current user ID
	userID, err := auth.GetUserID(r)
	if err != nil || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is a member of the group
	var isMember bool
	err = db.DBInstance.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)
	`, groupID, userID).Scan(&isMember)

	if err != nil {
		log.Printf("[ERROR] Failed to check group membership: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !isMember {
		http.Error(w, "You are not a member of this group", http.StatusForbidden)
		return
	}

	// Get messages for this group
	rows, err := db.DBInstance.DB.Query(`
		SELECT m.id, m.user_id, m.group_id, m.content, m.created_at,
			u.id, u.first_name, u.last_name, u.avatar
		FROM group_chat_messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.group_id = ?
		ORDER BY m.created_at ASC
	`, groupID)

	if err != nil {
		log.Printf("[ERROR] Database error fetching group messages: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []GroupMessageResponse

	for rows.Next() {
		var msg GroupMessageResponse
		var avatar sql.NullString
		err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.GroupID, &msg.Content, &msg.CreatedAt,
			&msg.Sender.ID, &msg.Sender.FirstName, &msg.Sender.LastName, &avatar,
		)
		if err != nil {
			log.Printf("[ERROR] Error scanning group message row: %v", err)
			continue
		}

		if avatar.Valid {
			avatarStr := avatar.String
			msg.Sender.Avatar = &avatarStr
		}

		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}

// SendGroupMessage handles sending a message to a group
func SendGroupMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user ID
	userID, err := auth.GetUserID(r)
	if err != nil || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		GroupID int    `json:"group_id"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Content == "" {
		http.Error(w, "Message content cannot be empty", http.StatusBadRequest)
		return
	}

	// Check if user is a member of the group
	var isMember bool
	err = db.DBInstance.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)
	`, request.GroupID, userID).Scan(&isMember)

	if err != nil {
		log.Printf("[ERROR] Failed to check group membership: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !isMember {
		http.Error(w, "You are not a member of this group", http.StatusForbidden)
		return
	}

	// Insert the message
	result, err := db.DBInstance.DB.Exec(`
		INSERT INTO group_chat_messages (group_id, user_id, content, created_at)
		VALUES (?, ?, ?, ?)
	`, request.GroupID, userID, request.Content, time.Now())

	if err != nil {
		log.Printf("[ERROR] Failed to store group message: %v", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	messageID, _ := result.LastInsertId()

	// Get sender info for the response
	var senderFirstName, senderLastName string
	var senderAvatar sql.NullString
	err = db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name, avatar
		FROM users
		WHERE id = ?
	`, userID).Scan(&senderFirstName, &senderLastName, &senderAvatar)

	if err != nil {
		log.Printf("[ERROR] Failed to get sender info: %v", err)
	}

	// Create response
	message := GroupMessageResponse{
		ID:        int(messageID),
		SenderID:  userID,
		GroupID:   request.GroupID,
		Content:   request.Content,
		CreatedAt: time.Now(),
		Sender: struct {
			ID        int     `json:"id"`
			FirstName string  `json:"firstName"`
			LastName  string  `json:"lastName"`
			Avatar    *string `json:"avatar,omitempty"`
		}{
			ID:        userID,
			FirstName: senderFirstName,
			LastName:  senderLastName,
		},
	}

	if senderAvatar.Valid {
		avatarStr := senderAvatar.String
		message.Sender.Avatar = &avatarStr
	}

	// Broadcast the message to other group members
	broadcastToGroupMembers(request.GroupID, userID, message)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": message,
	})
}

// broadcastToGroupMembers sends a message to all online group members except the sender
func broadcastToGroupMembers(groupID, senderID int, message interface{}) {
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

    // Determine the message type based on the type of the message parameter
    var msgType string
    switch message.(type) {
    case GroupMessageResponse:
        msgType = "group_message"
    case Message:
        // For Message type, the Type field is already set
        msg, ok := message.(Message)
        if ok {
            // Send the message to all online members except the sender
            for _, memberID := range memberIDs {
                if memberID != senderID || senderID == 0 {
                    SendToUser(memberID, msg)
                }
            }
            return
        }
    default:
        msgType = "group_update" // Default type
    }

    // Send the message to all online members except the sender
    for _, memberID := range memberIDs {
        if memberID != senderID || senderID == 0 {
            SendToUser(memberID, Message{
                Type:    msgType,
                Content: message,
            })
        }
    }
}
