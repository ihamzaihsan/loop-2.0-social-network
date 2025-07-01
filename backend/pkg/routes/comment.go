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

	// Parse multipart form for all requests that might contain files
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		log.Printf("Error parsing multipart form (might not be multipart): %v", err)
		// If it's not a multipart form, try to parse as JSON or regular form
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				log.Printf("Error parsing JSON: %v", err)
				http.Error(w, "Error parsing request", http.StatusBadRequest)
				return
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
	} else {
		// Successfully parsed multipart form
		request.Content = r.FormValue("content")
		log.Printf("Parsed content from form: %s", request.Content)

		// Try to get image file
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			log.Printf("Found image file in request: %s (%d bytes)",
				header.Filename, header.Size)

			// Use the utils function to handle the image upload
			imagePath, err := utils.HandleImageUpload(file, header)
			if err != nil {
				log.Printf("Error uploading image: %v", err)
				http.Error(w, "Failed to upload image", http.StatusInternalServerError)
				return
			}
			log.Printf("Image uploaded successfully to: %s", imagePath)
			request.Image = imagePath
		} else {
			if err != http.ErrMissingFile {
				log.Printf("Error getting form file: %v", err)
			} else {
				log.Printf("No image file found in request")
			}
		}
	}

	// Validate content length
	if len(request.Content) > 100 {
		http.Error(w, "Comment content exceeds maximum length of 100 characters", http.StatusBadRequest)
		return
	}

	// Validate that comment has content or image
	if strings.TrimSpace(request.Content) == "" && request.Image == "" {
		http.Error(w, "Comment must contain either text content or an image", http.StatusBadRequest)
		return
	}

	log.Printf("Creating comment for post %d with content: %s and image: %s",
		postID, request.Content, request.Image)

	comment, err := query.CreateComment(userID, postID, request)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	log.Printf("Comment created successfully with ID: %d", comment.ID)

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
