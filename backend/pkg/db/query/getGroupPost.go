package query

import (
	"database/sql"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetGroupPost returns a specific post by ID
func GetGroupPost(postID int) (*models.GroupPost, error) {
	var post models.GroupPost
	var avatar sql.NullString

	err := db.DBInstance.DB.QueryRow(`
		SELECT gp.id, gp.group_id, gp.user_id, gp.content, gp.image, gp.created_at,
			   u.first_name, u.last_name, u.avatar,
			   (SELECT COUNT(*) FROM group_comments WHERE post_id = gp.id) as comment_count
		FROM group_posts gp
		JOIN users u ON gp.user_id = u.id
		WHERE gp.id = ?
	`, postID).Scan(
		&post.ID,
		&post.GroupID,
		&post.UserID,
		&post.Content,
		&post.Image,
		&post.CreatedAt,
		&post.FirstName,
		&post.LastName,
		&avatar,
		&post.CommentCount,
	)

	if err != nil {
		return nil, err
	}

	if avatar.Valid {
		post.Avatar = &avatar.String
	}

	return &post, nil
}
