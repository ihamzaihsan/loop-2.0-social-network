package auth

import (
	"encoding/json"
	"net/http"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var requestData struct {
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		FirstName string  `json:"firstName"`
		LastName  string  `json:"lastName"`
		DOB       string  `json:"dob"`
		Avatar    *string `json:"avatar"`
		Nickname  *string `json:"nickname"`
		AboutMe   *string `json:"aboutMe"`
		IsPublic  bool    `json:"isPublic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if requestData.Email == "" || requestData.Password == "" || requestData.FirstName == "" || requestData.LastName == "" || requestData.DOB == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

    dob, err := time.Parse("2006-01-02", requestData.DOB)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	user := models.User{
		Email:     requestData.Email,
		FirstName: requestData.FirstName,
		LastName:  requestData.LastName,
		DOB:       dob,
		Avatar:    requestData.Avatar,
		Nickname:  requestData.Nickname,
		AboutMe:   requestData.AboutMe,
		Is_Public: requestData.IsPublic,
	}

	var exists bool
	err = db.DBInstance.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", user.Email).Scan(&exists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestData.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	result, err := db.DBInstance.DB.Exec(`
        INSERT INTO users (email, password, first_name, last_name, dob, avatar, nickname, about_me, is_public)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Email, hashedPassword, user.FirstName, user.LastName, user.DOB, user.Avatar, user.Nickname, user.AboutMe, user.Is_Public)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	userID, _ := result.LastInsertId()
	user.ID = int(userID)

	session, err := CreateSession(user.ID)
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}
	SetSessionCookie(w, session)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})
}
