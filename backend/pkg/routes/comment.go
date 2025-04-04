package routes

import (
	"encoding/json"
	"log"
	"net/http"
	auth "socialNetwork/pkg/auth"
	query "socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
	"strings"

	"socialNetwork/pkg/utils"
	"strconv"
)

func HandleComments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetComments(w, r)
	case http.MethodPost:
		CreateComment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	// Add better logging to diagnose the issue
	log.Printf("Creating comment: Request method: %s, Content-Type: %s", r.Method, r.Header.Get("Content-Type"))

	userID, _ := auth.GetUserID(r)
	postID, err := strconv.Atoi(r.URL.Query().Get("postId"))
	if err != nil {
		log.Printf("Invalid post ID: %v", err)
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Check if post exists
	_, err = query.GetPostByIDQuery(postID)
	if err != nil {
		log.Printf("Post not found: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	var request models.CommentRequest

	// Check content type to determine how to parse the request
	contentType := r.Header.Get("Content-Type")
	log.Printf("Content-Type: %s", contentType)

	if strings.Contains(contentType, "application/json") {
		// Parse JSON request
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			http.Error(w, "Error parsing JSON request", http.StatusBadRequest)
			return
		}
	} else if strings.Contains(contentType, "multipart/form-data") {
		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Printf("Error parsing multipart form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		request.Content = r.FormValue("content")

		// Try to get image if present
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			imagePath, err := utils.HandleImageUpload(file, header)
			if err != nil {
				log.Printf("Error uploading image: %v", err)
				http.Error(w, "Failed to upload image", http.StatusInternalServerError)
				return
			}
			request.Image = imagePath
		} else if err != http.ErrMissingFile {
			// Only log if it's not just a missing file
			log.Printf("Error getting form file: %v", err)
		}
	} else {
		// Try to parse regular form
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		request.Content = r.FormValue("content")
	}

	log.Printf("Creating comment for post %d with content: %s", postID, request.Content)

	comment, err := query.CreateComment(userID, postID, request)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"comment": comment,
	})
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.URL.Query().Get("postId"))
	if err != nil {
		log.Printf("Invalid post ID: %v", err)
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Check if post exists
	_, err = query.GetPostByIDQuery(postID)
	if err != nil {
		log.Printf("Post not found: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	comments, err := query.GetCommentsByPostID(postID)
	if err != nil {
		log.Printf("Error fetching comments: %v", err)
		// Return empty comments instead of error
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"comments": []interface{}{},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"comments": comments,
	})
}
