package query

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
	"socialNetwork/pkg/websocket"
)

// RequestFollow commits the relationship and its notification together.
func RequestFollow(followerID, followedID uint) (string, error) {
	if followerID == 0 || followedID == 0 || followerID == followedID {
		return "", errors.New("invalid follow target")
	}
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var followID int64
	var status string
	err = tx.QueryRow(`SELECT id, status FROM followers WHERE follower_id = ? AND following_id = ?`, followerID, followedID).Scan(&followID, &status)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if err == nil && (status == "accept" || status == "pending") {
		return status, nil
	}
	var private bool
	if err := tx.QueryRow(`SELECT isprivate FROM users WHERE id = ?`, followedID).Scan(&private); err != nil {
		return "", err
	}
	status = "accept"
	kind := "new_follower"
	if private {
		status = "pending"
		kind = "follow_request"
	}
	if followID == 0 {
		result, err := tx.Exec(`INSERT INTO followers (follower_id, following_id, status, created_at) VALUES (?, ?, ?, ?)`, followerID, followedID, status, time.Now())
		if err != nil {
			return "", err
		}
		followID, err = result.LastInsertId()
		if err != nil {
			return "", err
		}
	} else {
		if _, err := tx.Exec(`UPDATE followers SET status = ?, created_at = ? WHERE id = ?`, status, time.Now(), followID); err != nil {
			return "", err
		}
	}
	var first, last string
	if err := tx.QueryRow(`SELECT first_name, last_name FROM users WHERE id = ?`, followerID).Scan(&first, &last); err != nil {
		return "", err
	}
	content := first + " " + last + " started following you"
	if private {
		content = first + " " + last + " wants to follow you"
	}
	created := time.Now()
	result, err := tx.Exec(`INSERT INTO notifications (user_id, from_user_id, type, related_id, content, status, created_at) VALUES (?, ?, ?, ?, ?, 'unread', ?)`, followedID, followerID, kind, followID, content, created)
	if err != nil {
		return "", err
	}
	notificationID, err := result.LastInsertId()
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	actions := []string{}
	if private {
		actions = []string{"accept", "reject"}
	}
	websocket.SendToUser(int(followedID), websocket.Message{Type: "notification", Content: map[string]interface{}{
		"id": notificationID, "type": kind, "user_id": followedID, "from_user_id": followerID,
		"sender_name": first + " " + last, "content": content, "related_id": followID,
		"status": "unread", "created_at": created.Format(time.RFC3339), "actions": actions,
	}})
	publishSocialChange(followerID, followedID)
	return status, nil
}

// UnfollowUser removes a follow relationship
func UnfollowUser(followerID, followedID uint) error {
	_, err := db.DBInstance.DB.Exec(`
		DELETE FROM followers 
		WHERE follower_id = ? AND following_id = ?
	`, followerID, followedID)

	if err == nil {
		publishSocialChange(followerID, followedID)
	}
	return err
}

