package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"socialNetwork/pkg/db"
	"strings"
	"time"
)

const googleFlowCookie = "loop_google_flow"

var googleHTTPClient = &http.Client{Timeout: 10 * time.Second}

type googleConfig struct{ clientID, secret, redirectURI, frontend string }
type googleIdentity struct {
	Subject  string `json:"sub"`
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
	First    string `json:"given_name"`
	Last     string `json:"family_name"`
}
type googleFlow struct {
	Verifier, Intent, Subject, Email, First, Last string
	UserID                                        sql.NullInt64
}

func loadGoogleConfig() (googleConfig, bool) {
	c := googleConfig{os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_REDIRECT_URI"), strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")}
	if c.frontend == "" {
		c.frontend = "http://localhost:3000"
	}
	validURL := func(raw string) bool {
		u, err := url.Parse(raw)
		return err == nil && u.Host != "" && u.User == nil && u.RawQuery == "" && u.Fragment == "" && (u.Scheme == "https" || u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))
	}
	u, _ := url.Parse(c.redirectURI)
	return c, c.clientID != "" && c.secret != "" && validURL(c.frontend) && validURL(c.redirectURI) && u.Path == strings.TrimRight(os.Getenv("PUBLIC_API_PATH"), "/")+"/auth/google/callback"
}

func googleJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func googleCookie(w http.ResponseWriter, c googleConfig, value string, age int) {
	http.SetCookie(w, &http.Cookie{Name: googleFlowCookie, Value: value, Path: strings.TrimRight(os.Getenv("PUBLIC_API_PATH"), "/") + "/auth/google", HttpOnly: true, Secure: strings.HasPrefix(c.redirectURI, "https:"), SameSite: http.SameSiteLaxMode, MaxAge: age})
}

func randomGoogleToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
func googleTokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func readGoogleFlow(r *http.Request, registration bool) (googleFlow, string, error) {
	var f googleFlow
	cookie, err := r.Cookie(googleFlowCookie)
	if err != nil || len(cookie.Value) != 43 {
		return f, "", errors.New("missing Google flow")
	}
	key := googleTokenHash(cookie.Value)
	err = db.DBInstance.DB.QueryRow(`SELECT verifier,intent,user_id,subject,email,first_name,last_name FROM oauth_flows WHERE token_hash=? AND expires_at>?`, key, time.Now().Unix()).Scan(&f.Verifier, &f.Intent, &f.UserID, &f.Subject, &f.Email, &f.First, &f.Last)
	if err == nil && registration != (f.Intent == "register") {
		err = errors.New("incorrect Google flow stage")
	}
	return f, key, err
}

func GoogleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	_, enabled := loadGoogleConfig()
	googleJSON(w, 200, map[string]bool{"enabled": enabled})
}

func GoogleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	c, enabled := loadGoogleConfig()
	if !enabled {
		http.Error(w, "Google sign-in is currently unavailable", 503)
		return
	}
	intent := r.URL.Query().Get("intent")
	if intent == "" {
		intent = "signin"
	}
	var actor interface{}
	if intent == "link" || intent == "reauth" {
		session, err := GetSessionFromCookie(r)
		if err != nil || session == nil {
			http.Error(w, "Sign in before connecting Google", 401)
			return
		}
		actor = session.UserID
	} else if intent != "signin" {
		http.Error(w, "Invalid sign-in action", 400)
		return
	}
	state, err := randomGoogleToken()
	if err != nil {
		http.Error(w, "Unable to start sign-in", 500)
		return
	}
	verifier, err := randomGoogleToken()
	if err != nil {
		http.Error(w, "Unable to start sign-in", 500)
		return
	}
	db.DBInstance.DB.Exec(`DELETE FROM oauth_flows WHERE expires_at<=?`, time.Now().Unix())
	// Replace this browser's abandoned flow without keeping unnecessary state.
	if old, err := r.Cookie(googleFlowCookie); err == nil {
		db.DBInstance.DB.Exec(`DELETE FROM oauth_flows WHERE token_hash=?`, googleTokenHash(old.Value))
	}
	if _, err = db.DBInstance.DB.Exec(`INSERT INTO oauth_flows(token_hash,verifier,intent,user_id,expires_at) VALUES (?,?,?,?,?)`, googleTokenHash(state), verifier, intent, actor, time.Now().Add(10*time.Minute).Unix()); err != nil {
		http.Error(w, "Unable to start sign-in", 500)
		return
	}
	googleCookie(w, c, state, 600)
	challenge := sha256.Sum256([]byte(verifier))
	params := url.Values{"client_id": {c.clientID}, "redirect_uri": {c.redirectURI}, "response_type": {"code"}, "scope": {"openid email profile"}, "state": {state}, "code_challenge": {base64.RawURLEncoding.EncodeToString(challenge[:])}, "code_challenge_method": {"S256"}, "prompt": {"select_account"}}
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, "https://accounts.google.com/o/oauth2/v2/auth?"+params.Encode(), http.StatusSeeOther)
}

