package services

import (
	"encoding/json"
	"log"
	"net/http"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db/query"
)

// GetAllGroups handles the request to get all available groups
func GetAllGroups(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
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

	// Get all groups
	groups, err := query.GetAllGroups(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get all groups: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get groups"})
		return
	}

	// Return the groups
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}
