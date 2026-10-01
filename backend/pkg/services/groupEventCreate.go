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
	log.Printf("[INFO] Creating group event: groupID=%d, userID=%d, title=%s", groupID, userID, title)

	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(groupID, userID)
	if err != nil || !isMember {
		log.Printf("[ERROR] User %d is not a member of group %d: %v", userID, groupID, err)
		return 0, err
	}

	// Create event
	eventID, err := query.CreateGroupEvent(groupID, title, description, eventTime)
	if err != nil {
		log.Printf("[ERROR] Failed to create event: %v", err)
		return 0, err
	}

	log.Printf("[INFO] Created event with ID: %d", eventID)

	// Get group members for notification
	members, err := query.GetGroupMembers(groupID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group members: %v", err)
		return eventID, err
	}

	log.Printf("[INFO] Found %d group members", len(members))

	// Get user info for notification
	var firstName, lastName string
	err = db.DBInstance.DB.QueryRow(
		"SELECT first_name, last_name FROM users WHERE id = ?", userID,
	).Scan(&firstName, &lastName)

	if err != nil {
		log.Printf("[ERROR] Failed to get user info: %v", err)
		return eventID, err
	}

	creatorName := firstName + " " + lastName
	log.Printf("[INFO] Creator name: %s", creatorName)

	// Get group info
	group, err := query.GetGroupByID(groupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group info: %v", err)
		return eventID, err
	}

	log.Printf("[INFO] Group title: %s", group.Title)

	// Create notification content
	notificationContent := creatorName + " created a new event in group " + group.Title + ": " + title

	// Notify group members
	for _, member := range members {
		if member.UserID != userID {
			log.Printf("[INFO] Processing member: %d", member.UserID)

			// Create notification in database
			notificationID, err := CreateNotification(
				member.UserID,
				userID,
				"group_event",
				groupID,
				notificationContent,
			)
			if err != nil {
				log.Printf("[ERROR] Failed to create notification for user %d: %v", member.UserID, err)
				continue
			}

			log.Printf("[INFO] Created notification %d for user %d", notificationID, member.UserID)

			// Send notification message (for notification system)
			notificationSent := ws.SendToUser(member.UserID, ws.Message{
				Type: "notification",
				Content: map[string]interface{}{
					"id":           notificationID,
					"type":         "group_event",
					"group_id":     groupID,
					"group_title":  group.Title,
					"event_id":     eventID,
					"event_title":  title,
					"creator_id":   userID,
					"creator_name": creatorName,
					"event_time":   eventTime,
					"content":      notificationContent,
					"status":       "unread",
					"actions":      []string{},
				},
			})

			log.Printf("[INFO] Notification WebSocket sent to user %d: %t", member.UserID, notificationSent)

			// Send group_event message (for real-time event updates)
			eventSent := ws.SendToUser(member.UserID, ws.Message{
				Type: "group_event",
				Content: map[string]interface{}{
					"id":              eventID,
					"group_id":        groupID,
					"title":           title,
					"description":     description,
					"event_time":      eventTime,
					"created_at":      time.Now(),
					"creator_id":      userID,
					"going_count":     0,
					"not_going_count": 0,
				},
			})

			log.Printf("[INFO] Group event WebSocket sent to user %d: %t", member.UserID, eventSent)
		}
	}

	return eventID, nil
}
