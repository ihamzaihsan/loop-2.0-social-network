package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func CreateComment(userID, postID int, request models.CommentRequest) (*models.CommentResponse, error) {
	// Add more detailed logging
	log.Printf("Creating comment in database: postID=%d, userID=%d, content=%s", postID, userID, request.Content)

	// Start a transaction to ensure data consistency
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return nil, err
	}

	var result sql.Result
	var commentID int64

	// Try to insert with image field first
	query := `INSERT INTO comments (post_id, user_id, content, image, created_at) 
              VALUES (?, ?, ?, ?, datetime('now'))`

	result, err = tx.Exec(query, postID, userID, request.Content, request.Image)

	if err != nil {
		// If there's an error, it might be because the image column doesn't exist yet
		log.Printf("Error with first insert attempt (might be missing image column): %v", err)

		// Try again without the image field
		query = `INSERT INTO comments (post_id, user_id, content, created_at) 
                VALUES (?, ?, ?, datetime('now'))`

		result, err = tx.Exec(query, postID, userID, request.Content)

		if err != nil {
			tx.Rollback()
			log.Printf("Error inserting comment (second attempt): %v", err)
			return nil, err
		}
	}

	// Get the last insert ID
	commentID, err = result.LastInsertId()
	if err != nil {
		tx.Rollback()
		log.Printf("Error getting last insert ID: %v", err)
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, err
	}

	// Log successful insertion
	log.Printf("Successfully inserted comment with ID %d", commentID)

	// Fetch the newly created comment with author information
	var commentResponse models.CommentResponse

	// Try to fetch with image field
	err = db.DBInstance.DB.QueryRow(`
		SELECT 
			c.id, c.post_id, c.user_id, c.content, IFNULL(c.image, '') as image, c.created_at,
			u.first_name, u.last_name, u.nickname, u.avatar
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`, commentID).Scan(
		&commentResponse.ID,
		&commentResponse.PostID,
		&commentResponse.UserID,
		&commentResponse.Content,
		&commentResponse.Image,
		&commentResponse.CreatedAt,
		&commentResponse.Author.FirstName,
		&commentResponse.Author.LastName,
		&commentResponse.Author.Nickname,
		&commentResponse.Author.Avatar,
	)

	if err != nil {
		// If there's an error, it might be because the image column doesn't exist
		log.Printf("Error fetching comment with image: %v", err)

		// Try again without the image field
		err = db.DBInstance.DB.QueryRow(`
			SELECT 
				c.id, c.post_id, c.user_id, c.content, c.created_at,
				u.first_name, u.last_name, u.nickname, u.avatar
			FROM comments c
			JOIN users u ON c.user_id = u.id
			WHERE c.id = ?
		`, commentID).Scan(
			&commentResponse.ID,
			&commentResponse.PostID,
			&commentResponse.UserID,
			&commentResponse.Content,
			&commentResponse.CreatedAt,
			&commentResponse.Author.FirstName,
			&commentResponse.Author.LastName,
			&commentResponse.Author.Nickname,
			&commentResponse.Author.Avatar,
		)

		if err != nil {
			log.Printf("Error fetching created comment (second attempt): %v", err)
			return nil, err
		}
	}

	return &commentResponse, nil
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
