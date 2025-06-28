package routes

import (
	"encoding/json"
	"net/http"
	auth "socialNetwork/pkg/auth"
	query "socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
	"socialNetwork/pkg/utils"
	"strconv"
	"strings"
)

func HandlePosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetPosts(w, r)
	case http.MethodPost:
		CreatePost(w, r)
	case http.MethodPut:
		UpdatePost(w, r)
	case http.MethodDelete:
		DeletePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// create the posts
func CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// If parsing multipart form fails, try to parse as JSON
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			var jsonRequest models.PostRequest
			if err := json.NewDecoder(r.Body).Decode(&jsonRequest); err != nil {
				http.Error(w, "Invalid request format", http.StatusBadRequest)
				return
			}

			// Validate content length
			if len(jsonRequest.Content) > 1000 {
				http.Error(w, "Post content exceeds maximum length of 1000 characters", http.StatusBadRequest)
				return
			}

			// Validate that post has content
			if strings.TrimSpace(jsonRequest.Content) == "" {
				http.Error(w, "Post must contain text content", http.StatusBadRequest)
				return
			}

			// Create post using the JSON data
			post, err := query.CreatePostQuery(userID, jsonRequest)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"post":    post,
			})
			return
		}

		http.Error(w, "Invalid request format or file too large", http.StatusBadRequest)
		return
	}

	var request models.PostRequest
	request.Content = r.FormValue("content")
	request.Privacy = r.FormValue("privacy")

	// Validate content length
	if len(request.Content) > 1000 {
		http.Error(w, "Post content exceeds maximum length of 1000 characters", http.StatusBadRequest)
		return
	}

	// Check if image will be provided
	_, _, imageErr := r.FormFile("image")
	hasImage := imageErr == nil

	// Validate that post has content or image
	if strings.TrimSpace(request.Content) == "" && !hasImage {
		http.Error(w, "Post must contain either text content or an image", http.StatusBadRequest)
		return
	}

	// Parse viewer IDs for private posts
	if request.Privacy == "private" {
		viewerIdsStr := r.FormValue("viewerIds")
		if viewerIdsStr != "" {
			var viewerIds []int
			if err := json.Unmarshal([]byte(viewerIdsStr), &viewerIds); err == nil {
				request.ViewerIDs = viewerIds
			}
		}
	}

	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imagePath, err := utils.HandleImageUpload(file, header)
		if err != nil {
			http.Error(w, "Failed to upload image", http.StatusInternalServerError)
			return
		}
		request.Image = imagePath
	}

	post, err := query.CreatePostQuery(userID, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"post":    post,
	})
}

// update the post
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// First, check if the post exists and belongs to the user
	post, err := query.GetPostByIDQuery(postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if post.UserID != userID {
		http.Error(w, "You don't have permission to edit this post", http.StatusForbidden)
		return
	}

	// Parse request body
	var request models.PostRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Update only the fields that are allowed to be updated
	// In this case, content and privacy
	updateData := models.PostRequest{
		Content: request.Content,
		Privacy: request.Privacy,
		// Don't allow image to be updated
		Image: post.Image,
	}

	if err := query.UpdatePostQuery(postID, userID, updateData); err != nil {
		http.Error(w, "Failed to update post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Post updated successfully",
	})
}

// delete the posts
func DeletePost(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	if err := query.DeletePostQuery(postID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// Add this helper function at the top of the file (outside any other function)
func fixImagePath(image string) string {
	if image != "" {
		// Convert backslashes to forward slashes
		image = strings.ReplaceAll(image, "\\", "/")

		// If the image path starts with "./uploads", convert it to "/uploads"
		if strings.HasPrefix(image, "./uploads") {
			image = strings.Replace(image, "./uploads", "/uploads", 1)
		}

		// Ensure the path starts with /uploads if it doesn't already
		if !strings.HasPrefix(image, "/uploads") && !strings.HasPrefix(image, "http") {
			if strings.HasPrefix(image, "uploads") {
				image = "/" + image
			}
		}
	}
	return image
}
func GetPosts(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserID(r)

	postIDStr := r.URL.Query().Get("id")
	if postIDStr != "" {
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		post, err := query.GetPostByIDQuery(postID)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		if post.UserID != userID && post.Privacy != "public" {
			http.Error(w, "Unauthorized to view this post", http.StatusForbidden)
			return
		}

		post.Image = fixImagePath(post.Image)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"post":    post,
		})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	posts, total, err := query.GetVisiblePosts(userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range posts {
		posts[i].Image = fixImagePath(posts[i].Image)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"posts":   posts,
		"total":   total,
		"page":    page,
	})
}
