package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func CreatePostQuery(userID int, request models.PostRequest) (*models.Post, error) {
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO posts (user_id,content,image,privacy) VALUES (?,?,?,?)`
	result, err := tx.Exec(query, userID, request.Content, request.Image, request.Privacy)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if request.Privacy == "private" {
		viewerQuery := `INSERT INTO post_viewers (post_id, viewer_id) VALUES (?, ?)`
		for _, viewerID := range request.ViewerIDs {
			_, err := tx.Exec(viewerQuery, postID, viewerID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}
	
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}
	return GetPostByIDQuery(int(postID))
}
