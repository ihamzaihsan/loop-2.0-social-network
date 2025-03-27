package services

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/routes"
	"time"
)

// InviteToGroupService handles inviting users to a group
func InviteToGroupService(groupID, inviterID int, userIDs []int) error {
	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(groupID, inviterID)
	if err != nil || !isMember {
		return err
	}

	// Get group info
	group, err := query.GetGroupByID(groupID, inviterID)
	if err != nil {
		return err
	}

	// Invite each user
	for _, inviteeID := range userIDs {
		err := query.InviteUserToGroup(groupID, inviterID, inviteeID)
		if err != nil {
			log.Printf("[ERROR] Failed to invite user %d to group %d: %v", inviteeID, groupID, err)
			continue
		}

		// Create notification
		_, err = db.DBInstance.DB.Exec(
			"INSERT INTO notifications (to_user_id, from_user_id, content, type, read, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			inviteeID, inviterID, "You have been invited to join the group: "+group.Title, "group_invitation", false, time.Now(),
		)
		if err != nil {
			log.Printf("[ERROR] Failed to create notification: %v", err)
		}

		// Send WebSocket notification if user is online
		routes.SendToUser(inviteeID, routes.Message{
			Type: "group_invitation",
			Content: map[string]interface{}{
				"group_id":    groupID,
				"group_title": group.Title,
				"inviter_id":  inviterID,
			},
		})
	}

	return nil
}
