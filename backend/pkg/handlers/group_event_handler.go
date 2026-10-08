package handlers

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/websocket"
	"time"
)

// HandleGroupEventCreation handles group event creation with notifications
func HandleGroupEventCreation(userID int, content map[string]interface{}) {
	// Extract event details
	groupIDFloat, ok := content["group_id"].(float64)
	if !ok {
		log.Printf("[ERROR] Invalid group_id format")
		return
	}
	groupID := int(groupIDFloat)

	title, ok := content["title"].(string)
	if !ok {
		log.Printf("[ERROR] Invalid event title format")
		return
	}

	description, ok := content["description"].(string)
	if !ok {
		log.Printf("[ERROR] Invalid event description format")
		return
	}

	eventTimeStr, ok := content["event_time"].(string)
	if !ok {
		log.Printf("[ERROR] Invalid event time format")
		return
	}

	eventTime, err := time.Parse(time.RFC3339, eventTimeStr)
	if err != nil {
		log.Printf("[ERROR] Failed to parse event time: %v", err)
		return
	}

	// Check if user is a member of the group
	var isMember bool
	err = db.DBInstance.DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM group_members 
            WHERE group_id = ? AND user_id = ? AND status = 'active'
        )
    `, groupID, userID).Scan(&isMember)

	if err != nil {
		log.Printf("[ERROR] Failed to check group membership: %v", err)
		return
	}

	if !isMember {
		log.Printf("[ERROR] User %d is not a member of group %d", userID, groupID)
		return
	}

	// Create the event
	eventID, err := createEventInDatabase(groupID, title, description, eventTime, content)
	if err != nil {
		log.Printf("[ERROR] Failed to create event: %v", err)
		return
	}

	log.Printf("[INFO] Created group event %d", eventID)

	// Create notifications for group members
	createEventNotifications(groupID, userID, int(eventID), title)

	// Get response options and broadcast
	responseOptions := getEventResponseOptions(int(eventID))

	// Create event object for broadcasting
	event := map[string]interface{}{
		"id":               eventID,
		"group_id":         groupID,
		"creator_id":       userID,
		"title":            title,
		"description":      description,
		"event_time":       eventTime,
		"created_at":       time.Now(),
		"going_count":      0,
		"not_going_count":  0,
		"response_options": responseOptions,
	}

	// Broadcast to all group members
	BroadcastToGroupMembers(groupID, userID, websocket.Message{
		Type:    "group_event",
		Content: event,
	})
}

// Helper function to create event in database
func createEventInDatabase(groupID int, title, description string, eventTime time.Time, content map[string]interface{}) (int64, error) {
	// Insert the event
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
        INSERT INTO group_events (group_id, title, description, event_time, created_at)
        VALUES (?, ?, ?, ?, ?)
    `, groupID, title, description, eventTime, time.Now())

	if err != nil {
		return 0, err
	}

	eventID, _ := result.LastInsertId()

	// Add default response options
	optionsInterface, ok := content["options"].([]interface{})
	if !ok || len(optionsInterface) == 0 {
		optionsInterface = []interface{}{"Going", "Not Going"}
	}

	for _, optionInterface := range optionsInterface {
		optionText, ok := optionInterface.(string)
		if !ok {
			continue
		}

		_, err := tx.Exec(`
            INSERT INTO event_response_options (event_id, option_text)
            VALUES (?, ?)
        `, eventID, optionText)

		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return eventID, nil
}

// Helper function to get response options
func getEventResponseOptions(eventID int) []map[string]interface{} {
	rows, err := db.DBInstance.DB.Query(`
        SELECT id, option_text
        FROM event_response_options
        WHERE event_id = ?
    `, eventID)

	if err != nil {
		log.Printf("[ERROR] Failed to get event options: %v", err)
		return nil
	}
	defer rows.Close()

	var responseOptions []map[string]interface{}
	for rows.Next() {
		var id int
		var optionText string
		if err := rows.Scan(&id, &optionText); err != nil {
			log.Printf("[ERROR] Failed to scan event option: %v", err)
			continue
		}

		responseOptions = append(responseOptions, map[string]interface{}{
			"id":          id,
			"event_id":    eventID,
			"option_text": optionText,
		})
	}

	return responseOptions
}

// Helper function to create notifications
func createEventNotifications(groupID, creatorID, eventID int, eventTitle string) {
	// Get group members
	rows, err := db.DBInstance.DB.Query(`
		SELECT user_id FROM group_members WHERE group_id = ? AND user_id != ?
	`, groupID, creatorID)

	if err != nil {
		log.Printf("[ERROR] Failed to get group members: %v", err)
		return
	}
	defer rows.Close()

	// Collect member IDs first to avoid database locking
	var memberIDs []int
	for rows.Next() {
		var memberID int
		if err := rows.Scan(&memberID); err != nil {
			continue
		}
		memberIDs = append(memberIDs, memberID)
	}

	// Get creator and group info
	var creatorFirstName, creatorLastName, groupTitle string

	err = db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name FROM users WHERE id = ?
	`, creatorID).Scan(&creatorFirstName, &creatorLastName)

	if err != nil {
		log.Printf("[ERROR] Failed to get creator info: %v", err)
		return
	}

	err = db.DBInstance.DB.QueryRow(`
		SELECT title FROM groups WHERE id = ?
	`, groupID).Scan(&groupTitle)

	if err != nil {
		log.Printf("[ERROR] Failed to get group title: %v", err)
		return
	}

	creatorName := creatorFirstName + " " + creatorLastName
	notificationContent := creatorName + " created a new event in group " + groupTitle + ": " + eventTitle

	// Create notifications for each member
	for _, memberID := range memberIDs {
		// CHANGE: Use groupID as related_id instead of eventID
		notificationResult, err := db.DBInstance.DB.Exec(`
			INSERT INTO notifications (user_id, from_user_id, type, related_id, content, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, memberID, creatorID, "group_event", groupID, notificationContent, "unread", time.Now()) // Changed eventID to groupID

		if err != nil {
			log.Printf("[ERROR] Failed to create notification for user %d: %v", memberID, err)
			continue
		}

		notificationID, _ := notificationResult.LastInsertId()
		log.Printf("[INFO] Created notification %d for user %d", notificationID, memberID)

		// Send notification via WebSocket
		notificationSent := websocket.SendToUser(memberID, websocket.Message{
			Type: "notification",
			Content: map[string]interface{}{
				"id":           notificationID,
				"type":         "group_event",
				"group_id":     groupID, // This should be the group ID
				"group_title":  groupTitle,
				"event_id":     eventID, // Keep event ID as separate field
				"event_title":  eventTitle,
				"creator_id":   creatorID,
				"creator_name": creatorName,
				"content":      notificationContent,
				"status":       "unread",
				"actions":      []string{},
				"related_id":   groupID, // Make sure this is group ID
			},
		})

		log.Printf("[INFO] Notification WebSocket sent to user %d: %t", memberID, notificationSent)
	}
}

// BroadcastToGroupMembers broadcasts a message to all members of a group
func BroadcastToGroupMembers(groupID, excludeUserID int, message websocket.Message) {
	websocket.PublishChange("groups")
	// Get all members of the group
	rows, err := db.DBInstance.DB.Query(`
        SELECT user_id FROM group_members
        WHERE group_id = ? AND status = 'active'
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

	// Send the message to all members except the sender
	for _, memberID := range memberIDs {
		if memberID != excludeUserID {
			websocket.SendToUser(memberID, message)
		}
	}
}
