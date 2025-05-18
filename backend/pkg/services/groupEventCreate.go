package services

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	ws "socialNetwork/pkg/websocket"
	"time"
)

// CreateGroupEventService creates a new event in a group
func CreateGroupEventService(groupID, userID int, title, description string, eventTime time.Time) (int, error) {
	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(groupID, userID)
	if err != nil || !isMember {
		return 0, err
	}

	// Create event
	eventID, err := query.CreateGroupEvent(groupID, title, description, eventTime)
	if err != nil {
		log.Printf("[ERROR] Failed to create event: %v", err)
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
			creatorName := firstName + " " + lastName
			// Get group info
			group, err := query.GetGroupByID(groupID, userID)
			if err != nil {
				log.Printf("[ERROR] Failed to get group info: %v", err)
				return eventID, nil
			}

			// Create notification content
			notificationContent := creatorName + " created a new event in group " + group.Title + ": " + title

			// Notify group members
			for _, member := range members {
				if member.UserID != userID {
					// Create notification in database
					notificationID, err := CreateNotification(
						member.UserID,
						userID,
						"group_event",
						eventID,
						notificationContent,
					)
					if err != nil {
						log.Printf("[ERROR] Failed to create notification: %v", err)
						continue
					}

					// Send WebSocket notification
					ws.SendToUser(member.UserID, ws.Message{
						Type: "notification",
						Content: map[string]interface{}{
							"id":           notificationID,
							"type":         "group_event",
							"group_id":     groupID,
							"group_title":  group.Title,
							"event_id":     eventID,
							"title":        title,
							"creator_id":   userID,
							"creator_name": creatorName,
							"event_time":   eventTime,
							"content":      notificationContent,
						},
					})
				}
			}
		}
	}

	return eventID, nil
}
