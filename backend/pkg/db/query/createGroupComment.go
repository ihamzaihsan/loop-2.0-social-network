package query

import (
	"socialNetwork/pkg/db"
	"time"
)

// CreateGroupComment adds a comment to a group post
func CreateGroupComment(postID, userID int, content string) (int, error) {
	result, err := db.DBInstance.DB.Exec(`
		INSERT INTO group_comments (post_id, user_id, content, created_at)
		VALUES (?, ?, ?, ?)
	`, postID, userID, content, time.Now())

	if err != nil {
		return 0, err
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(commentID), nil
}
