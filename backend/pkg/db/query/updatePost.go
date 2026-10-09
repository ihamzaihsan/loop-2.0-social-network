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

	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE posts SET content=?,privacy=? WHERE id=? AND user_id=?`, request.Content, request.Privacy, postID, userID); err != nil {
		return err
	}
	if request.Privacy != "private" || request.ViewerIDs != nil {
		if _, err = tx.Exec(`DELETE FROM post_viewers WHERE post_id=?`, postID); err != nil {
			return err
		}
		if request.Privacy == "private" {
			for _, id := range request.ViewerIDs {
				if _, err = tx.Exec(`INSERT INTO post_viewers(post_id,viewer_id) VALUES (?,?)`, postID, id); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit()
}
