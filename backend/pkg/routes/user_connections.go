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

// GetUserConnections handles requests to view a user's followers or following
func GetUserConnections(w http.ResponseWriter, r *http.Request) {
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

	// Extract user ID and connection type from URL path
	// URL format: /user/{id}/connections/{type}
	// where type is either "followers" or "following"
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	userIDStr := pathParts[2]
	connectionType := pathParts[4]

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if connection type is valid
	if connectionType != "followers" && connectionType != "following" {
		http.Error(w, "Invalid connection type", http.StatusBadRequest)
		return
	}

	// Check if the user exists and get privacy status
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
			// Return empty list for private accounts that the user doesn't follow
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"message":     "This account is private",
				"connections": []models.Follow{},
			})
			return
		}
	}

	var connections []models.Follow
	var count int
	var fetchErr error

	// Fetch the appropriate connection list
	if connectionType == "followers" {
		connections, count, fetchErr = query.GetFollowers(uint(userID))
	} else {
		connections, count, fetchErr = query.GetFollowing(uint(userID))
	}

	if fetchErr != nil {
		log.Printf("Error fetching %s: %v", connectionType, fetchErr)
		http.Error(w, "Error retrieving connections", http.StatusInternalServerError)
		return
	}

	// Return the connections
	response := map[string]interface{}{
		"success":        true,
		"count":          count,
		"connections":    connections,
		"connectionType": connectionType,
		"userId":         userID,
		"isCurrentUser":  userID == currentUserID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
