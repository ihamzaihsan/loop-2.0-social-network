package query

import (
	"fmt"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func UpdatePostQuery(postID int, userID int, request models.PostRequest) error {
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return err
	}

	query := `UPDATE posts SET content =?,image=? privacy=?
	WHERE id = ? AND user_id = ?`
	result, err := tx.Exec(query, request.Content, request.Image, request.Privacy, postID, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rows == 0 {
		tx.Rollback()
		return fmt.Errorf("post not found or unauthorized")
	}

	if request.Privacy == "private" {
		_, err := tx.Exec(`DELETE FROM post_viewers WHERE post_id = ?)`, postID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, viewerID := range request.ViewerIDs {
		_, err = tx.Exec(`INSERT INTO post_viewers (post_id, viewer_id) VALUES (?, ?)`,
			postID, viewerID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	
	return tx.Commit()
}
