package auth

import (
	"net/http"
	"socialNetwork/pkg/models"
	"time"

	"github.com/google/uuid"
)

var SessionStore = models.SessionStore{}

func CreateSession(userID int) (*models.Session, error) {
	// Check if user already has an active session
	if oldSessionID, exists := SessionStore.UserSessions.Load(userID); exists {
		// Invalidate the old session
		if oldSession, ok := SessionStore.Sessions.Load(oldSessionID); ok {
			if session, ok := oldSession.(*models.Session); ok {
				session.IsActive = false
			}
		}
		// Remove the old session from the store
		SessionStore.Sessions.Delete(oldSessionID)
	}

	// Create new session
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	SessionStore.Sessions.Store(session.ID, session)
	
	// Update the user's active session mapping
	SessionStore.UserSessions.Store(userID, session.ID)
	
	return session, nil
}

func SetSessionCookie(w http.ResponseWriter, Session *models.Session) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    Session.ID,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)
}

func GetSessionFromCookie(r *http.Request) (*models.Session, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}
	sessionID := cookie.Value
	if sessionInterface, ok := SessionStore.Sessions.Load(sessionID); ok {
		if session, ok := sessionInterface.(*models.Session); ok {
			return session, nil
		}
	}

	return nil, nil
}

func GetUserID(r *http.Request) (int, error) {
	session, err := GetSessionFromCookie(r)
	if err != nil {
		return 0, err
	}
	return session.UserID, nil
}

// Add this function to invalidate a session
func InvalidateSession(sessionID string) {
	if sessionInterface, ok := SessionStore.Sessions.Load(sessionID); ok {
		if session, ok := sessionInterface.(*models.Session); ok {
			// Remove from user sessions map
			SessionStore.UserSessions.Delete(session.UserID)
			// Delete from sessions map
			SessionStore.Sessions.Delete(sessionID)
		}
	}
}
