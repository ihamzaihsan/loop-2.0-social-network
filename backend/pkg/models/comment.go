package models

import "time"

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"postId"`
	UserID    int       `json:"userId"`
	Content   string    `json:"content"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"createdAt"`
}

type CommentRequest struct {
	Content string `json:"content"`
	Image   string `json:"image"`
}

type CommentResponse struct {
	Comment
	Author User `json:"author"`
}
