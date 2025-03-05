package models

import (
	"sync"
	"time"
)

type Session struct {
	ID        string
	UserID    int
	IsActive  bool
	CreatedAt time.Time
}

type SessionStore struct {
	Sessions sync.Map
}
