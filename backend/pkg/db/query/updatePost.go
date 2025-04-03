package query

import (
	"fmt"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func UpdatePostQuery(postID, userID int, request models.PostRequest) error {
	// First verify the post exists and belongs to the user
	var count int
	err := db.DBInstance.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ? AND user_id = ?",
		postID, userID).Scan(&count)

	if err != nil {
		log.Printf("[ERROR] Error checking post ownership: %v", err)
		return err
	}

	if count == 0 {
		return fmt.Errorf("post not found or you don't have permission to edit it")
	}

	// Update the post
	_, err = db.DBInstance.DB.Exec(`
		UPDATE posts 
		SET content = ?, privacy = ?
		WHERE id = ? AND user_id = ?
	`, request.Content, request.Privacy, postID, userID)

	if err != nil {
		log.Printf("[ERROR] Error updating post: %v", err)
		return err
	}

	return nil
}
