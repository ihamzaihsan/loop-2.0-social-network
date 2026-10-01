package models

import "time"

// Notification represents a notification in the system
type Notification struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	FromUserID int       `json:"from_user_id,omitempty"`
	Type       string    `json:"type"`
	RelatedID  int       `json:"related_id,omitempty"`
	Content    string    `json:"content,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`

	// Additional fields for frontend display
	SenderName   string   `json:"sender_name,omitempty"`
	SenderAvatar string   `json:"sender_avatar,omitempty"`
	GroupTitle   string   `json:"group_title,omitempty"`
	Actions      []string `json:"actions,omitempty"`
}

// NotificationResponse represents a response containing notification data
type NotificationResponse struct {
	Success       bool           `json:"success"`
	Notifications []Notification `json:"notifications"`
	UnreadCount   int            `json:"unread_count"`
}

// NotificationCountResponse represents the unread notification count
type NotificationCountResponse struct {
	Success     bool `json:"success"`
	UnreadCount int  `json:"unread_count"`
}

// NotificationAction represents an action to take on a notification
type NotificationAction struct {
	Action string `json:"action"`
}
