package services

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
)

// CreateGroupCommentService adds a comment to a group post
func CreateGroupCommentService(postID, userID int, content string) (int, error) {
	// Get post to check group membership
	post, err := query.GetGroupPost(postID)
	if err != nil {
		return 0, err
	}

	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(post.GroupID, userID)
	if err != nil || !isMember {
		return 0, err
	}

	// Create comment
	commentID, err := query.CreateGroupComment(postID, userID, content)
	if err != nil {
		log.Printf("[ERROR] Failed to create comment: %v", err)
		return 0, err
	}

	// Only prepare notification data if post creator is different from commenter
	if post.UserID != userID {
		// Get user info
		var firstName, lastName string
		err = db.DBInstance.DB.QueryRow(
			"SELECT first_name, last_name FROM users WHERE id = ?", userID,
		).Scan(&firstName, &lastName)

		if err == nil {
			// Store notification in database instead of directly sending WebSocket message
			_, err := db.DBInstance.DB.Exec(`
				INSERT INTO notifications (user_id, type, related_id, content, created_at)
				VALUES (?, 'group_comment', ?, ?, NOW())
			`, post.UserID, commentID, truncateString(content, 50))

			if err != nil {
				log.Printf("[ERROR] Failed to store notification: %v", err)
			}
		}
	}

	return commentID, nil
}

// Helper function to truncate string
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}
