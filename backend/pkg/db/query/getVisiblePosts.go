package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func GetVisiblePosts(userID, limit, offset int) ([]models.PostResponse, int, error) {
	var total int
	countQuery := `
        SELECT COUNT(*) FROM posts p
        WHERE p.privacy = 'public'
        OR (p.privacy = 'almost_private' AND EXISTS (
            SELECT 1 FROM followers 
            WHERE follower_id = ? AND following_id = p.user_id
        ))
        OR (p.privacy = 'private' AND EXISTS (
            SELECT 1 FROM post_viewers 
            WHERE viewer_id = ? AND post_id = p.id
        ))
    `
	err := db.DBInstance.DB.QueryRow(countQuery, userID, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
        SELECT
            p.id, p.user_id, p.content, p.image, p.privacy, p.created_at,
            u.first_name, u.last_name, u.nickname, u.avatar
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.privacy = 'public'
        OR (p.privacy = 'almost_private' AND EXISTS (
            SELECT 1 FROM followers
            WHERE follower_id = ? AND following_id = p.user_id
        ))
        OR (p.privacy = 'private' AND EXISTS (
            SELECT 1 FROM post_viewers
            WHERE viewer_id = ? AND post_id = p.id
        ))
        ORDER BY p.created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := db.DBInstance.DB.Query(query, userID, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []models.PostResponse
	for rows.Next() {
		var post models.PostResponse
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &post.Image, &post.Privacy, &post.CreatedAt,
			&post.Author.FirstName, &post.Author.LastName, &post.Author.Nickname, &post.Author.Avatar,
		)
		if err != nil {
			return nil, 0, err
		}

		// Set default values for like count and isLiked since we don't have the likes table
		post.LikeCount = 0
		post.IsLiked = false

		posts = append(posts, post)
	}

	return posts, total, nil
}
