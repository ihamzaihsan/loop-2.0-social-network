package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
	"strings"
)

func fixImagePath(image string) string { return strings.Replace(image, "./uploads", "/uploads", 1) }
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

func GetPostsByUserID(id int) ([]models.PostResponse, error) {
	posts, _, err := VisiblePosts(id, id, -1, 0, "")
	return posts, err
}
func GetPostsByUserIDWithPrivacy(target, viewer int) ([]models.PostResponse, error) {
	posts, _, err := VisiblePosts(viewer, target, -1, 0, "")
	return posts, err
}
func GetUserPostsCount(id int) (int, error) { return GetUserPostsCountWithPrivacy(id, id) }
func GetUserPostsCountWithPrivacy(target, viewer int) (int, error) {
	_, total, err := VisiblePosts(viewer, target, 0, 0, "")
	return total, err
}
