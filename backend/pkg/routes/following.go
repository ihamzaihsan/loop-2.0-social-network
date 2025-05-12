package routes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
)

// ServeFollowingUsers handles requests to get users that the current user is following
func ServeFollowingUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current user ID from the session
	userID, err := auth.GetUserID(r)
	if err != nil || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query the database directly to get following users with their details
	rows, err := db.DBInstance.DB.Query(`
		SELECT f.id, f.follower_id, f.following_id, f.status, 
			   u.first_name, u.last_name, u.nickname, u.avatar
		FROM followers f
		JOIN users u ON f.following_id = u.id
		WHERE f.follower_id = ? AND f.status = 'accept'
	`, userID)

	if err != nil {
		log.Printf("[ERROR] Failed to get following users: %v", err)
		http.Error(w, "Failed to get following users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var following []map[string]interface{}
	for rows.Next() {
		var id, followerID, followingID int
		var status, firstName, lastName string
		var nickname, avatar sql.NullString

		if err := rows.Scan(&id, &followerID, &followingID, &status, &firstName, &lastName, &nickname, &avatar); err != nil {
			log.Printf("[ERROR] Error scanning following row: %v", err)
			continue
		}

		// Create a user object with the retrieved data
		user := map[string]interface{}{
			"id":         id,
			"followerID": followerID,
			"followedID": followingID,
			"status":     status,
			"firstName":  firstName,
			"lastName":   lastName,
		}

		// Add nickname if available
		if nickname.Valid {
			user["nickname"] = nickname.String
		}

		// Add avatar if available
		if avatar.Valid {
			user["avatar"] = avatar.String
		}

		// Set display name (nickname or first+last name)
		if nickname.Valid {
			user["username"] = nickname.String
		} else {
			user["username"] = firstName + " " + lastName
		}

		following = append(following, user)
	}

	// Log the results for debugging
	log.Printf("[INFO] Found %d following users for user ID %d", len(following), userID)

	// Return the following users as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"following": following,
		"count":     len(following),
	})
}