// AcceptFollowRequest changes a follow request status to accepted
func AcceptFollowRequest(requestID int, followedID uint) (uint, error) {
	// First get the follower ID
	var followerID uint
	err := db.DBInstance.DB.QueryRow(`
		SELECT follower_id 
		FROM followers 
		WHERE id = ? AND following_id = ? AND status = 'pending'
	`, requestID, followedID).Scan(&followerID)

	if err != nil {
		return 0, err
	}

	// Update status to accept
	_, err = db.DBInstance.DB.Exec(`
		UPDATE followers 
		SET status = 'accept' 
		WHERE id = ? AND following_id = ?
	`, requestID, followedID)

	if err != nil {
		return 0, err
	}

	// Get user names for notification
	var followerFirstName, followerLastName string
	var followedFirstName, followedLastName string

	// Get follower's name
	err = db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name FROM users WHERE id = ?
	`, followerID).Scan(&followerFirstName, &followerLastName)
	if err != nil {
		log.Printf("Error getting follower's name: %v", err)
	}

	// Get followed user's name
	err = db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name FROM users WHERE id = ?
	`, followedID).Scan(&followedFirstName, &followedLastName)
	if err != nil {
		log.Printf("Error getting followed user's name: %v", err)
	}

	// Create notification content
	followerName := followerFirstName + " " + followerLastName
	followedName := followedFirstName + " " + followedLastName
	notificationContent := followedName + " accepted your follow request"

	// Create notification in database for follower
	notificationRes, err := db.DBInstance.DB.Exec(`
		INSERT INTO notifications (user_id, from_user_id, type, related_id, content, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, followerID, followedID, "follow_accept", requestID, notificationContent, "unread", time.Now())

	if err != nil {
		log.Printf("Error creating notification: %v", err)
	} else {
		// Get notification ID
		notificationID, _ := notificationRes.LastInsertId()

		// Send WebSocket notification to follower
		websocket.SendToUser(int(followerID), websocket.Message{
			Type: "notification",
			Content: map[string]interface{}{
				"id":           notificationID,
				"type":         "follow_accept",
				"user_id":      followerID,
				"from_user_id": followedID,
				"sender_name":  followedName,
				"content":      notificationContent,
				"related_id":   requestID,
				"status":       "unread",
				"created_at":   time.Now().Format(time.RFC3339),
			},
		})

		// Send follow status update message
		websocket.SendToUser(int(followerID), websocket.Message{
			Type: "follow_status_update",
			Content: map[string]interface{}{
				"followed_id": followedID,
				"status":      "accept",
				"username":    followedName,
			},
		})

		// Get the notification ID for the original follow request
		var originalNotificationID int
		err = db.DBInstance.DB.QueryRow(`
			SELECT id FROM notifications 
			WHERE type = 'follow_request' 
			AND related_id = ? 
			AND user_id = ?
		`, requestID, followedID).Scan(&originalNotificationID)

		if err == nil {
			// Delete the original follow request notification
			_, err = db.DBInstance.DB.Exec(`
				DELETE FROM notifications 
				WHERE id = ?
			`, originalNotificationID)
			if err != nil {
				log.Printf("Error deleting original notification: %v", err)
			}

			// Also send a "follow_request_handled" message to followed user to update UI
			websocket.SendToUser(int(followedID), websocket.Message{
				Type: "follow_request_handled",
				Content: map[string]interface{}{
					"request_id":    requestID,
					"follower_id":   followerID,
					"action":        "accept",
					"follower_name": followerName,
				},
			})
		}
	}

	publishSocialChange(followerID, followedID)
	return followerID, nil
}

// RejectFollowRequest changes a follow request status to rejected
func RejectFollowRequest(requestID int, followedID uint) error {
	// First get the follower ID
	var followerID uint
	err := db.DBInstance.DB.QueryRow(`
		SELECT follower_id 
		FROM followers 
		WHERE id = ? AND following_id = ? AND status = 'pending'
	`, requestID, followedID).Scan(&followerID)

	if err != nil {
		return err
	}

	// Update status to reject
	_, err = db.DBInstance.DB.Exec(`
		UPDATE followers 
		SET status = 'reject' 
		WHERE id = ? AND following_id = ?
	`, requestID, followedID)

	if err != nil {
		return err
	}

	// Get user names for notification purposes
	var followerFirstName, followerLastName string
	var followedFirstName, followedLastName string

	// Get follower's name
	err = db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name FROM users WHERE id = ?
	`, followerID).Scan(&followerFirstName, &followerLastName)
	if err != nil {
		log.Printf("Error getting follower's name: %v", err)
	}

	// Get followed user's name
	err = db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name FROM users WHERE id = ?
	`, followedID).Scan(&followedFirstName, &followedLastName)
	if err != nil {
		log.Printf("Error getting followed user's name: %v", err)
	}

	// Create notification content
	followerName := followerFirstName + " " + followerLastName
	followedName := followedFirstName + " " + followedLastName
	notificationContent := followedName + " rejected your follow request"

	// Create notification for follower
	notificationRes, err := db.DBInstance.DB.Exec(`
		INSERT INTO notifications (user_id, from_user_id, type, related_id, content, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, followerID, followedID, "follow_reject", requestID, notificationContent, "unread", time.Now())

	if err != nil {
		log.Printf("Error creating notification: %v", err)
	} else {
		// Get notification ID
		notificationID, _ := notificationRes.LastInsertId()

		// Send WebSocket notification to follower
		websocket.SendToUser(int(followerID), websocket.Message{
			Type: "notification",
			Content: map[string]interface{}{
				"id":           notificationID,
				"type":         "follow_reject",
				"user_id":      followerID,
				"from_user_id": followedID,
				"sender_name":  followedName,
				"content":      notificationContent,
				"related_id":   requestID,
				"status":       "unread",
				"created_at":   time.Now().Format(time.RFC3339),
			},
		})

		// Get the notification ID for the original follow request
		var originalNotificationID int
		err = db.DBInstance.DB.QueryRow(`
			SELECT id FROM notifications 
			WHERE type = 'follow_request' 
			AND related_id = ? 
			AND user_id = ?
		`, requestID, followedID).Scan(&originalNotificationID)

		if err == nil {
			// Delete the original follow request notification
			_, err = db.DBInstance.DB.Exec(`
				DELETE FROM notifications 
				WHERE id = ?
			`, originalNotificationID)
			if err != nil {
				log.Printf("Error deleting original notification: %v", err)
			}

			// Also send a "follow_request_handled" message to update the UI
			websocket.SendToUser(int(followedID), websocket.Message{
				Type: "follow_request_handled",
				Content: map[string]interface{}{
					"request_id":    requestID,
					"follower_id":   followerID,
					"action":        "reject",
					"follower_name": followerName,
				},
			})
		}
	}

	publishSocialChange(followerID, followedID)
	return nil
}

