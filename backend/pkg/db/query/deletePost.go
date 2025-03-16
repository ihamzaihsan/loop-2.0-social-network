package query

import (
	"fmt"
	"socialNetwork/pkg/db"
)

func DeletePostQuery(postID, userID int) error {
	query := `DELETE FROM posts WHERE id = ? AND user_id = ?`
	result, err := db.DBInstance.DB.Exec(query, postID, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("post not found or unauthorized")
	}
	return nil
}
