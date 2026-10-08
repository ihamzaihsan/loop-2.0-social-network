package auth

import (
	"database/sql"
	"log"
	"net/http"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/events"
	"socialNetwork/pkg/models"
	"time"

	"github.com/google/uuid"
)

var SessionStore = models.SessionStore{}

// InitSessionStore loads active sessions from database into memory
func InitSessionStore() error {
	log.Println("Initializing session store from database...")

	rows, err := db.DBInstance.DB.Query(`
		SELECT token, user_id, is_active, expires_at, created_at 
		FROM sessions 
		WHERE is_active = 1 AND expires_at > datetime('now')
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var session models.Session
		if err := rows.Scan(&session.ID, &session.UserID, &session.IsActive, &session.ExpiresAt, &session.CreatedAt); err != nil {
			log.Printf("Error scanning session row: %v", err)
			continue
		}

		// Store in memory
		SessionStore.Sessions.Store(session.ID, &session)
		SessionStore.UserSessions.Store(session.UserID, session.ID)
		count++
	}

	log.Printf("Loaded %d active sessions into memory", count)
	return nil
}

// CreateSession creates a new session for a user
func CreateSession(userID int) (*models.Session, error) {
	// Check if user already has an active session
	if oldSessionID, exists := SessionStore.UserSessions.Load(userID); exists {
		// Invalidate the old session
		InvalidateSession(oldSessionID.(string))
	}

	// Create new session with 24-hour expiration
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		IsActive:  true,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	// Store in database
	_, err := db.DBInstance.DB.Exec(`
		INSERT INTO sessions (token, user_id, is_active, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, session.ID, session.UserID, session.IsActive, session.ExpiresAt, session.CreatedAt)

	if err != nil {
		log.Printf("Error storing session in database: %v", err)
		return nil, err
	}

	// Store in memory
	SessionStore.Sessions.Store(session.ID, session)
	SessionStore.UserSessions.Store(userID, session.ID)

	return session, nil
}

// SetSessionCookie sets the session cookie in the response
func SetSessionCookie(w http.ResponseWriter, session *models.Session) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    session.ID,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	}
	http.SetCookie(w, cookie)
}

// GetSessionFromCookie retrieves a session from a cookie
func GetSessionFromCookie(r *http.Request) (*models.Session, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}

	sessionID := cookie.Value

	// Try to get from memory first
	if sessionInterface, ok := SessionStore.Sessions.Load(sessionID); ok {
		if session, ok := sessionInterface.(*models.Session); ok {
			// Check if session is expired
			if time.Now().After(session.ExpiresAt) {
				InvalidateSession(sessionID)
				return nil, nil
			}
			return session, nil
		}
	}

	// If not in memory, try database
	var session models.Session
	err = db.DBInstance.DB.QueryRow(`
		SELECT token, user_id, is_active, expires_at, created_at 
		FROM sessions 
		WHERE token = ? AND is_active = 1 AND expires_at > datetime('now')
	`, sessionID).Scan(&session.ID, &session.UserID, &session.IsActive, &session.ExpiresAt, &session.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Store in memory for future use
	SessionStore.Sessions.Store(session.ID, &session)
	SessionStore.UserSessions.Store(session.UserID, session.ID)

	return &session, nil
}

// GetUserID gets the user ID from a session
func GetUserID(r *http.Request) (int, error) {
	session, err := GetSessionFromCookie(r)
	if err != nil || session == nil {
		return 0, err
	}
	return session.UserID, nil
}

// InvalidateSession invalidates a session
func InvalidateSession(sessionID string) {
	// Get session from memory
	sessionInterface, ok := SessionStore.Sessions.Load(sessionID)
	if !ok {
		return
	}

	session, ok := sessionInterface.(*models.Session)
	if !ok {
		return
	}

	// Remove from user sessions map
	SessionStore.UserSessions.Delete(session.UserID)

	// Remove from sessions map
	SessionStore.Sessions.Delete(sessionID)

	// Update database
	_, err := db.DBInstance.DB.Exec(`
		UPDATE sessions 
		SET is_active = 0 
		WHERE token = ?
	`, sessionID)

	if err != nil {
		log.Printf("Error invalidating session in database: %v", err)
	}
	events.Publish(events.Event{Type: events.SessionInvalidated, UserID: session.UserID})
}

// CleanupExpiredSessions removes expired sessions from database and memory
func CleanupExpiredSessions() {
	log.Println("Cleaning up expired sessions...")

	// Update database
	result, err := db.DBInstance.DB.Exec(`
		UPDATE sessions 
		SET is_active = 0 
		WHERE expires_at <= datetime('now') AND is_active = 1
	`)

	if err != nil {
		log.Printf("Error cleaning up expired sessions in database: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Marked %d expired sessions as inactive in database", rowsAffected)

	// Clean memory
	// This is a bit tricky with sync.Map since we can't iterate and modify
	// We'll collect keys to delete first
	var keysToDelete []string

	SessionStore.Sessions.Range(func(key, value interface{}) bool {
		if session, ok := value.(*models.Session); ok {
			if time.Now().After(session.ExpiresAt) {
				keysToDelete = append(keysToDelete, key.(string))
				SessionStore.UserSessions.Delete(session.UserID)
			}
		}
		return true
	})

	// Now delete the collected keys
	for _, key := range keysToDelete {
		SessionStore.Sessions.Delete(key)
	}

	log.Printf("Removed %d expired sessions from memory", len(keysToDelete))
}
