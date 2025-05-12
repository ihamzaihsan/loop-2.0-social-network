package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	
	auth "socialNetwork/pkg/auth"
	query "socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

// GetUserProfile handles requests to view a specific user's profile
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current user ID from session for permission checking
	currentUserID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract user ID from URL path
	// URL format: /profile/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	
	userIDStr := pathParts[len(pathParts)-1]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get user info
	user, err := query.GetUserInfo(userID)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user is private and not the current user
	if user.IsPrivate && userID != currentUserID {
		// Check if the current user follows this user
		isFollowing, err := query.CheckIfFollowing(uint(currentUserID), uint(userID))
		if err != nil || !isFollowing {
			// Return limited profile data for private accounts
			limitedProfile := map[string]interface{}{
				"user": map[string]interface{}{
					"id":        user.ID,
					"firstName": user.FirstName,
					"lastName":  user.LastName,
					"nickname":  user.Nickname,
					"avatar":    user.Avatar,
					"isPrivate": user.IsPrivate,
				},
				"isPrivate": true,
			}
			json.NewEncoder(w).Encode(limitedProfile)
			return
		}
	}

	// Get posts, followers, following counts
	posts, err := query.GetPostsByUserID(userID)
	if err != nil {
		log.Printf("Error getting user posts: %v", err)
		posts = []models.PostResponse{} // Use the correct type
	}

	postsCount, _ := query.GetUserPostsCount(userID)
	followers, followersCount, _ := query.GetFollowers(uint(userID))
	following, followingCount, _ := query.GetFollowing(uint(userID))

	// Build profile response
	profile := map[string]interface{}{
		"user":           user,
		"posts":          posts,
		"postsCount":     postsCount,
		"followersCount": followersCount,
		"followingCount": followingCount,
		"followers":      followers,
		"following":      following,
		"isCurrentUser":  userID == currentUserID,
	}

	json.NewEncoder(w).Encode(profile)
}