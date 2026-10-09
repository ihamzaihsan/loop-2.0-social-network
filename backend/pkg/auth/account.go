package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"socialNetwork/pkg/db"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ValidEmail(value string) bool {
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address == value && len(value) <= 254 && strings.Contains(value, "@")
}
func ValidPassword(value string) bool { return len(value) >= 8 && len(value) <= 72 }

func IsModerator(userID int) bool {
	var active, moderator bool
	if db.DBInstance.DB.QueryRow(`SELECT is_suspended=0,is_moderator FROM users WHERE id=?`, userID).Scan(&active, &moderator) != nil || !active {
		return false
	}
	if moderator {
		return true
	}
	for _, candidate := range strings.Split(os.Getenv("MODERATOR_USER_IDS"), ",") {
		if strings.TrimSpace(candidate) == strconv.Itoa(userID) {
			return true
		}
	}
	return false
}

type throttleEntry struct {
	count int
	until time.Time
}

var recoveryThrottle = struct {
	sync.Mutex
	entries map[string]throttleEntry
}{entries: make(map[string]throttleEntry)}

func AllowRecovery(key string) bool {
	recoveryThrottle.Lock()
	defer recoveryThrottle.Unlock()
	now := time.Now()
	for k, e := range recoveryThrottle.entries {
		if now.After(e.until) {
			delete(recoveryThrottle.entries, k)
		}
	}
	e := recoveryThrottle.entries[key]
	if e.until.IsZero() {
		e.until = now.Add(15 * time.Minute)
	}
	if e.count >= 5 {
		return false
	}
	e.count++
	recoveryThrottle.entries[key] = e
	return true
}

func smtpReset(recipient, link string) error {
	host, port, from := os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"), os.Getenv("SMTP_FROM")
	if host == "" || !ValidEmail(from) {
		return errors.New("password recovery email delivery is not configured")
	}
	if port == "" {
		port = "587"
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	if port == "465" {
		conn = tls.Client(conn, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if port != "465" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
				return err
			}
		} else if os.Getenv("SMTP_ALLOW_INSECURE") != "true" {
			return errors.New("SMTP requires TLS")
		}
	}
	if username := os.Getenv("SMTP_USERNAME"); username != "" {
		if err = client.Auth(smtp.PlainAuth("", username, os.Getenv("SMTP_PASSWORD"), host)); err != nil {
			return err
		}
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	if err = client.Rcpt(recipient); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "From: %s\r\nTo: %s\r\nSubject: Reset your Loop password\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nOpen this link to reset your password (expires in 30 minutes):\r\n%s\r\n\r\nIf you did not request this, ignore this email.\r\n", from, recipient, link)
	if err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	host := ClientIP(r)
	if !AllowRecovery("forgot:" + host) {
		http.Error(w, "Try again in 15 minutes", 429)
		return
	}
	var input struct {
		Email string `json:"email"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input) != nil || !ValidEmail(strings.TrimSpace(input.Email)) {
		http.Error(w, "Enter a valid email", 400)
		return
	}
	if os.Getenv("SMTP_HOST") == "" || !ValidEmail(os.Getenv("SMTP_FROM")) {
		http.Error(w, "Password recovery email delivery is not configured. Contact the site administrator.", 503)
		return
	}
	var id int
	email := strings.TrimSpace(input.Email)
	err := db.DBInstance.DB.QueryRow(`SELECT id FROM users WHERE email=? AND is_suspended=0`, email).Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Recovery unavailable", 500)
		return
	}
	if err == nil {
		bytes := make([]byte, 32)
		if _, err = rand.Read(bytes); err != nil {
			http.Error(w, "Unable to create reset link", 500)
			return
		}
		token := hex.EncodeToString(bytes)
		hash := sha256.Sum256([]byte(token))
		digest := hex.EncodeToString(hash[:])
		tx, e := db.DBInstance.DB.Begin()
		if e != nil {
			http.Error(w, "Recovery unavailable", 500)
			return
		}
		defer tx.Rollback()
		if _, e = tx.Exec(`DELETE FROM password_resets WHERE user_id=? OR expires_at < ?`, id, time.Now()); e == nil {
			_, e = tx.Exec(`INSERT INTO password_resets(token_hash,user_id,expires_at) VALUES (?,?,?)`, digest, id, time.Now().Add(30*time.Minute))
		}
		if e != nil || tx.Commit() != nil {
			http.Error(w, "Recovery unavailable", 500)
			return
		}
		base := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")
		if base == "" {
			base = "http://localhost:3000"
		}
		if e = smtpReset(email, base+"/reset-password?token="+token); e != nil {
			db.DBInstance.DB.Exec(`DELETE FROM password_resets WHERE token_hash=?`, digest)
			http.Error(w, "Email delivery failed. Please try again later.", 503)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "If an eligible account exists, a password reset link has been sent."})
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	host := ClientIP(r)
	if !AllowRecovery("reset:" + host) {
		http.Error(w, "Try again in 15 minutes", 429)
		return
	}
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input) != nil || len(input.Token) != 64 || !ValidPassword(input.Password) {
		http.Error(w, "Use a valid reset link and a password of 8–72 bytes", 400)
		return
	}
	digest := sha256.Sum256([]byte(input.Token))
	hash := hex.EncodeToString(digest[:])
	password, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Unable to reset password", 500)
		return
	}
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		http.Error(w, "Recovery unavailable", 500)
		return
	}
	defer tx.Rollback()
	var id int
	if tx.QueryRow(`SELECT user_id FROM password_resets WHERE token_hash=? AND expires_at>?`, hash, time.Now()).Scan(&id) != nil {
		http.Error(w, "Reset link is invalid or expired", 400)
		return
	}
	if _, err = tx.Exec(`UPDATE users SET password=? WHERE id=?`, password, id); err != nil {
		http.Error(w, "Recovery unavailable", 500)
		return
	}
	if _, err = tx.Exec(`DELETE FROM password_resets WHERE user_id=?`, id); err != nil {
		http.Error(w, "Recovery unavailable", 500)
		return
	}
	if err = tx.Commit(); err != nil {
		http.Error(w, "Recovery unavailable", 500)
		return
	}
	RevokeUserSessions(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
