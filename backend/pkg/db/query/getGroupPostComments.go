package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetGroupPostComments returns all comments for a post
func GetGroupPostComments(postID int) ([]models.GroupComment, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT gc.id, gc.post_id, gc.user_id, gc.content, gc.created_at,
			   u.first_name, u.last_name, u.avatar
		FROM group_comments gc
		JOIN users u ON gc.user_id = u.id
		WHERE gc.post_id = ?
		ORDER BY gc.created_at ASC
	`, postID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.GroupComment
	for rows.Next() {
		var comment models.GroupComment
		var avatar sql.NullString

		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.FirstName,
			&comment.LastName,
			&avatar,
		); err != nil {
			log.Printf("Error scanning group comment row: %v", err)
			continue
		}

		if avatar.Valid {
			comment.Avatar = &avatar.String
		}

		comments = append(comments, comment)
	}

	return comments, nil
}
