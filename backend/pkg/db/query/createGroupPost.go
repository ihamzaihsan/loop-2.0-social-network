package query

import (
	"socialNetwork/pkg/db"
	"time"
)

// CreateGroupPost creates a new post in a group
func CreateGroupPost(groupID, userID int, content, image string) (int, error) {
	result, err := db.DBInstance.DB.Exec(`
		INSERT INTO group_posts (group_id, user_id, content, image, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, groupID, userID, content, image, time.Now())

	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(postID), nil
}
