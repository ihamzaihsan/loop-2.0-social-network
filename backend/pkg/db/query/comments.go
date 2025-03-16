package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func CreateComment(userID, postID int, request models.CommentRequest) (*models.Comment, error) {
	query := `INSERT INTO comments (post_id, user_id, content, image) VALUES (?, ?, ?, ?)`
	result, err := db.DBInstance.DB.Exec(query, postID, userID, request.Content, request.Image)
	if err != nil {
		return nil, err
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetCommentByID(int(commentID))
}

func GetCommentByID(commentID int) (*models.Comment, error) {
	query := `SELECT id, post_id, user_id, content, image, created_at 
              FROM comments 
              WHERE id = ?`

	var comment models.Comment
	err := db.DBInstance.DB.QueryRow(query, commentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.Image,
		&comment.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func GetCommentsByPostID(postID int) ([]models.CommentResponse, error) {
	query := `
        SELECT 
            c.id, c.post_id, c.user_id, c.content, c.image, c.created_at,
            u.first_name, u.last_name, u.nickname, u.avatar
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at DESC
    `

	rows, err := db.DBInstance.DB.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.CommentResponse
	for rows.Next() {
		var comment models.CommentResponse
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.Image, &comment.CreatedAt,
			&comment.Author.FirstName, &comment.Author.LastName, &comment.Author.Nickname, &comment.Author.Avatar,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
