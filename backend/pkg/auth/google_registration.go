package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"socialNetwork/pkg/db"
	ws "socialNetwork/pkg/websocket"
	"strings"
	"time"
	"unicode/utf8"
)

// GoogleRegistration completes required profile fields without inventing a birth date.
func GoogleRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	c, enabled := loadGoogleConfig()
	if !enabled {
		http.Error(w, "Google sign-in is currently unavailable", 503)
		return
	}
	f, key, err := readGoogleFlow(r, true)
	if err != nil {
		googleJSON(w, 401, map[string]string{"error": "Your Google sign-in has expired. Start again."})
		return
	}
	if r.Method == http.MethodGet {
		googleJSON(w, 200, map[string]string{"email": f.Email, "firstName": f.First, "lastName": f.Last})
		return
	}
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "JSON required", 415)
		return
	}
	var input struct {
		First    string `json:"firstName"`
		Last     string `json:"lastName"`
		DOB      string `json:"dob"`
		Nickname string `json:"nickname"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(new(interface{})) != io.EOF {
		googleJSON(w, 400, map[string]string{"error": "Invalid profile fields"})
		return
	}
	input.First, input.Last, input.Nickname = strings.TrimSpace(input.First), strings.TrimSpace(input.Last), strings.TrimSpace(input.Nickname)
	dob, err := time.Parse("2006-01-02", input.DOB)
	if err != nil || dob.After(time.Now()) || dob.Year() < 1900 || input.First == "" || input.Last == "" || utf8.RuneCountInString(input.First) > 100 || utf8.RuneCountInString(input.Last) > 100 || utf8.RuneCountInString(input.Nickname) > 100 {
		googleJSON(w, 400, map[string]string{"error": "Enter your name and a valid date of birth"})
		return
	}
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		googleJSON(w, 500, map[string]string{"error": "Unable to create account"})
		return
	}
	defer tx.Rollback()
	result, err := tx.Exec(`DELETE FROM oauth_flows WHERE token_hash=? AND intent='register' AND expires_at>?`, key, time.Now().Unix())
	if err != nil {
		googleJSON(w, 500, map[string]string{"error": "Unable to create account"})
		return
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		googleJSON(w, 401, map[string]string{"error": "Your Google sign-in has expired. Start again."})
		return
	}
	// Recheck within the write transaction to avoid email races with normal signup.
	var exists bool
	if tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email)=LOWER(?))`, f.Email).Scan(&exists) != nil {
		googleJSON(w, 500, map[string]string{"error": "Unable to create account"})
		return
	}
	if exists {
		googleJSON(w, 409, map[string]string{"error": "This email already has an account. Sign in with your password and connect Google in Settings."})
		return
	}
	result, err = tx.Exec(`INSERT INTO user_ids DEFAULT VALUES`)
	if err != nil {
		googleJSON(w, 500, map[string]string{"error": "Unable to create account"})
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		googleJSON(w, 500, map[string]string{"error": "Unable to create account"})
		return
	}
	_, err = tx.Exec(`INSERT INTO users(id,email,password,first_name,last_name,dob,nickname,isprivate) VALUES (?,?,'',?,?,?,NULLIF(?,''),1)`, id, f.Email, input.First, input.Last, input.DOB, input.Nickname)
	if err == nil {
		_, err = tx.Exec(`INSERT INTO oauth_identities(provider,subject,user_id) VALUES ('google',?,?)`, f.Subject, id)
	}
	if err != nil {
		googleJSON(w, 409, map[string]string{"error": "This Google account is already connected. Start sign-in again."})
		return
	}
	if tx.Commit() != nil {
		googleJSON(w, 500, map[string]string{"error": "Unable to create account"})
		return
	}
	googleCookie(w, c, "", -1)
	session, err := CreateSession(int(id))
	if err != nil {
		googleJSON(w, 500, map[string]string{"error": "Account created. Please sign in with Google again."})
		return
	}
	if _, err = db.DBInstance.DB.Exec(`UPDATE sessions SET auth_provider='google' WHERE token=?`, session.ID); err != nil {
		InvalidateSession(session.ID)
		googleJSON(w, 500, map[string]string{"error": "Account created. Please sign in with Google again."})
		return
	}
	SetSessionCookie(w, session)
	ws.PublishChange("users")
	googleJSON(w, 201, map[string]bool{"success": true})
}
