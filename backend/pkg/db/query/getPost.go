package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func GetPostByIDQuery(postID int) (*models.Post, error) {
	query := `SELECT id, user_id,content,image,privacy,created_at FROM posts WHERE id = ?`
	var post models.Post
	err := db.DBInstance.DB.QueryRow(query, postID).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&post.Image,
		&post.Privacy,
		&post.CreatedAt,
	)
	return &post, err
}
