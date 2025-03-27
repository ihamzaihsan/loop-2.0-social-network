package services

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/routes"
	"time"
)

// HandleGroupMembershipRequestService processes a group invitation or join request
func HandleGroupMembershipRequestService(groupID, userID, actorID int, action, requestType string) error {
	// Validate action
	if action != "accept" && action != "reject" {
		return sql.ErrNoRows
	}

	// Validate request type
	if requestType != "invitation" && requestType != "request" {
		return sql.ErrNoRows
	}

	// For invitations, the actor must be the invitee
	if requestType == "invitation" && userID != actorID {
		return sql.ErrNoRows
	}

	// For join requests, the actor must be the group creator
	if requestType == "request" {
		isCreator, err := query.IsGroupCreator(groupID, actorID)
		if err != nil || !isCreator {
			return sql.ErrNoRows
		}
	}

	// Process the request
	err := query.HandleGroupMembershipRequest(groupID, userID, action, requestType)
	if err != nil {
		log.Printf("[ERROR] Failed to process membership request: %v", err)
		return err
	}

	// Get group info
	group, err := query.GetGroupByID(groupID, actorID)
	if err == nil {
		// Create notification
		var notificationContent string
		var notificationType string
		var toUserID int

		if requestType == "invitation" {
			// Notification to group creator
			toUserID = group.CreatorID
			if action == "accept" {
				notificationContent = "A user has accepted your invitation to join the group: " + group.Title
				notificationType = "group_invitation_accepted"
			} else {
				notificationContent = "A user has declined your invitation to join the group: " + group.Title
				notificationType = "group_invitation_declined"
			}
		} else {
			// Notification to requester
			toUserID = userID
			if action == "accept" {
				notificationContent = "Your request to join the group has been accepted: " + group.Title
				notificationType = "group_join_request_accepted"
			} else {
				notificationContent = "Your request to join the group has been declined: " + group.Title
				notificationType = "group_join_request_declined"
			}
		}

		// Create notification in database
		_, err = db.DBInstance.DB.Exec(
			"INSERT INTO notifications (to_user_id, from_user_id, content, type, read, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			toUserID, actorID, notificationContent, notificationType, false, time.Now(),
		)
		if err != nil {
			log.Printf("[ERROR] Failed to create notification: %v", err)
		}

		// Send WebSocket notification
		routes.SendToUser(toUserID, routes.Message{
			Type: notificationType,
			Content: map[string]interface{}{
				"group_id":    groupID,
				"group_title": group.Title,
				"action":      action,
			},
		})
	}

	return nil
}