// Provider calls stay server-side. No browser-supplied email or decoded JWT is trusted.
func exchangeGoogleIdentity(r *http.Request, c googleConfig, code, verifier string) (googleIdentity, error) {
	var identity googleIdentity
	form := url.Values{"code": {code}, "client_id": {c.clientID}, "client_secret": {c.secret}, "redirect_uri": {c.redirectURI}, "grant_type": {"authorization_code"}, "code_verifier": {verifier}}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return identity, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := googleHTTPClient.Do(req)
	if err != nil {
		return identity, err
	}
	defer response.Body.Close()
	var token struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&token) != nil || token.AccessToken == "" || !strings.EqualFold(token.TokenType, "Bearer") {
		return identity, errors.New("Google rejected authorization")
	}
	req, err = http.NewRequestWithContext(r.Context(), http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return identity, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	profile, err := googleHTTPClient.Do(req)
	if err != nil {
		return identity, err
	}
	defer profile.Body.Close()
	if profile.StatusCode != 200 || json.NewDecoder(io.LimitReader(profile.Body, 1<<20)).Decode(&identity) != nil || identity.Subject == "" || len(identity.Subject) > 255 || !identity.Verified || !ValidEmail(identity.Email) {
		return identity, errors.New("Google identity unavailable")
	}
	return identity, nil
}

