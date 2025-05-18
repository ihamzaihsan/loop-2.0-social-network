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
	notificationID, err := query.CreateNotification(userID, fromUserID, notificationType, relatedID, content)
	if err != nil {
		return 0, err
	}

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
		}
	}

	// Get group title if it's a group-related notification
	if notificationType == "group_invitation" || notificationType == "group_join_request" || notificationType == "group_event" {
		group, err := query.GetGroupByID(relatedID, userID)
		if err == nil {
			notification.GroupTitle = group.Title
		}
	}

	// Set available actions based on notification type
	notification.Actions = getNotificationActions(notificationType)

	// Send WebSocket notification
	websocket.SendToUser(userID, websocket.Message{
		Type:    "notification",
		Content: notification,
	})

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

// Helper function to get available actions based on notification type
func getNotificationActions(notificationType string) []string {
	switch notificationType {
	case "follow_request":
		return []string{"accept", "reject"}
	case "group_invitation":
		return []string{"accept", "reject"}
	case "group_join_request":
		return []string{"accept", "reject"}
	default:
		return []string{}
	}
}
