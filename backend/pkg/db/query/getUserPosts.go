package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func GetUserPosts(userID int) ([]models.Post, int, error) {
	var posts []models.Post
	var count int

	countQuery := `SELECT COUNT(*) FROM posts WHERE user_id = ?`
	err := db.DBInstance.DB.QueryRow(countQuery, userID).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	postQuery := `
        SELECT id, user_id, content, image, privacy, created_at 
        FROM posts 
        WHERE user_id = ? 
        ORDER BY created_at DESC
	`
	rows, errr := db.DBInstance.DB.Query(postQuery)
	if errr != nil {
		return nil, 0, errr
	}
	defer rows.Close()
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Content,
			&post.Image, &post.Privacy, &post.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, post)
	}
	return posts, count, nil
}
