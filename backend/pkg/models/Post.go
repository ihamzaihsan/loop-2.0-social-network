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
