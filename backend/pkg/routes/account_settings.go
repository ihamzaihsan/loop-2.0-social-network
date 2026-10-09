package routes

import (
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/utils"
	"strings"
)

func AccountSettings(w http.ResponseWriter, r *http.Request) {
	id := viewer(r)
	if r.Method == http.MethodGet {
		user, err := query.GetUserInfo(id)
		if err != nil {
			http.Error(w, "Account unavailable", 500)
			return
		}
		var hasPassword, googleConnected bool
		if db.DBInstance.DB.QueryRow(`SELECT password!='',EXISTS(SELECT 1 FROM oauth_identities WHERE provider='google' AND user_id=users.id) FROM users WHERE id=?`, id).Scan(&hasPassword, &googleConnected) != nil {
			http.Error(w, "Account unavailable", 500)
			return
		}
		respond(w, map[string]interface{}{"user": user, "isModerator": auth.IsModerator(id), "hasPassword": hasPassword, "googleConnected": googleConnected})
		return
	}
	if r.Method == http.MethodPut {
		r.Body = http.MaxBytesReader(w, r.Body, 6<<20)
		if err := r.ParseMultipartForm(6 << 20); err != nil {
			http.Error(w, "Invalid form", 400)
			return
		}
		defer r.MultipartForm.RemoveAll()
		first, last, nickname, about := strings.TrimSpace(r.FormValue("firstName")), strings.TrimSpace(r.FormValue("lastName")), strings.TrimSpace(r.FormValue("nickname")), strings.TrimSpace(r.FormValue("aboutMe"))
		if !validText(first, 100, true) || !validText(last, 100, true) || !validText(nickname, 100, false) || !validText(about, 500, false) {
			http.Error(w, "Check name and biography lengths", 400)
			return
		}
		var avatar interface{}
		if r.FormValue("removeAvatar") == "true" {
			avatar = ""
		} else if file, header, err := r.FormFile("avatar"); err == nil {
			defer file.Close()
			path, err := utils.HandleImageUpload(file, header, viewer(r))
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			avatar = path
		}
		_, err := db.DBInstance.DB.Exec(`UPDATE users SET first_name=?,last_name=?,nickname=NULLIF(?,''),about_me=NULLIF(?,''),avatar=CASE WHEN ? IS NULL THEN avatar ELSE NULLIF(?,'') END WHERE id=?`, first, last, nickname, about, avatar, avatar, id)
		if err != nil {
			http.Error(w, "Unable to save profile", 500)
			return
		}
		respond(w, map[string]bool{"success": true})
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var input struct {
		CurrentPassword string `json:"currentPassword"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		Confirmation    string `json:"confirmation"`
	}
	if !decode(w, r, &input) {
		return
	}
	var saved string
	if db.DBInstance.DB.QueryRow(`SELECT password FROM users WHERE id=?`, id).Scan(&saved) != nil {
		http.Error(w, "Account unavailable", 500)
		return
	}
	if saved == "" && !auth.RecentGoogleAuthentication(r) {
		http.Error(w, "Sign in with Google again to confirm this account change", 403)
		return
	}
	if saved != "" && bcrypt.CompareHashAndPassword([]byte(saved), []byte(input.CurrentPassword)) != nil {
		http.Error(w, "Current password is incorrect", 403)
		return
	}
	if r.Method == http.MethodDelete {
		if input.Confirmation != "DELETE" {
			http.Error(w, "Type DELETE to confirm", 400)
			return
		}
		tx, err := db.DBInstance.DB.Begin()
		if err != nil {
			http.Error(w, "Unable to delete account", 500)
			return
		}
		defer tx.Rollback()
		if err = query.DeleteAccountTx(tx, id); err != nil {
			http.Error(w, "Unable to delete account", 500)
			return
		}
		if tx.Commit() != nil {
			http.Error(w, "Unable to delete account", 500)
			return
		}
		auth.RevokeUserSessions(id)
		http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
		respond(w, map[string]bool{"success": true})
		return
	}
	email := strings.TrimSpace(input.Email)
	if !auth.ValidEmail(email) || (input.Password != "" && !auth.ValidPassword(input.Password)) {
		http.Error(w, "Use a valid email and a password of 8–72 bytes", 400)
		return
	}
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Unable to change password", 500)
			return
		}
		saved = string(hash)
	}
	if _, err := db.DBInstance.DB.Exec(`UPDATE users SET email=?,password=? WHERE id=?`, email, saved, id); err != nil {
		http.Error(w, "Email is unavailable", 409)
		return
	}
	db.DBInstance.DB.Exec(`DELETE FROM password_resets WHERE user_id=?`, id)
	auth.RevokeUserSessions(id)
	respond(w, map[string]bool{"success": true, "loginRequired": true})
}
