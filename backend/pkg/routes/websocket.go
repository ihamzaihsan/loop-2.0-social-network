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

			handlePrivateMessage(userID, contentMap)
		case "group_message":
			log.Printf("[INFO] Handling group message from user %d", userID)
			contentMap, ok := msg.Content.(map[string]interface{})
			if !ok {
				log.Printf("[ERROR] Invalid message content format")
				continue
			}
			handleGroupMessage(userID, contentMap)
		
		case "group_post":
			log.Printf("[INFO] Handling group post from user %d", userID)
			contentMap, ok := msg.Content.(map[string]interface{})
			if !ok {
				log.Printf("[ERROR] Invalid message content format")
				continue
			}
			handleGroupPost(userID, contentMap)
			
		case "group_event":
			log.Printf("[INFO] Handling group event from user %d", userID)
			contentMap, ok := msg.Content.(map[string]interface{})
			if !ok {
				log.Printf("[ERROR] Invalid message content format")
				continue
			}
			handleGroupEvent(userID, contentMap)
		case "event_response":
			log.Printf("[INFO] Handling event response from user %d", userID)
			contentMap, ok := msg.Content.(map[string]interface{})
			if !ok {
				log.Printf("[ERROR] Invalid message content format")
				continue
			}
			handleEventResponse(userID, contentMap)	

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

// Add these functions to your existing websocket.go file

// Handle group message
func handleGroupMessage(userID int, content map[string]interface{}) {
	// Extract group_id
	groupIDFloat, ok := content["group_id"].(float64)
	if !ok {
		log.Printf("[ERROR] Invalid group_id format")
		return
	}
	groupID := int(groupIDFloat)

	// Extract message content
	messageContent, ok := content["content"].(string)
	if !ok {
		log.Printf("[ERROR] Invalid content format")
		return
	}

	log.Printf("[INFO] Processing group message from user %d to group %d: %s", userID, groupID, messageContent)

	// Check if user is a member of the group
	var isMember bool
	err := db.DBInstance.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)
	`, groupID, userID).Scan(&isMember)

	if err != nil {
		log.Printf("[ERROR] Failed to check group membership: %v", err)
		return
	}

	if !isMember {
		log.Printf("[ERROR] User %d is not a member of group %d", userID, groupID)
		return
	}

	// Insert the message
	result, err := db.DBInstance.DB.Exec(
		"INSERT INTO group_chat_messages (group_id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
		groupID, userID, messageContent, time.Now(),
	)
	if err != nil {
		log.Printf("[ERROR] Failed to store group message: %v", err)
		return
	}

	messageID, _ := result.LastInsertId()
	log.Printf("[INFO] Stored group message with ID %d", messageID)

	// Get sender info
	var senderFirstName, senderLastName string
	var senderAvatar sql.NullString
	err = db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name, avatar
		FROM users
		WHERE id = ?
	`, userID).Scan(&senderFirstName, &senderLastName, &senderAvatar)

	if err != nil {
		log.Printf("[ERROR] Failed to get sender info: %v", err)
		return
	}

	// Create message object
	message := GroupMessageResponse{
		ID:        int(messageID),
		SenderID:  userID,
		GroupID:   groupID,
		Content:   messageContent,
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

	// Broadcast to group members
	broadcastToGroupMembers(groupID, userID, message)
}

