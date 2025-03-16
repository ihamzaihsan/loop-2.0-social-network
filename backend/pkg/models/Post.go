package models

import "time"

type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Content   string    `json:"content"`
	Image     string    `json:"image"`
	Privacy   string    `json:"privacy"`
	CreatedAt time.Time `json:"createdAt"`
}

type PostRequest struct {
	Content   string `json:"content"`
	Image     string `json:"image"`
	Privacy   string `json:"privacy"`
	ViewerIDs []int  `json:"viewerIds,omitempty"`
}

type PostResponse struct {
	Post
	Author    User `json:"author"`
	LikeCount int  `json:"likeCount"`
	IsLiked   bool `json:"isLiked"`
}
