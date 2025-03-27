package services

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/routes"
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

	// Notify post creator if different from commenter
	if post.UserID != userID {
		// Get user info
		var firstName, lastName string
		err = db.DBInstance.DB.QueryRow(
			"SELECT first_name, last_name FROM users WHERE id = ?", userID,
		).Scan(&firstName, &lastName)

		if err == nil {
			// Send WebSocket notification
			routes.SendToUser(post.UserID, routes.Message{
				Type: "group_comment",
				Content: map[string]interface{}{
					"group_id":        post.GroupID,
					"post_id":         post.ID,
					"comment_id":      commentID,
					"user_id":         userID,
					"user_name":       firstName + " " + lastName,
					"content_preview": truncateString(content, 50),
				},
			})
		}
	}

	return commentID, nil
}