func handleGroupPost(userID int, content map[string]interface{}) {
    // Extract group_id and post content
    groupIDFloat, ok := content["group_id"].(float64)
    if !ok {
        log.Printf("[ERROR] Invalid group_id format")
        return
    }
    groupID := int(groupIDFloat)
    
    postContent, ok := content["content"].(string)
    if !ok {
        log.Printf("[ERROR] Invalid post content format")
        return
    }
    
    // Optional image
    var image string
    if imgContent, ok := content["image"].(string); ok {
        image = imgContent
    }
    
    log.Printf("[INFO] Processing group post from user %d to group %d", userID, groupID)
    
    // Check if user is a member of the group
    var isMember bool
    err := db.DBInstance.DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM group_members 
            WHERE group_id = ? AND user_id = ?
        )
    `, groupID, userID).Scan(&isMember)
    
    if err != nil {
        log.Printf("[ERROR] Failed to check group membership: %v", err)
        return
    }
    
    if !isMember {
        log.Printf("[ERROR] User %d is not a member of group %d", userID, groupID)
        return
    }
    
    // Insert the post
    result, err := db.DBInstance.DB.Exec(`
        INSERT INTO group_posts (group_id, user_id, content, image, created_at)
        VALUES (?, ?, ?, ?, ?)
    `, groupID, userID, postContent, image, time.Now())
    
    if err != nil {
        log.Printf("[ERROR] Failed to store group post: %v", err)
        return
    }
    
    postID, _ := result.LastInsertId()
    log.Printf("[INFO] Stored group post with ID %d", postID)
    
    // Get user info for the post
    var firstName, lastName string
    var avatar sql.NullString
    err = db.DBInstance.DB.QueryRow(`
        SELECT first_name, last_name, avatar
        FROM users
        WHERE id = ?
    `, userID).Scan(&firstName, &lastName, &avatar)
    
    if err != nil {
        log.Printf("[ERROR] Failed to get user info: %v", err)
        return
    }
    
    // Create post object for broadcasting
    post := map[string]interface{}{
        "id":           postID,
        "group_id":     groupID,
        "user_id":      userID,
        "content":      postContent,
        "image":        image,
        "created_at":   time.Now(),
        "first_name":   firstName,
        "last_name":    lastName,
        "comment_count": 0,
    }
    
    if avatar.Valid {
        post["avatar"] = avatar.String
    }
    
    // Broadcast to all group members
    broadcastToGroupMembers(groupID, userID, Message{
        Type:    "group_post",
        Content: post,
    })
}

func handleGroupEvent(userID int, content map[string]interface{}) {
    // Extract event details
    groupIDFloat, ok := content["group_id"].(float64)
    if !ok {
        log.Printf("[ERROR] Invalid group_id format")
        return
    }
    groupID := int(groupIDFloat)
    
    title, ok := content["title"].(string)
    if !ok {
        log.Printf("[ERROR] Invalid event title format")
        return
    }
    
    description, ok := content["description"].(string)
    if !ok {
        log.Printf("[ERROR] Invalid event description format")
        return
    }
    
    eventTimeStr, ok := content["event_time"].(string)
    if !ok {
        log.Printf("[ERROR] Invalid event time format")
        return
    }
    
    eventTime, err := time.Parse(time.RFC3339, eventTimeStr)
    if err != nil {
        log.Printf("[ERROR] Failed to parse event time: %v", err)
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
        return
    }
    
    if !isMember {
        log.Printf("[ERROR] User %d is not a member of group %d", userID, groupID)
        return
    }
    
    // Insert the event
    tx, err := db.DBInstance.DB.Begin()
    if err != nil {
        log.Printf("[ERROR] Failed to begin transaction: %v", err)
        return
    }
    
    result, err := tx.Exec(`
    INSERT INTO group_events (group_id, title, description, event_time, created_at)
    VALUES (?, ?, ?, ?, ?)
	`, groupID, title, description, eventTime, time.Now())

    
    if err != nil {
        tx.Rollback()
        log.Printf("[ERROR] Failed to store group event: %v", err)
        return
    }
    
    eventID, _ := result.LastInsertId()
    
    // Add default response options
    optionsInterface, ok := content["options"].([]interface{})
    if !ok || len(optionsInterface) == 0 {
        // Default options if none provided
        optionsInterface = []interface{}{"Going", "Not Going"}
    }
    
    for _, optionInterface := range optionsInterface {
        optionText, ok := optionInterface.(string)
        if !ok {
            continue
        }
        
        _, err := tx.Exec(`
            INSERT INTO event_response_options (event_id, option_text)
            VALUES (?, ?)
        `, eventID, optionText)
        
        if err != nil {
            tx.Rollback()
            log.Printf("[ERROR] Failed to store event option: %v", err)
            return
        }
    }
    
    if err := tx.Commit(); err != nil {
        log.Printf("[ERROR] Failed to commit transaction: %v", err)
        return
    }
    
    log.Printf("[INFO] Stored group event with ID %d", eventID)
    
    // Get response options for the event
    rows, err := db.DBInstance.DB.Query(`
        SELECT id, option_text
        FROM event_response_options
        WHERE event_id = ?
    `, eventID)
    
    if err != nil {
        log.Printf("[ERROR] Failed to get event options: %v", err)
        return
    }
    
    var responseOptions []map[string]interface{}
    for rows.Next() {
        var id int
        var optionText string
        if err := rows.Scan(&id, &optionText); err != nil {
            log.Printf("[ERROR] Failed to scan event option: %v", err)
            continue
        }
        
        responseOptions = append(responseOptions, map[string]interface{}{
            "id":          id,
            "event_id":    eventID,
            "option_text": optionText,
        })
    }
    rows.Close()
    
    // Create event object for broadcasting
    event := map[string]interface{}{
        "id":              eventID,
        "group_id":        groupID,
        "creator_id":      userID,
        "title":           title,
        "description":     description,
        "event_time":      eventTime,
        "created_at":      time.Now(),
        "going_count":     0,
        "not_going_count": 0,
        "response_options": responseOptions,
    }
    
    // Broadcast to all group members
    broadcastToGroupMembers(groupID, userID, Message{
        Type:    "group_event",
        Content: event,
    })
}

func handleEventResponse(userID int, content map[string]interface{}) {
    // Extract event_id and option_id
    eventIDFloat, ok := content["event_id"].(float64)
    if !ok {
        log.Printf("[ERROR] Invalid event_id format")
        return
    }
    eventID := int(eventIDFloat)
    
    optionIDFloat, ok := content["option_id"].(float64)
    if !ok {
        log.Printf("[ERROR] Invalid option_id format")
        return
    }
    optionID := int(optionIDFloat)
    
    // Get the group ID for this event
    var groupID int
    err := db.DBInstance.DB.QueryRow(`
        SELECT group_id FROM group_events WHERE id = ?
    `, eventID).Scan(&groupID)
    
    if err != nil {
        log.Printf("[ERROR] Failed to get group ID for event: %v", err)
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
        return
    }
    
    if !isMember {
        log.Printf("[ERROR] User %d is not a member of group %d", userID, groupID)
        return
    }
    
    // Record the response
    err = query.RespondToEvent(eventID, userID, optionID)
    if err != nil {
        log.Printf("[ERROR] Failed to record event response: %v", err)
        return
    }
    
    // Get the option text
    var optionText string
    err = db.DBInstance.DB.QueryRow(`
        SELECT option_text FROM event_response_options WHERE id = ?
    `, optionID).Scan(&optionText)
    
    if err != nil {
        log.Printf("[ERROR] Failed to get option text: %v", err)
        return
    }
    
    // Count responses for this event
    var goingCount, notGoingCount int
    err = db.DBInstance.DB.QueryRow(`
        SELECT 
            COUNT(CASE WHEN ero.option_text = 'Going' THEN 1 END) as going_count,
            COUNT(CASE WHEN ero.option_text = 'Not Going' THEN 1 END) as not_going_count
        FROM event_responses er
        JOIN event_response_options ero ON er.response_option_id = ero.id
        WHERE er.event_id = ?
    `, eventID).Scan(&goingCount, &notGoingCount)
    
    if err != nil {
        log.Printf("[ERROR] Failed to count event responses: %v", err)
        return
    }
    
    // Create response object for broadcasting
    responseData := map[string]interface{}{
        "event_id":        eventID,
        "user_id":         userID,
        "option_id":       optionID,
        "response":        optionText,
        "going_count":     goingCount,
        "not_going_count": notGoingCount,
        "current_user_id": userID,
    }
    
    // Broadcast to all group members
    broadcastToGroupMembers(groupID, 0, Message{
        Type:    "event_response",
        Content: responseData,
    })
}