// GetFollowers returns the list of users who follow a user
func GetFollowers(userID uint) ([]models.Follow, int, error) {
	// First get the count
	var count int
	countErr := db.DBInstance.DB.QueryRow(`
		SELECT COUNT(*) FROM followers
		WHERE following_id = ? AND status = 'accept'
	`, userID).Scan(&count)

	if countErr != nil {
		return nil, 0, countErr
	}

	rows, err := db.DBInstance.DB.Query(`
		SELECT f.id, f.follower_id, f.following_id, f.status,
			u.first_name, u.last_name, u.avatar,
			(SELECT COUNT(*) FROM followers 
			WHERE follower_id = ? AND following_id = f.follower_id AND status = 'accept') as is_following
		FROM followers f
		JOIN users u ON f.follower_id = u.id
		WHERE f.following_id = ? AND f.status = 'accept'
	`, userID, userID)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var followers []models.Follow
	for rows.Next() {
		var follower models.Follow
		var firstName, lastName sql.NullString
		var avatar sql.NullString
		var isFollowing int

		err := rows.Scan(
			&follower.ID,
			&follower.FollowerID,
			&follower.FollowedID,
			&follower.Status,
			&firstName,
			&lastName,
			&avatar,
			&isFollowing,
		)

		if err != nil {
			log.Printf("Error scanning follower: %v", err)
			continue
		}

		// Combine first and last name to create a username
		if firstName.Valid && lastName.Valid {
			follower.Username = firstName.String + " " + lastName.String
		} else if firstName.Valid {
			follower.Username = firstName.String
		} else if lastName.Valid {
			follower.Username = lastName.String
		}

		if avatar.Valid {
			follower.Avatar = avatar.String
		}

		followers = append(followers, follower)
	}

	return followers, count, nil
}

// GetFollowing returns the list of users a user follows
func GetFollowing(userID uint) ([]models.Follow, int, error) {
	// First get the count
	var count int
	countErr := db.DBInstance.DB.QueryRow(`
		SELECT COUNT(*) FROM followers
		WHERE follower_id = ? AND status = 'accept'
	`, userID).Scan(&count)

	if countErr != nil {
		return nil, 0, countErr
	}

	rows, err := db.DBInstance.DB.Query(`
		SELECT f.id, f.follower_id, f.following_id, f.status,
			u.first_name, u.last_name, u.avatar
		FROM followers f
		JOIN users u ON f.following_id = u.id
		WHERE f.follower_id = ? AND f.status = 'accept'
	`, userID)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var following []models.Follow
	for rows.Next() {
		var follow models.Follow
		var firstName, lastName sql.NullString
		var avatar sql.NullString

		err := rows.Scan(
			&follow.ID,
			&follow.FollowerID,
			&follow.FollowedID,
			&follow.Status,
			&firstName,
			&lastName,
			&avatar,
		)

		if err != nil {
			log.Printf("Error scanning following: %v", err)
			continue
		}

		// Combine first and last name to create a username
		if firstName.Valid && lastName.Valid {
			follow.Username = firstName.String + " " + lastName.String
		} else if firstName.Valid {
			follow.Username = firstName.String
		} else if lastName.Valid {
			follow.Username = lastName.String
		}

		if avatar.Valid {
			follow.Avatar = avatar.String
		}

		following = append(following, follow)
	}

	return following, count, nil
}

