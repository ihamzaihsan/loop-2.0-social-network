package query

import (
	"fmt"
	"socialNetwork/pkg/db"
)

func DeletePostQuery(postID, userID int) error {
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var owned bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM posts WHERE id = ? AND user_id = ?)`, postID, userID).Scan(&owned); err != nil {
		return err
	}
	if !owned {
		return fmt.Errorf("post not found or unauthorized")
	}
	// Comments use a restrictive foreign key; remove children before the post.
	for _, statement := range []string{
		`DELETE FROM comments WHERE post_id = ?`,
		`DELETE FROM likes WHERE post_id = ?`,
		`DELETE FROM post_viewers WHERE post_id = ?`,
		`DELETE FROM posts WHERE id = ?`,
	} {
		if _, err := tx.Exec(statement, postID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
