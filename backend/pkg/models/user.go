package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	DOB       time.Time `json:"dob"`
	Avatar    *string   `json:"avatar"`
	Nickname  *string   `json:"nickname"`
	AboutMe   *string   `json:"aboutMe"`
	IsPublic bool      `json:"isPublic"`
	CreatedAt time.Time `json:"createdAt"`
}