// GetFollowRequests returns pending follow requests for a user
func GetFollowRequests(userID uint) ([]models.Follow, error) {
	log.Printf("Repository: Fetching follow requests for user ID: %d", userID)

	// First, check if there are any pending requests
	var count int
	err := db.DBInstance.DB.QueryRow(`
        SELECT COUNT(*) FROM followers 
        WHERE following_id = ? AND status = 'pending'
    `, userID).Scan(&count)

	if err != nil {
		log.Printf("Error counting pending requests: %v", err)
		return nil, err
	}

	log.Printf("Repository: Found %d pending follow requests in database", count)

	// Try with first_name and last_name instead of username
	rows, err := db.DBInstance.DB.Query(`
        SELECT f.id, f.follower_id, f.following_id, f.status,
            u.first_name, u.last_name, u.avatar
        FROM followers f
        JOIN users u ON f.follower_id = u.id
        WHERE f.following_id = ? AND f.status = 'pending'
    `, userID)

	if err != nil {
		log.Printf("Error querying follow requests: %v", err)
		return nil, err
	}
	defer rows.Close()

	var requests []models.Follow
	for rows.Next() {
		var request models.Follow
		var firstName, lastName sql.NullString
		var avatar sql.NullString

		err := rows.Scan(
			&request.ID,
			&request.FollowerID,
			&request.FollowedID,
			&request.Status,
			&firstName,
			&lastName,
			&avatar,
		)

		if err != nil {
			log.Printf("Error scanning follow request: %v", err)
			continue
		}

		// Combine first and last name for username
		if firstName.Valid && lastName.Valid {
			request.Username = firstName.String + " " + lastName.String
		} else if firstName.Valid {
			request.Username = firstName.String
		} else if lastName.Valid {
			request.Username = lastName.String
		} else {
			request.Username = "Unknown User"
		}

		if avatar.Valid {
			request.Avatar = avatar.String
		}

		requests = append(requests, request)
		log.Printf("Repository: Found request ID=%d from follower ID=%d", request.ID, request.FollowerID)
	}

	if len(requests) == 0 {
		log.Printf("Repository: No follow requests found after scanning rows")
	}

	return requests, nil
}

// GetFollowStatuses returns the status of all follow relationships for a user
func GetFollowStatuses(userID uint) ([]models.Follow, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT id, follower_id, following_id, status
		FROM followers
		WHERE follower_id = ?
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []models.Follow
	for rows.Next() {
		var status models.Follow

		err := rows.Scan(
			&status.ID,
			&status.FollowerID,
			&status.FollowedID,
			&status.Status,
		)

		if err != nil {
			log.Printf("Error scanning follow status: %v", err)
			continue
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// GetFriends returns users who have a mutual follow relationship with the given user
func GetFriends(userID uint) ([]models.Follow, int, error) {
	// First get the count
	var count int
	countErr := db.DBInstance.DB.QueryRow(`
		SELECT COUNT(*) FROM followers f1
		JOIN followers f2 ON f1.follower_id = f2.following_id AND f1.following_id = f2.follower_id
		WHERE f1.follower_id = ? AND f1.status = 'accept' AND f2.status = 'accept'
	`, userID).Scan(&count)

	if countErr != nil {
		return nil, 0, countErr
	}

	// Query to get mutual followers (friends)
	rows, err := db.DBInstance.DB.Query(`
		SELECT f1.id, f1.follower_id, f1.following_id, f1.status,
			u.first_name, u.last_name, u.avatar
		FROM followers f1
		JOIN followers f2 ON f1.following_id = f2.follower_id AND f1.follower_id = f2.following_id
		JOIN users u ON f1.following_id = u.id
		WHERE f1.follower_id = ? AND f1.status = 'accept' AND f2.status = 'accept'
	`, userID)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	friends := make([]models.Follow, 0)
	for rows.Next() {
		var friend models.Follow
		var firstName, lastName sql.NullString
		var avatar sql.NullString

		err := rows.Scan(
			&friend.ID,
			&friend.FollowerID,
			&friend.FollowedID,
			&friend.Status,
			&firstName,
			&lastName,
			&avatar,
		)

		if err != nil {
			log.Printf("Error scanning friend: %v", err)
			continue
		}

		// Combine first and last name to create a username
		if firstName.Valid && lastName.Valid {
			friend.Username = firstName.String + " " + lastName.String
		} else if firstName.Valid {
			friend.Username = firstName.String
		} else if lastName.Valid {
			friend.Username = lastName.String
		}

		if avatar.Valid {
			friend.Avatar = avatar.String
		}

		friends = append(friends, friend)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return friends, count, nil
}

// CheckIfFollowing checks if userID follows targetUserID
func CheckIfFollowing(userID, targetUserID uint) (bool, error) {
	var count int
	err := db.DBInstance.DB.QueryRow(`
		SELECT COUNT(*) FROM followers 
		WHERE follower_id = ? AND following_id = ? AND status = 'accept'
	`, userID, targetUserID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Refresh all relationship views after follow, unfollow, acceptance, or rejection.
func publishSocialChange(followerID, followedID uint) {
	message := websocket.Message{Type: "social_graph_updated", Content: map[string]interface{}{
		"follower_id": followerID, "followed_id": followedID,
	}}
	websocket.SendToUser(int(followerID), message)
	websocket.SendToUser(int(followedID), message)
}
