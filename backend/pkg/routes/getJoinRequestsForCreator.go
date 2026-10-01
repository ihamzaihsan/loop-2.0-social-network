package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db/query"
)

// GetJoinRequestsForCreator handles the request to get join requests for groups where the user is the creator
func GetJoinRequestsForCreator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the user ID from the session
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("[ERROR] Failed to get user ID: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get join requests for groups where the user is the creator
	requests, err := query.GetJoinRequestsForCreator(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get join requests: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get join requests"})
		return
	}

	// Return the requests
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"requests": requests,
	})
}
