package services

import (
	"fmt"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
	"time"
)

// MessageService handles message-related operations
type MessageService struct{}

// SaveAndBroadcastGroupMessage handles saving and broadcasting a group message
func (s *MessageService) SaveAndBroadcastGroupMessage(groupID, userID int, content string) (int64, error) {
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
		return 0, err
	}

	if !isMember {
		log.Printf("[ERROR] User %d is not a member of group %d", userID, groupID)
		return 0, fmt.Errorf("user is not a member of this group")
	}

	// Insert the message
	result, err := db.DBInstance.DB.Exec(`
		INSERT INTO group_chat_messages (group_id, user_id, content, created_at)
		VALUES (?, ?, ?, ?)
	`, groupID, userID, content, time.Now())

	if err != nil {
		log.Printf("[ERROR] Failed to store group message: %v", err)
		return 0, err
	}

	messageID, _ := result.LastInsertId()
	log.Printf("[INFO] Stored group message with ID %d", messageID)

	// Get sender info
	sender, err := query.GetUserInfo(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get sender info: %v", err)
		return messageID, err
	}

	// Create message object
	message := models.GroupMessage{
		ID:        int(messageID),
		SenderID:  userID,
		GroupID:   groupID,
		Content:   content,
		CreatedAt: time.Now(),
		Sender: struct {
			ID        int    `json:"id"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Avatar    string `json:"avatar,omitempty"`
		}{
			ID:        userID,
			FirstName: sender.FirstName,
			LastName:  sender.LastName,
			Avatar:    getAvatarString(sender.Avatar),
		},
	}

	// Broadcast to all group members
	broadcastGroupMessage(groupID, userID, message)

	return messageID, nil
}

// Helper function to safely handle nullable string pointers
func getAvatarString(avatar *string) string {
	if avatar == nil {
		return ""
	}
	return *avatar
}

// Add this function to your message_service.go file
func broadcastGroupMessage(groupID, userID int, message interface{}) {
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

	// Send the message to all online members except the sender
	for _, memberID := range memberIDs {
		if memberID != userID {
			// Use the WebSocket service to send the message
			// This assumes you have a SendToUser function in the services package
			SendToUser(memberID, Message{
				Type:    "group_message",
				Content: message,
			})
		}
	}
}
