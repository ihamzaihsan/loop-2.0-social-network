package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
	"socialNetwork/pkg/utils"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var validInputRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*(),.?" ' ':{}|<>/\s]+$`)

func Register(w http.ResponseWriter, r *http.Request) {

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Check content type to determine how to parse the request
	contentType := r.Header.Get("Content-Type")

	var (
		email      string
		password   string
		firstName  string
		lastName   string
		dob        string
		nickname   string
		aboutMe    string
		isPrivate  bool
		avatarPath string
	)

	if strings.Contains(contentType, "multipart/form-data") {
		// Parse multipart form for file uploads
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Get form values
		email = r.FormValue("email")
		password = r.FormValue("password")
		firstName = r.FormValue("firstName")
		lastName = r.FormValue("lastName")
		dob = r.FormValue("dob")
		nickname = r.FormValue("nickname")
		aboutMe = r.FormValue("aboutMe")
		isPrivate = r.FormValue("isPrivate") == "true"

		// Handle avatar file upload
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()

			// Use the centralized image upload handler for consistency
			imagePath, err := utils.HandleImageUpload(file, header)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to upload avatar: %v", err), http.StatusInternalServerError)
				return
			}

			// Store the web-accessible path
			avatarPath = imagePath
		}
	} else {
		// Parse JSON for non-file requests
		var requestData struct {
			Email     string  `json:"email"`
			Password  string  `json:"password"`
			FirstName string  `json:"firstName"`
			LastName  string  `json:"lastName"`
			DOB       string  `json:"dob"`
			Nickname  *string `json:"nickname"`
			AboutMe   *string `json:"aboutMe"`
			IsPrivate bool    `json:"isPrivate"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		email = requestData.Email
		password = requestData.Password
		firstName = requestData.FirstName
		lastName = requestData.LastName
		dob = requestData.DOB
		if requestData.Nickname != nil {
			nickname = *requestData.Nickname
		}
		if requestData.AboutMe != nil {
			aboutMe = *requestData.AboutMe
		}
		isPrivate = requestData.IsPrivate
	}

	if db.DBInstance.DB == nil {
		log.Println("Database connection is nil")
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	email = strings.TrimSpace(email)
	if !ValidEmail(email) || !ValidPassword(password) {
		http.Error(w, "Use a valid email and a password of 8?72 bytes", 400)
		return
	}
	if email == "" || password == "" || firstName == "" || lastName == "" || dob == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Validate field lengths
	if len(firstName) > 100 {
		http.Error(w, "First name exceeds maximum length of 100 characters", http.StatusBadRequest)
		return
	}
	if len(lastName) > 100 {
		http.Error(w, "Last name exceeds maximum length of 100 characters", http.StatusBadRequest)
		return
	}
	if len(nickname) > 100 {
		http.Error(w, "Nickname exceeds maximum length of 100 characters", http.StatusBadRequest)
		return
	}
	if len(password) > 100 {
		http.Error(w, "Password exceeds maximum length of 100 characters", http.StatusBadRequest)
		return
	}

	if !validInputRegex.MatchString(firstName) ||
		!validInputRegex.MatchString(lastName) ||
		(nickname != "" && !validInputRegex.MatchString(nickname)) ||
		!validInputRegex.MatchString(password) {
		http.Error(w, "Only letters, numbers, spaces, and these symbols are allowed: !@#$%^&*(),.?\" ':{}|<>/\\", http.StatusBadRequest)
		return
	}

	dobTime, err := time.Parse("2006-01-02", dob)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Set up avatar pointer for database
	var avatarPtr *string
	if avatarPath != "" {
		avatarPtr = &avatarPath
	}

	// Set up other optional field pointers
	var nicknamePtr, aboutMePtr *string
	if nickname != "" {
		nicknamePtr = &nickname
	}
	if aboutMe != "" {
		aboutMePtr = &aboutMe
	}

	user := models.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		DOB:       dobTime,
		Avatar:    avatarPtr,
		Nickname:  nicknamePtr,
		AboutMe:   aboutMePtr,
		IsPrivate: isPrivate,
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		http.Error(w, "Error creating user", 500)
		return
	}
	defer tx.Rollback()
	allocation, err := tx.Exec(`INSERT INTO user_ids DEFAULT VALUES`)
	if err != nil {
		http.Error(w, "Error creating user", 500)
		return
	}
	userID, err := allocation.LastInsertId()
	if err != nil {
		http.Error(w, "Error creating user", 500)
		return
	}
	_, err = tx.Exec(`INSERT INTO users(id,email,password,first_name,last_name,dob,avatar,nickname,about_me,isprivate) VALUES (?,?,?,?,?,?,?,?,?,?)`, userID, user.Email, hashedPassword, user.FirstName, user.LastName, user.DOB, user.Avatar, user.Nickname, user.AboutMe, user.IsPrivate)
	if err != nil || tx.Commit() != nil {
		http.Error(w, "Unable to create account; email may be unavailable", 409)
		return
	}
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
