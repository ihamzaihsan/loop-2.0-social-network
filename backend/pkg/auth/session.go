package auth

import (
	"net/http"
	"socialNetwork/pkg/models"
	"time"

	"github.com/google/uuid"
)

var SessionStore = models.SessionStore{}

func CreateSession(userID int) (*models.Session, error) {
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	SessionStore.Sessions.Store(session.ID, session)
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
