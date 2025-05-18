package services

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	ws "socialNetwork/pkg/websocket"
)

// RequestToJoinGroupService handles user requests to join a group
func RequestToJoinGroupService(groupID, userID int) error {
	// Get group info
	group, err := query.GetGroupByID(groupID, userID)
	if err != nil {
		return err
	}

	// Create join request
	err = query.RequestToJoinGroup(groupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to create join request: %v", err)
		return err
	}

	// Get requester's name
	var firstName, lastName string
	err = db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name FROM users WHERE id = ?
	`, userID).Scan(&firstName, &lastName)
	if err != nil {
		log.Printf("[ERROR] Failed to get user name: %v", err)
		firstName = "A user"
		lastName = ""
	}
	userName := firstName + " " + lastName

	// Create notification content
	notificationContent := userName + " has requested to join your group: " + group.Title

	// Create notification for group creator
	notificationID, err := CreateNotification(
		group.CreatorID,
		userID,
		"group_join_request",
		groupID,
		notificationContent,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create notification: %v", err)
	} else {
		log.Printf("[INFO] Created group join request notification: %d", notificationID)
	}

	// Send WebSocket notification to creator if online
	ws.SendToUser(group.CreatorID, ws.Message{
		Type: "notification",
		Content: map[string]interface{}{
			"id":          notificationID,
			"type":        "group_join_request",
			"group_id":    groupID,
			"group_title": group.Title,
			"user_id":     userID,
			"user_name":   userName,
			"content":     notificationContent,
			"actions":     []string{"accept", "reject"},
		},
	})

	return nil
}
