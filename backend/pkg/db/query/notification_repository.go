package query

import (
	"database/sql"
	"log"
	"time"

	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
	"socialNetwork/pkg/websocket"
)

// CreateNotification creates a new notification in the database
func CreateNotification(userID, fromUserID int, notificationType string, relatedID int, content string) (int64, error) {
	stmt, err := db.DBInstance.DB.Prepare(`
		INSERT INTO notifications (user_id, from_user_id, type, related_id, content, status, created_at)
		VALUES (?, ?, ?, ?, ?, 'unread', ?)
	`)
	if err != nil {
		log.Printf("[ERROR] Failed to prepare notification insert: %v", err)
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID, fromUserID, notificationType, relatedID, content, time.Now())
	if err != nil {
		log.Printf("[ERROR] Failed to create notification: %v", err)
		return 0, err
	}

	return res.LastInsertId()
}

// GetUserNotifications retrieves notifications for a specific user
func GetUserNotifications(userID int, limit int) ([]models.Notification, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT n.id, n.user_id, n.from_user_id, n.type, n.related_id, n.content, n.status, n.created_at,
		       u.first_name, u.last_name, u.avatar
		FROM notifications n
		LEFT JOIN users u ON n.from_user_id = u.id
		WHERE n.user_id = ?
		ORDER BY n.created_at DESC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve notifications: %v", err)
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		var firstName, lastName sql.NullString
		var avatar sql.NullString
		var fromUserID sql.NullInt64
		var relatedID sql.NullInt64
		var content sql.NullString

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&fromUserID,
			&notification.Type,
			&relatedID,
			&content,
			&notification.Status,
			&notification.CreatedAt,
			&firstName,
			&lastName,
			&avatar,
		)
		if err != nil {
			log.Printf("[ERROR] Error scanning notification: %v", err)
			continue
		}

		// Set FromUserID only if the value is valid
		if fromUserID.Valid {
			notification.FromUserID = int(fromUserID.Int64)
		}

		// Set RelatedID only if the value is valid
		if relatedID.Valid {
			notification.RelatedID = int(relatedID.Int64)
		}

		// Set Content only if the value is valid
		if content.Valid {
			notification.Content = content.String
		}

		// Construct sender name
		if firstName.Valid && lastName.Valid {
			notification.SenderName = firstName.String + " " + lastName.String
		} else if firstName.Valid {
			notification.SenderName = firstName.String
		} else if lastName.Valid {
			notification.SenderName = lastName.String
		}

		// Set avatar if available
		if avatar.Valid {
			notification.SenderAvatar = avatar.String
		}

		// Set available actions based on notification type
		notification.Actions = getNotificationActions(notification.Type)

		// If it's a group related notification and we have a valid RelatedID, get the group title
		if notification.Type == "group_invitation" || notification.Type == "group_join_request" || notification.Type == "group_event" {
			if relatedID.Valid {
				var groupTitle string
				err := db.DBInstance.DB.QueryRow(`
					SELECT title FROM groups WHERE id = ?
				`, notification.RelatedID).Scan(&groupTitle)
				if err == nil {
					notification.GroupTitle = groupTitle
				}
			}
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(notificationID, userID int) error {
	_, err := db.DBInstance.DB.Exec(`
		UPDATE notifications
		SET status = 'read'
		WHERE id = ? AND user_id = ?
	`, notificationID, userID)
	if err == nil {
		websocket.SendToUser(userID, websocket.Message{Type: "notification_changed", Content: map[string]interface{}{"user_id": userID}})
	}
	return err
}

// MarkAllNotificationsAsRead marks all notifications as read for a user
func MarkAllNotificationsAsRead(userID int) error {
	_, err := db.DBInstance.DB.Exec(`
		UPDATE notifications
		SET status = 'read'
		WHERE user_id = ? AND status = 'unread'
	`, userID)
	if err == nil {
		websocket.SendToUser(userID, websocket.Message{Type: "notification_changed", Content: map[string]interface{}{"user_id": userID}})
	}
	return err
}

// GetUnreadNotificationCount gets the count of unread notifications for a user
func GetUnreadNotificationCount(userID int) (int, error) {
	var count int
	err := db.DBInstance.DB.QueryRow(`
		SELECT COUNT(*) FROM notifications
		WHERE user_id = ? AND status = 'unread'
	`, userID).Scan(&count)
	return count, err
}

// DeleteNotification deletes a notification
func DeleteNotification(notificationID, userID int) error {
	_, err := db.DBInstance.DB.Exec(`
		DELETE FROM notifications
		WHERE id = ? AND user_id = ?
	`, notificationID, userID)
	if err == nil {
		websocket.SendToUser(userID, websocket.Message{Type: "notification_changed", Content: map[string]interface{}{"user_id": userID}})
	}
	return err
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
