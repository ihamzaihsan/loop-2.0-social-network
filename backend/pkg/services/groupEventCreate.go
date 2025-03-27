package services

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/routes"
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
			// Notify group members
			for _, member := range members {
				if member.UserID != userID {
					// Create notification in database
					_, err = db.DBInstance.DB.Exec(
						"INSERT INTO notifications (to_user_id, from_user_id, content, type, read, created_at) VALUES (?, ?, ?, ?, ?, ?)",
						member.UserID, userID, firstName + " " + lastName + " created a new event: " + title, "group_event", false, time.Now(),
					)
					if err != nil {
						log.Printf("[ERROR] Failed to create notification: %v", err)
					}

					// Send WebSocket notification
					routes.SendToUser(member.UserID, routes.Message{
						Type: "group_event",
						Content: map[string]interface{}{
							"group_id": groupID,
							"event_id": eventID,
							"title": title,
							"creator_id": userID,
							"creator_name": firstName + " " + lastName,
							"event_time": eventTime,
						},
					})
				}
			}
		}
	}

	return eventID, nil
}
