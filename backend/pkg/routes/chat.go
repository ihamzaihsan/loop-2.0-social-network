package routes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
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
		log.Printf("[ERROR] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current user ID from the session
	userID, err := auth.GetUserID(r)
	if err != nil || userID == 0 {
		log.Printf("[ERROR] Unauthorized access attempt: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}


			var count int
			err = db.DBInstance.DB.QueryRow("SELECT COUNT(*) FROM followers WHERE follower_id = ?", userID).Scan(&count)
			if err != nil {
				log.Printf("[ERROR] Failed to count followed users: %v", err)
			}
	

			// Query to get all users that the current user is following with complete information
				query := `
						SELECT 
							u.id,
							u.first_name,
							u.last_name,
							u.nickname,
							u.avatar
						FROM users u
						JOIN followers f ON u.id = f.following_id
						WHERE f.follower_id = ?
						ORDER BY u.first_name, u.last_name
					`



			rows, err := db.DBInstance.DB.Query(query, userID)
			if err != nil {
				log.Printf("[ERROR] Failed to query followed users: %v", err)
				http.Error(w, "Failed to get followed users", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var followedUsers []map[string]interface{}

			for rows.Next() {
				var id int
				var firstName, lastName string
				var nickname, avatar *string

				err := rows.Scan(&id, &firstName, &lastName, &nickname, &avatar)
				if err != nil {
					log.Printf("[ERROR] Failed to scan followed user row: %v", err)
					continue
				}

				user := map[string]interface{}{
					"id":        id,
					"firstName": firstName,
					"lastName":  lastName,
				}

				if nickname != nil {
					user["nickname"] = *nickname
				}

				if avatar != nil {
					user["avatar"] = *avatar
				}

				followedUsers = append(followedUsers, user)
			}

	

			
	
	// Return the followed users as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   followedUsers,
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
			0 AS unread_count  -- Default to 0 since we don't track read status
		FROM last_messages lm
		JOIN users u ON lm.contact_id = u.id
		ORDER BY lm.last_message_time DESC
	`

	rows, err := db.DBInstance.DB.Query(query, userID, userID, userID)
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
