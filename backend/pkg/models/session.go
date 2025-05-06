package models

import (
	"sync"
	"time"
)

type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	IsActive  bool      `json:"is_active"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionStore struct {
	Sessions     sync.Map // Maps sessionID to *Session
	UserSessions sync.Map // Maps userID to sessionID
}
