package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func CreateComment(userID, postID int, request models.CommentRequest) (*models.CommentResponse, error) {
	// Add more detailed logging
	log.Printf("Creating comment in database: postID=%d, userID=%d, content=%s, image=%s", 
		postID, userID, request.Content, request.Image)

	// Start a transaction to ensure data consistency
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return nil, err
	}

	var result sql.Result
	var commentID int64

	// Check if the image column exists in the comments table
	var hasImageColumn bool
	err = db.DBInstance.DB.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM pragma_table_info('comments') 
		WHERE name = 'image'
	`).Scan(&hasImageColumn)
	
	if err != nil {
		log.Printf("Error checking for image column: %v", err)
		tx.Rollback()
		return nil, err
	}

	log.Printf("Comments table has image column: %v", hasImageColumn)

	if hasImageColumn {
		// Insert with image field
		query := `INSERT INTO comments (post_id, user_id, content, image, created_at) 
				VALUES (?, ?, ?, ?, datetime('now'))`

		result, err = tx.Exec(query, postID, userID, request.Content, request.Image)
		if err != nil {
			log.Printf("Error inserting comment with image: %v", err)
			tx.Rollback()
			return nil, err
		}
	} else {
		// Insert without image field
		query := `INSERT INTO comments (post_id, user_id, content, created_at) 
				VALUES (?, ?, ?, datetime('now'))`

		result, err = tx.Exec(query, postID, userID, request.Content)
		if err != nil {
			log.Printf("Error inserting comment without image: %v", err)
			tx.Rollback()
			return nil, err
		}
		
		// If we successfully inserted without the image column, we should add the column
		_, err = tx.Exec(`ALTER TABLE comments ADD COLUMN image TEXT`)
		if err != nil {
			log.Printf("Error adding image column to comments table: %v", err)
			// Continue anyway, as the comment was inserted successfully
		} else {
			log.Printf("Added image column to comments table")
			
			// If we have an image, update the row
			if request.Image != "" {
				_, err = tx.Exec(`UPDATE comments SET image = ? WHERE id = last_insert_rowid()`, request.Image)
				if err != nil {
					log.Printf("Error updating comment with image: %v", err)
					// Continue anyway, as the comment was inserted successfully
				}
			}
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
	if hasImageColumn {
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
	} else {
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
		
		// Set the image field manually if we have one
		if request.Image != "" {
			commentResponse.Image = request.Image
		}
	}

	if err != nil {
		log.Printf("Error fetching created comment: %v", err)
		return nil, err
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
