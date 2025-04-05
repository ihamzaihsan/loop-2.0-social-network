package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
	"strings"
)

func fixImagePath(image string) string {
	if image != "" {
		// If the image path starts with "./uploads", convert it to "/uploads"
		if strings.HasPrefix(image, "./uploads") {
			return strings.Replace(image, "./uploads", "/uploads", 1)
		}
	}
	return image
}

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
	
	// Fix the image path
	post.Image = fixImagePath(post.Image)
	
	return &post, err
}

func GetPostsByUserID(userID int) ([]models.PostResponse, error) {
	log.Printf("Fetching posts for user ID: %d", userID)
	
	
	query := `
		SELECT p.id, p.user_id, p.content, p.image, p.privacy, p.created_at,
			   u.first_name, u.last_name, u.nickname, u.avatar
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
	`
	
	rows, err := db.DBInstance.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var posts []models.PostResponse
	for rows.Next() {
		var post models.PostResponse
		var author models.User
		var nickname, avatar sql.NullString
		
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&post.Image,
			&post.Privacy,
			&post.CreatedAt,
			&author.FirstName,
			&author.LastName,
			&nickname,
			&avatar,
		)
		
		if err != nil {
			log.Printf("Error scanning post row: %v", err)
			continue
		}
		
		if nickname.Valid {
			nicknameStr := nickname.String
			author.Nickname = &nicknameStr
		}
		if avatar.Valid {
			avatarStr := avatar.String
			author.Avatar = &avatarStr
		}
		
		// Set default values for like count and isLiked since we don't have the post_likes table
		post.LikeCount = 0
		post.IsLiked = false
		
		// Fix the image path
		post.Image = fixImagePath(post.Image)
		
		post.Author = author
		posts = append(posts, post)
	}
	
	log.Printf("Found %d posts for user ID: %d", len(posts), userID)
	return posts, nil
}

func GetUserPostsCount(userID int) (int, error) {
	var count int
	err := db.DBInstance.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ?", userID).Scan(&count)
	return count, err
}