func googleRedirectError(w http.ResponseWriter, r *http.Request, c googleConfig, reason string, settings bool) {
	page := "/login"
	if settings {
		page = "/settings"
	}
	http.Redirect(w, r, c.frontend+page+"?google_error="+url.QueryEscape(reason), http.StatusSeeOther)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	c, enabled := loadGoogleConfig()
	if !enabled {
		http.Error(w, "Google sign-in is currently unavailable", 503)
		return
	}
	f, key, err := readGoogleFlow(r, false)
	cookie, cookieErr := r.Cookie(googleFlowCookie)
	state := r.URL.Query().Get("state")
	if err != nil || cookieErr != nil || len(state) != 43 || subtle.ConstantTimeCompare([]byte(state), []byte(cookie.Value)) != 1 {
		googleRedirectError(w, r, c, "expired", false)
		return
	}
	// Claim the callback once, including cancelled or failed provider requests.
	result, err := db.DBInstance.DB.Exec(`DELETE FROM oauth_flows WHERE token_hash=?`, key)
	if err != nil {
		googleRedirectError(w, r, c, "failed", false)
		return
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		googleRedirectError(w, r, c, "expired", false)
		return
	}
	googleCookie(w, c, "", -1)
	settings := f.Intent == "link" || f.Intent == "reauth"
	if r.URL.Query().Get("error") != "" || r.URL.Query().Get("code") == "" {
		googleRedirectError(w, r, c, "cancelled", settings)
		return
	}
	identity, err := exchangeGoogleIdentity(r, c, r.URL.Query().Get("code"), f.Verifier)
	if err != nil {
		googleRedirectError(w, r, c, "failed", settings)
		return
	}
	var userID int
	var suspended bool
	err = db.DBInstance.DB.QueryRow(`SELECT u.id,u.is_suspended FROM oauth_identities i JOIN users u ON u.id=i.user_id WHERE i.provider='google' AND i.subject=?`, identity.Subject).Scan(&userID, &suspended)
	if err != nil && err != sql.ErrNoRows {
		googleRedirectError(w, r, c, "failed", settings)
		return
	}
	if settings {
		session, sessionErr := GetSessionFromCookie(r)
		if sessionErr != nil || session == nil || !f.UserID.Valid || session.UserID != int(f.UserID.Int64) {
			googleRedirectError(w, r, c, "expired", true)
			return
		}
		if err == nil && userID != session.UserID {
			googleRedirectError(w, r, c, "linked_elsewhere", true)
			return
		}
		if f.Intent == "reauth" && err == sql.ErrNoRows {
			googleRedirectError(w, r, c, "wrong_account", true)
			return
		}
		if f.Intent == "link" {
			if err == sql.ErrNoRows {
				if _, err = db.DBInstance.DB.Exec(`INSERT INTO oauth_identities(provider,subject,user_id) VALUES ('google',?,?)`, identity.Subject, session.UserID); err != nil {
					googleRedirectError(w, r, c, "linked_elsewhere", true)
					return
				}
			}
			redirectGooglePage(w, r, c.frontend+"/settings?google_connected=1")
			return
		}
		googleLogin(w, r, c, userID, "/settings?google_connected=1")
		return
	}
	if err == nil {
		if suspended {
			googleRedirectError(w, r, c, "unavailable", false)
			return
		}
		googleLogin(w, r, c, userID, "/home")
		return
	}
	// Never link solely by email: a password-authenticated owner must opt in.
	var exists bool
	if db.DBInstance.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email=? COLLATE NOCASE)`, identity.Email).Scan(&exists) != nil {
		googleRedirectError(w, r, c, "failed", false)
		return
	}
	if exists {
		googleRedirectError(w, r, c, "existing_account", false)
		return
	}
	ticket, err := randomGoogleToken()
	if err != nil {
		googleRedirectError(w, r, c, "failed", false)
		return
	}
	_, err = db.DBInstance.DB.Exec(`INSERT INTO oauth_flows(token_hash,intent,subject,email,first_name,last_name,expires_at) VALUES (?,'register',?,?,?,?,?)`, googleTokenHash(ticket), identity.Subject, identity.Email, identity.First, identity.Last, time.Now().Add(10*time.Minute).Unix())
	if err != nil {
		googleRedirectError(w, r, c, "failed", false)
		return
	}
	googleCookie(w, c, ticket, 600)
	redirectGooglePage(w, r, c.frontend+"/auth/google/complete")
}

func redirectGooglePage(w http.ResponseWriter, r *http.Request, target string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func googleLogin(w http.ResponseWriter, r *http.Request, c googleConfig, userID int, target string) {
	session, err := CreateSession(userID)
	if err != nil {
		googleRedirectError(w, r, c, "failed", false)
		return
	}
	if _, err = db.DBInstance.DB.Exec(`UPDATE sessions SET auth_provider='google' WHERE token=?`, session.ID); err != nil {
		InvalidateSession(session.ID)
		googleRedirectError(w, r, c, "failed", false)
		return
	}
	SetSessionCookie(w, session)
	redirectGooglePage(w, r, c.frontend+target)
}

// Passwordless account changes need a recent, successful Google authentication.
func RecentGoogleAuthentication(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}
	var recent bool
	err = db.DBInstance.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions WHERE token=? AND is_active=1 AND auth_provider='google' AND julianday(created_at)>julianday(?))`, cookie.Value, time.Now().Add(-10*time.Minute)).Scan(&recent)
	return err == nil && recent
}
