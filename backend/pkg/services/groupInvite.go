package services

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	ws "socialNetwork/pkg/websocket"
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

	// Get inviter's name
	var inviterFirstName, inviterLastName string
	err = db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name FROM users WHERE id = ?
	`, inviterID).Scan(&inviterFirstName, &inviterLastName)
	if err != nil {
		log.Printf("[ERROR] Failed to get inviter name: %v", err)
	}
	inviterName := inviterFirstName + " " + inviterLastName

	// Track successful and failed invitations
	var successfulInvites []int
	var alreadyInvited []int
	var alreadyMembers []int

	// Invite each user
	for _, inviteeID := range userIDs {
		err := query.InviteUserToGroup(groupID, inviterID, inviteeID)
		if err != nil {
			switch err {
			case query.ErrUserAlreadyInvited:
				alreadyInvited = append(alreadyInvited, inviteeID)
				continue
			case query.ErrUserAlreadyMember:
				alreadyMembers = append(alreadyMembers, inviteeID)
				continue
			default:
				log.Printf("[ERROR] Failed to invite user %d to group %d: %v", inviteeID, groupID, err)
				continue
			}
		}

		successfulInvites = append(successfulInvites, inviteeID)

		// Create notification
		notificationContent := inviterName + " invited you to join the group: " + group.Title
		notificationID, err := CreateNotification(
			inviteeID,
			inviterID,
			"group_invitation",
			groupID,
			notificationContent,
		)
		if err != nil {
			log.Printf("[ERROR] Failed to create notification: %v", err)
		} else {
			log.Printf("[INFO] Created group invitation notification: %d", notificationID)
		}

		// Send WebSocket notification if user is online
		ws.SendToUser(inviteeID, ws.Message{
			Type: "notification",
			Content: map[string]interface{}{
				"id":           notificationID,
				"type":         "group_invitation",
				"group_id":     groupID,
				"group_title":  group.Title,
				"inviter_id":   inviterID,
				"inviter_name": inviterName,
				"content":      notificationContent,
				"actions":      []string{"accept", "reject"},
			},
		})
	}

	// If no successful invites and all users were either already invited or members,
	// return an error to inform the user
	if len(successfulInvites) == 0 && (len(alreadyInvited) > 0 || len(alreadyMembers) > 0) {
		if len(alreadyInvited) > 0 && len(alreadyMembers) > 0 {
			return sql.ErrNoRows // This will be handled by the route to show appropriate message
		} else if len(alreadyInvited) > 0 {
			return query.ErrUserAlreadyInvited
		} else {
			return query.ErrUserAlreadyMember
		}
	}

	return nil
}

// GetGroupMembershipStatus checks the current status of a user's membership in a group
func GetGroupMembershipStatus(groupID, userID int) (string, error) {
	var status string
	err := db.DBInstance.DB.QueryRow(`
        SELECT status FROM group_members 
        WHERE group_id = ? AND user_id = ?
    `, groupID, userID).Scan(&status)

	if err == sql.ErrNoRows {
		return "not_member", nil
	}

	if err != nil {
		log.Printf("[ERROR] Failed to get group membership status: %v", err)
		return "", err
	}

	return status, nil
}
