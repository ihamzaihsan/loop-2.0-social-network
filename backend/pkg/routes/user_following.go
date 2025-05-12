package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	
	auth "socialNetwork/pkg/auth"
	query "socialNetwork/pkg/db/query"
)

// GetUserFollowing handles requests to view a user's following list
func GetUserFollowing(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get target user ID from query parameter, default to current user
	targetUserIDStr := r.URL.Query().Get("userId")
	targetUserID := userID
	
	if targetUserIDStr != "" {
		targetUserID, err = strconv.Atoi(targetUserIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
	}

	// Get following list
	following, count, err := query.GetFollowing(uint(targetUserID))
	if err != nil {
		log.Printf("Error getting following: %v", err)
		http.Error(w, "Error retrieving following", http.StatusInternalServerError)
		return
	}

	// Log the following data for debugging
	log.Printf("Found %d following for user ID %d", count, targetUserID)
	for i, f := range following {
		log.Printf("Following %d: ID=%d, Username=%s, Avatar=%s", 
			i, f.FollowedID, f.Username, f.Avatar)
	}

	// Return the following list
	response := map[string]interface{}{
		"success": true,
		"count": count,
		"following": following,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}