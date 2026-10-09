package routes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"socialNetwork/pkg/access"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/utils"
	"strconv"
)

type ChatContact struct {
	ID              int     `json:"id"`
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	Nickname        *string `json:"nickname,omitempty"`
	Avatar          *string `json:"avatar,omitempty"`
	LastMessage     *string `json:"lastMessage,omitempty"`
	LastMessageTime *string `json:"lastMessageTime,omitempty"`
	UnreadCount     int     `json:"unreadCount"`
}

// ServeChatContacts handles the request to get all chat contacts for the current user
func ServeChatContacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current user ID from the session
	userID, err := auth.GetUserID(r)
	if err != nil || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get chat contacts for the user
	contacts, err := GetChatContacts(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get chat contacts: %v", err)
		http.Error(w, "Failed to get chat contacts", http.StatusInternalServerError)
		return
	}

	// Return the contacts as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"contacts": contacts,
	})
}

// ServeFollowedUsers handles the request to get all users that the current user is following
func ServeFollowedUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current user ID from the session
	userID, err := auth.GetUserID(r)
	if err != nil || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query the database directly to get following users with their details
	rows, err := db.DBInstance.DB.Query(`
		SELECT f.id, f.follower_id, f.following_id, f.status, 
			   u.first_name, u.last_name, u.nickname, u.avatar
		FROM followers f
		JOIN users u ON f.following_id = u.id
		WHERE f.follower_id = ? AND f.status = 'accept'
	`, userID)

	if err != nil {
		log.Printf("[ERROR] Failed to get following users: %v", err)
		http.Error(w, "Failed to get following users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var following []map[string]interface{}
	for rows.Next() {
		var id, followerID, followingID int
		var status, firstName, lastName string
		var nickname, avatar sql.NullString

		if err := rows.Scan(&id, &followerID, &followingID, &status, &firstName, &lastName, &nickname, &avatar); err != nil {
			log.Printf("[ERROR] Error scanning following row: %v", err)
			continue
		}

		// Create a user object with the retrieved data
		user := map[string]interface{}{
			"id":         id,
			"followerID": followerID,
			"followedID": followingID,
			"status":     status,
			"firstName":  firstName,
			"lastName":   lastName,
		}

		// Add nickname if available
		if nickname.Valid {
			user["nickname"] = nickname.String
		}

		// Add avatar if available
		if avatar.Valid {
			user["avatar"] = avatar.String
		}

		// Set display name (nickname or first+last name)
		if nickname.Valid {
			user["username"] = nickname.String
		} else {
			user["username"] = firstName + " " + lastName
		}

		following = append(following, user)
	}

	// Log the results for debugging
	log.Printf("[INFO] Found %d following users for user ID %d", len(following), userID)

	// Return the following users as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"following": following,
		"count":     len(following),
	})
}

func GetChatContacts(userID int) ([]ChatContact, error) {
	var contacts []ChatContact

	// Query to get all users that the current user has chatted with,
	// along with their last message
	query := `
		WITH last_messages AS (
			SELECT 
				c.id AS chat_id,
				CASE 
					WHEN c.user1_id = ? THEN c.user2_id
					ELSE c.user1_id
				END AS contact_id,
				m.content AS last_message,
				m.created_at AS last_message_time,
				MAX(m.id) AS last_message_id
			FROM chats c
			JOIN messages m ON c.id = m.chat_id
			WHERE c.user1_id = ? OR c.user2_id = ?
			GROUP BY c.id
		)
		SELECT 
			u.id,
			u.first_name,
			u.last_name,
			u.nickname,
			u.avatar,
			lm.last_message,
			lm.last_message_time,
			(SELECT COUNT(*) FROM messages unread WHERE unread.chat_id = lm.chat_id AND unread.sender_id = lm.contact_id AND unread.is_read = 0) AS unread_count
		FROM last_messages lm
		JOIN users u ON lm.contact_id = u.id
 WHERE u.is_suspended=0 AND NOT EXISTS(SELECT 1 FROM user_blocks b WHERE (b.blocker_id=? AND b.blocked_id=u.id) OR (b.blocker_id=u.id AND b.blocked_id=?))
		ORDER BY lm.last_message_time DESC
	`

	rows, err := db.DBInstance.DB.Query(query, userID, userID, userID, userID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to query chat contacts: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var contact ChatContact
		var lastMessage sql.NullString
		var lastMessageTime sql.NullString
		var nickname sql.NullString
		var avatar sql.NullString

		err := rows.Scan(
			&contact.ID,
			&contact.FirstName,
			&contact.LastName,
			&nickname,
			&avatar,
			&lastMessage,
			&lastMessageTime,
			&contact.UnreadCount,
		)

		if err != nil {
			log.Printf("[ERROR] Failed to scan chat contact row: %v", err)
			continue
		}

		if nickname.Valid {
			contact.Nickname = &nickname.String
		}

		if avatar.Valid {
			contact.Avatar = &avatar.String
		}

		if lastMessage.Valid {
			contact.LastMessage = &lastMessage.String
		}

		if lastMessageTime.Valid {
			formattedTime := lastMessageTime.String
			contact.LastMessageTime = &formattedTime
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

func UploadChatImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imagePath, err := utils.HandleImageUpload(file, header, viewer(r))
	if err != nil {
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true, "imageUrl": "` + imagePath + `"}`))
}

// GetUserInfo handles the request to get user information by ID
func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the user ID from query parameters
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get the current user ID from the session for authorization
	currentUserID, err := auth.GetUserID(r)
	if err != nil || currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if access.Blocked(currentUserID, userID) {
		http.Error(w, "User unavailable", 404)
		return
	}
	// Get user info from database
	var user struct {
		ID        int     `json:"id"`
		FirstName string  `json:"firstName"`
		LastName  string  `json:"lastName"`
		Nickname  *string `json:"nickname,omitempty"`
		Avatar    *string `json:"avatar,omitempty"`
	}

	var nickname, avatar sql.NullString
	err = db.DBInstance.DB.QueryRow(`
		SELECT id, first_name, last_name, nickname, avatar
		FROM users
		WHERE id = ? AND is_suspended=0
	`, userID).Scan(&user.ID, &user.FirstName, &user.LastName, &nickname, &avatar)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("[ERROR] Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	if nickname.Valid {
		user.Nickname = &nickname.String
	}

	if avatar.Valid {
		user.Avatar = &avatar.String
	}

	// Return the user info as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
	})
}
