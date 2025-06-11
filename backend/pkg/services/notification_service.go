package services

import (
	"fmt"
	"log"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
	"socialNetwork/pkg/websocket"
)

// CreateNotification creates a new notification and sends it via WebSocket if the user is online
func CreateNotification(userID, fromUserID int, notificationType string, relatedID int, content string) (int64, error) {
	log.Printf("[INFO] Creating notification: userID=%d, fromUserID=%d, type=%s, relatedID=%d", userID, fromUserID, notificationType, relatedID)

	notificationID, err := query.CreateNotification(userID, fromUserID, notificationType, relatedID, content)
	if err != nil {
		log.Printf("[ERROR] Failed to create notification in database: %v", err)
		return 0, err
	}

	log.Printf("[INFO] Created notification in database with ID: %d", notificationID)

	// Send real-time notification via WebSocket if the user is online
	// Get additional notification information for rich display
	notification := models.Notification{
		ID:         int(notificationID),
		UserID:     userID,
		FromUserID: fromUserID,
		Type:       notificationType,
		RelatedID:  relatedID,
		Content:    content,
		Status:     "unread",
	}

	// Get sender details if fromUserID is provided
	if fromUserID > 0 {
		var firstName, lastName string
		err := query.GetUserNames(fromUserID, &firstName, &lastName)
		if err == nil {
			notification.SenderName = fmt.Sprintf("%s %s", firstName, lastName)
			log.Printf("[INFO] Set sender name: %s", notification.SenderName)
		} else {
			log.Printf("[ERROR] Failed to get sender name: %v", err)
		}
	}

	// Get group title if it's a group-related notification
	if notificationType == "group_invitation" || notificationType == "group_join_request" || notificationType == "group_event" {
		group, err := query.GetGroupByID(relatedID, userID)
		if err == nil {
			notification.GroupTitle = group.Title
			log.Printf("[INFO] Set group title: %s", notification.GroupTitle)
		} else {
			log.Printf("[ERROR] Failed to get group title: %v", err)
		}
	}

	// Set available actions based on notification type
	notification.Actions = getNotificationActions(notificationType)

	log.Printf("[INFO] Sending WebSocket notification to user %d", userID)

	// Send WebSocket notification
	sent := websocket.SendToUser(userID, websocket.Message{
		Type:    "notification",
		Content: notification,
	})

	log.Printf("[INFO] WebSocket notification sent to user %d: %t", userID, sent)

	return notificationID, nil
}

// GetUserNotifications retrieves notifications for a specific user
func GetUserNotifications(userID int, limit int) ([]models.Notification, int, error) {
	notifications, err := query.GetUserNotifications(userID, limit)
	if err != nil {
		return nil, 0, err
	}

	// Get unread count
	unreadCount, err := query.GetUnreadNotificationCount(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get unread count: %v", err)
		unreadCount = 0
	}

	return notifications, unreadCount, nil
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(notificationID, userID int) error {
	return query.MarkNotificationAsRead(notificationID, userID)
}

// MarkAllNotificationsAsRead marks all notifications as read for a user
func MarkAllNotificationsAsRead(userID int) error {
	return query.MarkAllNotificationsAsRead(userID)
}

// GetUnreadNotificationCount gets the count of unread notifications for a user
func GetUnreadNotificationCount(userID int) (int, error) {
	return query.GetUnreadNotificationCount(userID)
}

// DeleteNotification deletes a notification
func DeleteNotification(notificationID, userID int) error {
	return query.DeleteNotification(notificationID, userID)
}

// DeleteNotificationAfterAction deletes a notification after it has been acted upon
func DeleteNotificationAfterAction(notificationID, userID int) error {
	return query.DeleteNotification(notificationID, userID)
}

// Helper function to get available actions based on notification type
func getNotificationActions(notificationType string) []string {
	switch notificationType {
	case "follow_request":
		return []string{"accept", "reject"}
	case "group_invitation":
		return []string{"accept", "reject"}
	case "group_join_request":
		return []string{"accept", "reject"}
	case "group_event":
		return []string{}
	default:
		return []string{}
	}
}
