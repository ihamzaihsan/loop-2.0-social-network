package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	// CORS headers are now handled by CorsMiddleware, so we don't need to duplicate them here
	// But we'll keep the OPTIONS handling for safety
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.DBInstance.DB.QueryRow(
		"SELECT id, email, password, first_name, last_name, dob, avatar, nickname, about_me, isprivate, created_at FROM users WHERE email = ? AND is_suspended=0",
		loginReq.Email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.DOB, &user.Avatar, &user.Nickname, &user.AboutMe, &user.IsPrivate, &user.CreatedAt)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid email or password",
		})
		return
	}

	session, err := CreateSession(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create session",
		})
		return
	}

	SetSessionCookie(w, session)

	user.Password = ""

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
		"token":   session.ID,
	})
}
