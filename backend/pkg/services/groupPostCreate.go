package services

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/routes"
)

// CreateGroupPostService handles creating a post in a group
func CreateGroupPostService(groupID, userID int, content, image string) (int, error) {
	// Check if user is an active member of the group
	isMember, err := query.IsGroupMember(groupID, userID)
	if err != nil || !isMember {
		return 0, err
	}

	// Create post
	postID, err := query.CreateGroupPost(groupID, userID, content, image)
	if err != nil {
		log.Printf("[ERROR] Failed to create group post: %v", err)
		return 0, err
	}

	// Get group members for notification
	members, err := query.GetGroupMembers(groupID)
	if err == nil {
		// Get user info for notification
		var firstName, lastName string
		err = db.DBInstance.DB.QueryRow(
			"SELECT first_name, last_name FROM users WHERE id = ?", userID,
		).Scan(&firstName, &lastName)
		
		if err == nil {
			// Notify group members
			for _, member := range members {
				if member.UserID != userID {
					// Send WebSocket notification if member is online
					routes.SendToUser(member.UserID, routes.Message{
						Type: "group_post",
						Content: map[string]interface{}{
							"group_id":       groupID,
							"post_id":        postID,
							"user_id":        userID,
							"user_name":      firstName + " " + lastName,
							"content_preview": truncateString(content, 50),
						},
					})
				}
			}
		}
	}

	return postID, nil
}

// Helper function to truncate string
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}
