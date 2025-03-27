package services

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/routes"
	"time"
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

	// Create notification for group creator
	_, err = db.DBInstance.DB.Exec(
		"INSERT INTO notifications (to_user_id, from_user_id, content, type, read, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		group.CreatorID, userID, "A user has requested to join your group: "+group.Title, "group_join_request", false, time.Now(),
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create notification: %v", err)
	}

	// Send WebSocket notification to creator if online
	routes.SendToUser(group.CreatorID, routes.Message{
		Type: "group_join_request",
		Content: map[string]interface{}{
			"group_id":    groupID,
			"group_title": group.Title,
			"user_id":     userID,
		},
	})

	return nil
}
