package auth

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var validInputRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*(),.?" ' ':{}|<>/\s]+$`)

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

			// Save the avatar file
			uploadDir := "./uploads/avatars"
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
				return
			}

			filename := filepath.Join(uploadDir, header.Filename)
			out, err := os.Create(filename)
			if err != nil {
				http.Error(w, "Failed to create file", http.StatusInternalServerError)
				return
			}
			defer out.Close()

			_, err = io.Copy(out, file)
			if err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}

			// Store the web-accessible path
			avatarPath = "/uploads/avatars/" + header.Filename
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

	if email == "" || password == "" || firstName == "" || lastName == "" || dob == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
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

	result, err := db.DBInstance.DB.Exec(`
    INSERT INTO users (email, password, first_name, last_name, dob, avatar, nickname, about_me, isprivate)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Email, hashedPassword, user.FirstName, user.LastName, user.DOB, user.Avatar, user.Nickname, user.AboutMe, user.IsPrivate)
	if err != nil {
		log.Printf("Database error during user creation: %v", err)
		http.Error(w, "Error creating user: "+err.Error(), http.StatusInternalServerError)
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
		"token":   session.ID,
	})
}
