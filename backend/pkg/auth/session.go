package auth

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/events"
	"socialNetwork/pkg/models"
	"strings"
	"sync"
	"time"
)

// SQLite is the authority for session state; never cache revocation or expiry.
var sessionWrites sync.Mutex

func InitSessions() error {
	_, err := db.DBInstance.DB.Exec(`UPDATE sessions SET is_active=0 WHERE julianday(expires_at)<=julianday(?)`, time.Now().UTC())
	return err
}

// Rotation and insertion commit together, including concurrent logins.
func CreateSession(userID int) (*models.Session, error) {
	sessionWrites.Lock()
	defer sessionWrites.Unlock()
	now := time.Now().UTC()
	session := &models.Session{ID: uuid.NewString(), UserID: userID, IsActive: true, CreatedAt: now, ExpiresAt: now.Add(24 * time.Hour)}
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE sessions SET is_active=0 WHERE user_id=? AND is_active=1`, userID)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(`INSERT INTO sessions(token,user_id,is_active,expires_at,created_at) VALUES(?,?,1,?,?)`, session.ID, userID, session.ExpiresAt, now); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	if count, _ := result.RowsAffected(); count > 0 {
		events.Publish(events.Event{Type: events.SessionInvalidated, UserID: userID})
	}
	return session, nil
}

func sessionCookie() *http.Cookie {
	return &http.Cookie{Name: "session_token", Path: "/", HttpOnly: true, Secure: strings.HasPrefix(os.Getenv("FRONTEND_URL"), "https://"), SameSite: http.SameSiteLaxMode}
}

func SetSessionCookie(w http.ResponseWriter, session *models.Session) {
	cookie := sessionCookie()
	cookie.Value, cookie.Expires, cookie.MaxAge = session.ID, session.ExpiresAt, 86400
	http.SetCookie(w, cookie)
}

func ClearSessionCookie(w http.ResponseWriter) {
	cookie := sessionCookie()
	cookie.Expires, cookie.MaxAge = time.Unix(1, 0), -1
	http.SetCookie(w, cookie)
}

func GetSessionFromCookie(r *http.Request) (*models.Session, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}
	if _, err = uuid.Parse(cookie.Value); err != nil || len(cookie.Value) != 36 {
		return nil, nil
	}
	var session models.Session
	err = db.DBInstance.DB.QueryRow(`SELECT s.token,s.user_id,s.is_active,s.expires_at,s.created_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token=? AND s.is_active=1 AND julianday(s.expires_at)>julianday(?) AND u.is_suspended=0`, cookie.Value, time.Now().UTC()).Scan(&session.ID, &session.UserID, &session.IsActive, &session.ExpiresAt, &session.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func GetUserID(r *http.Request) (int, error) {
	session, err := GetSessionFromCookie(r)
	if err != nil {
		return 0, err
	}
	if session == nil {
		return 0, errors.New("no active session")
	}
	return session.UserID, nil
}

func InvalidateSession(token string) {
	sessionWrites.Lock()
	defer sessionWrites.Unlock()
	var userID int
	err := db.DBInstance.DB.QueryRow(`UPDATE sessions SET is_active=0 WHERE token=? AND is_active=1 RETURNING user_id`, token).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return
	}
	if err != nil {
		log.Printf("Session revocation failed: %v", err)
		return
	}
	events.Publish(events.Event{Type: events.SessionInvalidated, UserID: userID})
}

func RevokeUserSessions(userID int) {
	sessionWrites.Lock()
	defer sessionWrites.Unlock()
	if _, err := db.DBInstance.DB.Exec(`UPDATE sessions SET is_active=0 WHERE user_id=?`, userID); err != nil {
		log.Printf("Session revocation failed: %v", err)
		return
	}
	events.Publish(events.Event{Type: events.SessionInvalidated, UserID: userID})
}

func CleanupExpiredSessions() {
	sessionWrites.Lock()
	defer sessionWrites.Unlock()
	if _, err := db.DBInstance.DB.Exec(`UPDATE sessions SET is_active=0 WHERE julianday(expires_at)<=julianday(?) AND is_active=1`, time.Now().UTC()); err != nil {
		log.Printf("Session cleanup failed: %v", err)
	}
}
