package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

// FollowUser handles the HTTP request to follow another user
func FollowUser(w http.ResponseWriter, r *http.Request) {
	// Set headers for CORS
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the request body
	var req struct {
		FollowedID uint `json:"followed_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate that user isn't trying to follow themselves
	if uint(userID) == req.FollowedID {
		http.Error(w, "You cannot follow yourself", http.StatusBadRequest)
		return
	}

	// Use repository to handle follow logic
	status, err := query.RequestFollow(uint(userID), req.FollowedID)
	if err != nil {
		log.Printf("Error following user: %v", err)
		http.Error(w, "Error processing follow request", http.StatusInternalServerError)
		return
	}

	// Return response
	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  status,
		Message: fmt.Sprintf("Follow request %s", getStatusMessage(status)),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// UnfollowUser handles the HTTP request to unfollow a user
func UnfollowUser(w http.ResponseWriter, r *http.Request) {
	// Set headers for CORS
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the request body
	var req struct {
		FollowedID uint `json:"followed_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Use repository to unfollow
	err = query.UnfollowUser(uint(userID), req.FollowedID)
	if err != nil {
		log.Printf("Error unfollowing user: %v", err)
		http.Error(w, "Error processing unfollow request", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "Successfully unfollowed user",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// HandleFollowRequest processes accepting or rejecting a follow request
func HandleFollowRequest(w http.ResponseWriter, r *http.Request) {
	// Set headers for CORS
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// If it's a POST request with _method=PATCH, treat it as a PATCH
	if r.Method == http.MethodPost {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		var methodOverride struct {
			Method string `json:"_method"`
		}
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&methodOverride); err == nil && methodOverride.Method == "PATCH" {
			r.Method = http.MethodPatch
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Get the user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get request ID from URL params
	requestIDStr := r.URL.Query().Get("id")
	if requestIDStr == "" {
		http.Error(w, "Missing request ID", http.StatusBadRequest)
		return
	}

	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	// Parse request body to get status
	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate status
	if req.Status != "accept" && req.Status != "reject" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	var followerID uint
	if req.Status == "accept" {
		followerID, err = query.AcceptFollowRequest(requestID, uint(userID))
		if err != nil {
			log.Printf("Error accepting follow request: %v", err)
			http.Error(w, "Failed to accept follow request", http.StatusInternalServerError)
			return
		}
	} else {
		err = query.RejectFollowRequest(requestID, uint(userID))
		if err != nil {
			log.Printf("Error rejecting follow request: %v", err)
			http.Error(w, "Failed to reject follow request", http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	response := struct {
		Success    bool   `json:"success"`
		Message    string `json:"message"`
		FollowerID uint   `json:"follower_id,omitempty"`
	}{
		Success:    true,
		Message:    fmt.Sprintf("Follow request %s", req.Status+"ed"),
		FollowerID: followerID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// GetFollowers returns the authenticated user's followers
func GetFollowers(w http.ResponseWriter, r *http.Request) {
	// Set headers for CORS
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get followers from repository
	followers, _, err := query.GetFollowers(uint(userID))
	if err != nil {
		log.Printf("Error getting followers: %v", err)
		http.Error(w, "Error retrieving followers", http.StatusInternalServerError)
		return
	}

	// Return followers as JSON
	if err := json.NewEncoder(w).Encode(followers); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// GetFollowing returns the users the authenticated user is following
func GetFollowing(w http.ResponseWriter, r *http.Request) {
	// Set headers for CORS
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get following from repository
	following, _, err := query.GetFollowing(uint(userID))
	if err != nil {
		log.Printf("Error getting following: %v", err)
		http.Error(w, "Error retrieving following", http.StatusInternalServerError)
		return
	}

	// Return following as JSON
	if err := json.NewEncoder(w).Encode(following); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// GetFollowRequests returns pending follow requests for the authenticated user
func GetFollowRequests(w http.ResponseWriter, r *http.Request) {
	// Set headers for CORS
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("Fetching follow requests for user ID: %d", userID)

	// Get follow requests from repository
	requests, err := query.GetFollowRequests(uint(userID))
	if err != nil {
		log.Printf("Error getting follow requests: %v", err)
		http.Error(w, "Error retrieving follow requests", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d follow requests", len(requests))

	// Return requests as JSON with a success flag
	response := struct {
		Success  bool            `json:"success"`
		Requests []models.Follow `json:"requests"`
	}{
		Success:  true,
		Requests: requests,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// GetFollowStatuses returns all follow statuses for the authenticated user
func GetFollowStatuses(w http.ResponseWriter, r *http.Request) {
	// Set headers for CORS
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get follow statuses from repository
	statuses, err := query.GetFollowStatuses(uint(userID))
	if err != nil {
		log.Printf("Error getting follow statuses: %v", err)
		http.Error(w, "Error retrieving follow statuses", http.StatusInternalServerError)
		return
	}

	// Return statuses as JSON
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// Helper function to get a user-friendly message for a follow status
func getStatusMessage(status string) string {
	switch status {
	case "accept":
		return "accepted"
	case "pending":
		return "pending approval"
	case "reject":
		return "rejected"
	default:
		return status
	}
}
