package query

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
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
			_, _ = res.LastInsertId()

			// Create notification if the request is pending
			if isPrivate {
				// Get follower's name for notification
				var firstName, lastName string
				err = db.DBInstance.DB.QueryRow(`
					SELECT first_name, last_name FROM users WHERE id = ?
				`, followerID).Scan(&firstName, &lastName)
				if err != nil {
					log.Printf("Error getting follower name: %v", err)
				}

				// Create notification
				_, err = db.DBInstance.DB.Exec(`
					INSERT INTO notifications (to_user_id, from_user_id, content, type, read, created_at)
					VALUES (?, ?, ?, ?, ?, ?)
				`, followedID, followerID, firstName+" "+lastName+" wants to follow you", "follow_request", false, time.Now())

				if err != nil {
					log.Printf("Error creating notification: %v", err)
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

		// Get follower's name for notification
		var firstName, lastName string
		err = db.DBInstance.DB.QueryRow(`
			SELECT first_name, last_name FROM users WHERE id = ?
		`, followerID).Scan(&firstName, &lastName)
		if err != nil {
			log.Printf("Error getting follower name: %v", err)
		}

		// Create notification
		_, err = db.DBInstance.DB.Exec(`
			INSERT INTO notifications (to_user_id, from_user_id, content, type, read, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, followedID, followerID, firstName+" "+lastName+" wants to follow you", "follow_request", false, time.Now())

		if err != nil {
			log.Printf("Error creating notification: %v", err)
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

	// Get usernames for notification
	var followedUsername string
	err = db.DBInstance.DB.QueryRow(`
		SELECT username FROM users WHERE id = ?
	`, followedID).Scan(&followedUsername)

	if err != nil {
		log.Printf("Error getting followed username: %v", err)
	}

	// Create notification for follower
	_, err = db.DBInstance.DB.Exec(`
		INSERT INTO notifications (to_user_id, from_user_id, content, type, read, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, followerID, followedID, followedUsername+" accepted your follow request", "follow_accept", false, time.Now())

	if err != nil {
		log.Printf("Error creating notification: %v", err)
	}

	return followerID, nil
}

// RejectFollowRequest changes a follow request status to rejected
func RejectFollowRequest(requestID int, followedID uint) error {
	_, err := db.DBInstance.DB.Exec(`
		UPDATE followers 
		SET status = 'reject' 
		WHERE id = ? AND following_id = ?
	`, requestID, followedID)

	return err
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
