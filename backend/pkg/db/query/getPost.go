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

// GetPostsByUserIDWithPrivacy fetches posts for a user with privacy filtering
// viewerID is the ID of the user viewing the posts
// targetUserID is the ID of the user whose posts are being viewed
func GetPostsByUserIDWithPrivacy(targetUserID, viewerID int) ([]models.PostResponse, error) {
	log.Printf("Fetching posts for user ID: %d, viewer ID: %d", targetUserID, viewerID)

	// If viewer is the same as target user, show all posts
	if targetUserID == viewerID {
		return GetPostsByUserID(targetUserID)
	}

	// Check if viewer is following the target user (with accepted status)
	isFollowing, err := CheckIfFollowing(uint(viewerID), uint(targetUserID))
	if err != nil {
		log.Printf("Error checking follow status: %v", err)
		isFollowing = false
	}

	// Check if viewer and target user are mutual friends
	isMutualFollowing, err := CheckIfMutualFollowing(uint(viewerID), uint(targetUserID))
	if err != nil {
		log.Printf("Error checking mutual follow status: %v", err)
		isMutualFollowing = false
	}

	// Build query based on privacy rules
	var query string
	var args []interface{}

	if isMutualFollowing {
		// If mutual friends, show public, friends-only, and private posts where viewer is explicitly allowed
		query = `
			SELECT p.id, p.user_id, p.content, p.image, p.privacy, p.created_at,
				   u.first_name, u.last_name, u.nickname, u.avatar
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.user_id = ? AND (
				p.privacy = 'public'
				OR p.privacy = 'friends'
				OR (p.privacy = 'private' AND EXISTS (
					SELECT 1 FROM post_viewers
					WHERE post_id = p.id AND viewer_id = ?
				))
			)
			ORDER BY p.created_at DESC
		`
		args = []interface{}{targetUserID, viewerID}
	} else if isFollowing {
		// If following but not mutual friends, show public and private posts where viewer is explicitly allowed (no friends-only)
		query = `
			SELECT p.id, p.user_id, p.content, p.image, p.privacy, p.created_at,
				   u.first_name, u.last_name, u.nickname, u.avatar
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.user_id = ? AND (
				p.privacy = 'public'
				OR (p.privacy = 'private' AND EXISTS (
					SELECT 1 FROM post_viewers
					WHERE post_id = p.id AND viewer_id = ?
				))
			)
			ORDER BY p.created_at DESC
		`
		args = []interface{}{targetUserID, viewerID}
	} else {
		// If not following, only show public posts
		query = `
			SELECT p.id, p.user_id, p.content, p.image, p.privacy, p.created_at,
				   u.first_name, u.last_name, u.nickname, u.avatar
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.user_id = ? AND p.privacy = 'public'
			ORDER BY p.created_at DESC
		`
		args = []interface{}{targetUserID}
	}

	rows, err := db.DBInstance.DB.Query(query, args...)
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

	log.Printf("Found %d visible posts for user ID: %d (viewer: %d)", len(posts), targetUserID, viewerID)
	return posts, nil
}

// CheckIfMutualFollowing checks if two users have a mutual follow relationship (are friends)
func CheckIfMutualFollowing(userID1, userID2 uint) (bool, error) {
	var count int
	err := db.DBInstance.DB.QueryRow(`
		SELECT COUNT(*) FROM followers f1
		JOIN followers f2 ON f1.follower_id = f2.following_id AND f1.following_id = f2.follower_id
		WHERE f1.follower_id = ? AND f1.following_id = ?
		AND f1.status = 'accept' AND f2.status = 'accept'
	`, userID1, userID2).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserPostsCountWithPrivacy counts posts for a user with privacy filtering
// viewerID is the ID of the user viewing the posts
// targetUserID is the ID of the user whose posts are being counted
func GetUserPostsCountWithPrivacy(targetUserID, viewerID int) (int, error) {
	log.Printf("Counting posts for user ID: %d, viewer ID: %d", targetUserID, viewerID)

	// If viewer is the same as target user, count all posts
	if targetUserID == viewerID {
		return GetUserPostsCount(targetUserID)
	}

	// Check if viewer is following the target user (with accepted status)
	isFollowing, err := CheckIfFollowing(uint(viewerID), uint(targetUserID))
	if err != nil {
		log.Printf("Error checking follow status: %v", err)
		isFollowing = false
	}

	// Check if viewer and target user are mutual friends
	isMutualFollowing, err := CheckIfMutualFollowing(uint(viewerID), uint(targetUserID))
	if err != nil {
		log.Printf("Error checking mutual follow status: %v", err)
		isMutualFollowing = false
	}

	// Build count query based on privacy rules
	var countQuery string
	var args []interface{}

	if isMutualFollowing {
		// If mutual friends, count public, friends-only, and private posts where viewer is explicitly allowed
		countQuery = `
			SELECT COUNT(*) FROM posts p
			WHERE p.user_id = ? AND (
				p.privacy = 'public'
				OR p.privacy = 'friends'
				OR (p.privacy = 'private' AND EXISTS (
					SELECT 1 FROM post_viewers
					WHERE post_id = p.id AND viewer_id = ?
				))
			)
		`
		args = []interface{}{targetUserID, viewerID}
	} else if isFollowing {
		// If following but not mutual friends, count public and private posts where viewer is explicitly allowed (no friends-only)
		countQuery = `
			SELECT COUNT(*) FROM posts p
			WHERE p.user_id = ? AND (
				p.privacy = 'public'
				OR (p.privacy = 'private' AND EXISTS (
					SELECT 1 FROM post_viewers
					WHERE post_id = p.id AND viewer_id = ?
				))
			)
		`
		args = []interface{}{targetUserID, viewerID}
	} else {
		// If not following, only count public posts
		countQuery = `
			SELECT COUNT(*) FROM posts p
			WHERE p.user_id = ? AND p.privacy = 'public'
		`
		args = []interface{}{targetUserID}
	}

	var count int
	err = db.DBInstance.DB.QueryRow(countQuery, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	log.Printf("Found %d visible posts for user ID: %d (viewer: %d)", count, targetUserID, viewerID)
	return count, nil
}
