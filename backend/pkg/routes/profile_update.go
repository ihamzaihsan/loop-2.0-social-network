package routes

import (
	"encoding/json"
	"log"
	"net/http"
	auth "socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
)

// UpdatePrivacy handles requests to update a user's privacy setting
func UpdatePrivacy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var requestData struct {
		IsPrivate bool `json:"isPrivate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the user's privacy setting in the database
	_, err = db.DBInstance.DB.Exec(
		"UPDATE users SET isprivate = ? WHERE id = ?",
		requestData.IsPrivate, userID,
	)

	if err != nil {
		log.Printf("Error updating privacy setting: %v", err)
		http.Error(w, "Failed to update privacy setting", http.StatusInternalServerError)
		return
	}

	// Return success response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Privacy setting updated successfully",
		"isPrivate": requestData.IsPrivate,
	})
}
