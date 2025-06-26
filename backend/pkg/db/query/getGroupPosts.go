package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
	"strings"
)

// GetGroupPosts returns all posts in a group
func GetGroupPosts(groupID int) ([]models.GroupPost, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT gp.id, gp.group_id, gp.user_id, gp.content, gp.image, gp.created_at,
			   u.first_name, u.last_name, u.avatar,
			   (SELECT COUNT(*) FROM group_comments WHERE post_id = gp.id) as comment_count
		FROM group_posts gp
		JOIN users u ON gp.user_id = u.id
		WHERE gp.group_id = ?
		ORDER BY gp.created_at DESC
	`, groupID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.GroupPost
	for rows.Next() {
		var post models.GroupPost
		var avatar sql.NullString

		if err := rows.Scan(
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
		); err != nil {
			log.Printf("Error scanning group post row: %v", err)
			continue
		}

		if avatar.Valid {
			post.Avatar = &avatar.String
		}

		// Fix image path for proper URL construction
		if post.Image != "" {
			if strings.HasPrefix(post.Image, "./uploads") {
				post.Image = strings.Replace(post.Image, "./uploads", "/uploads", 1)
			}
		}

		// Fix avatar path as well
		if post.Avatar != nil && *post.Avatar != "" {
			if strings.HasPrefix(*post.Avatar, "./uploads") {
				fixed := strings.Replace(*post.Avatar, "./uploads", "/uploads", 1)
				post.Avatar = &fixed
			}
		}

		posts = append(posts, post)
	}

	return posts, nil
}
