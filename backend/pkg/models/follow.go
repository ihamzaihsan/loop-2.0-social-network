package models

import "time"

type Follow struct {
	ID         uint      `json:"id,omitempty"`
	FollowerID uint      `json:"follower_id,omitempty"`
	FollowedID uint      `json:"followed_id,omitempty"`
	Status     string    `json:"status,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	Username   string    `json:"username,omitempty"`
	Avatar     string    `json:"avatar,omitempty"`
}

type FollowRequest struct {
	ID         int    `json:"id"`
	FollowedID int    `json:"followed_id"`
	Username   string `json:"username"`
	Avatar     string `json:"avatar"`
}

type FollowResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type CloseFriends struct {
	Usernames []string `json:"selectedUsers"`
}
