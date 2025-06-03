package query

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
	"socialNetwork/pkg/websocket"
)

// RequestFollow handles the database operations for a follow request
func RequestFollow(followerID, followedID uint) (string, error) {
	// First check if the follow relationship already exists
	var status string
	var id int
	err := db.DBInstance.DB.QueryRow(`
		SELECT id, status FROM followers 
		WHERE follower_id = ? AND following_id = ?
	`, followerID, followedID).Scan(&id, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			// Need to determine if the followed user has a private account
			var isPrivate bool
			err := db.DBInstance.DB.QueryRow(`
				SELECT is_private FROM users WHERE id = ?
			`, followedID).Scan(&isPrivate)
			if err != nil {
				return "", err
			}

			// Set initial status based on account privacy
			initialStatus := "accept"
			if isPrivate {
				initialStatus = "pending"
			}

			// Insert the new follow relationship
			res, err := db.DBInstance.DB.Exec(`
				INSERT INTO followers (follower_id, following_id, status, created_at) 
				VALUES (?, ?, ?, ?)
			`, followerID, followedID, initialStatus, time.Now())
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") {
					return "", errors.New("follow relationship already exists")
				}
				return "", err
			}

			// Get the ID of the inserted row
			followID, _ := res.LastInsertId()

			// Create notification if the request is pending
			if isPrivate {
				var firstName, lastName string
				err = db.DBInstance.DB.QueryRow(`
					SELECT first_name, last_name FROM users WHERE id = ?
				`, followerID).Scan(&firstName, &lastName)

				if err == nil {
					notificationContent := firstName + " " + lastName + " wants to follow you"

					// Create the notification in database
					notificationID, err := db.DBInstance.DB.Exec(`
						INSERT INTO notifications (user_id, from_user_id, type, related_id, content, status, created_at)
						VALUES (?, ?, ?, ?, ?, ?, ?)
					`, followedID, followerID, "follow_request", followID, notificationContent, "unread", time.Now())

					if err != nil {
						log.Printf("Error creating notification: %v", err)
					} else {
						// Send immediate WebSocket notification
						id, _ := notificationID.LastInsertId()

						// Send real-time notification via WebSocket if the user is online
						websocket.SendToUser(int(followedID), websocket.Message{
							Type: "notification",
							Content: map[string]interface{}{
								"id":           id,
								"type":         "follow_request",
								"user_id":      followedID,
								"from_user_id": followerID,
								"sender_name":  firstName + " " + lastName,
								"content":      notificationContent,
								"related_id":   followID,
								"status":       "unread",
								"created_at":   time.Now().Format(time.RFC3339),
								"actions":      []string{"accept", "reject"},
							},
						})
					}
				}

				return "pending", nil
			}

			return initialStatus, nil
		}
		return "", err
	}

	// Follow relationship already exists, handle based on current status
	if status == "pending" {
		return "pending", nil
	} else if status == "reject" {
		// Update rejected request to pending
		_, err := db.DBInstance.DB.Exec(`
			UPDATE followers SET status = 'pending', created_at = ? WHERE id = ?
		`, time.Now(), id)
		if err != nil {
			return "", err
		}

		// Create notification
		notificationRes, err := db.DBInstance.DB.Exec(`
			INSERT INTO notifications (user_id, type, related_id, status, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, followedID, "follow_request", id, "unread", time.Now())

		if err != nil {
			log.Printf("Error creating notification: %v", err)
		} else {
			// Also send WebSocket notification for rejected requests that are re-requested
			notificationID, _ := notificationRes.LastInsertId()

			// Get user details for notification content
			var firstName, lastName string
			err = db.DBInstance.DB.QueryRow(`
				SELECT first_name, last_name FROM users WHERE id = ?
			`, followerID).Scan(&firstName, &lastName)

			if err == nil {
				notificationContent := firstName + " " + lastName + " wants to follow you"

				// Send the real-time notification
				websocket.SendToUser(int(followedID), websocket.Message{
					Type: "notification",
					Content: map[string]interface{}{
						"id":           notificationID,
						"type":         "follow_request",
						"user_id":      followedID,
						"from_user_id": followerID,
						"sender_name":  firstName + " " + lastName,
						"content":      notificationContent,
						"related_id":   id,
						"status":       "unread",
						"created_at":   time.Now().Format(time.RFC3339),
						"actions":      []string{"accept", "reject"},
					},
				})
			}
		}

		return "pending", nil
	}

	return status, nil
}

// UnfollowUser removes a follow relationship
func UnfollowUser(followerID, followedID uint) error {
	_, err := db.DBInstance.DB.Exec(`
		DELETE FROM followers 
		WHERE follower_id = ? AND following_id = ?
	`, followerID, followedID)

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
				"id": notificationID,
				"type": "follow_reject",
				"user_id": followerID,
				"from_user_id": followedID,
				"sender_name": followedName,
				"content": notificationContent,
				"related_id": requestID,
				"status": "unread",
				"created_at": time.Now().Format(time.RFC3339),
			},
		})
		
		// Also send a "follow_request_handled" message to update the UI
		websocket.SendToUser(int(followedID), websocket.Message{
			Type: "follow_request_handled",
			Content: map[string]interface{}{
				"request_id": requestID,
				"follower_id": followerID,
				"action": "reject",
				"follower_name": followerName,
			},
		})
	}

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

	var friends []models.Follow
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
