package query

import (
	"database/sql"
	"socialNetwork/pkg/db"
	"time"
)

// RespondToEvent records a user's response to an event
func RespondToEvent(eventID, userID, optionID int) error {
	// Check if the user has already responded to this event
	var existingID int
	err := db.DBInstance.DB.QueryRow(`
		SELECT id FROM event_responses 
		WHERE event_id = ? AND user_id = ?
	`, eventID, userID).Scan(&existingID)

	if err == nil {
		// User has already responded, update the existing response
		_, err = db.DBInstance.DB.Exec(`
			UPDATE event_responses 
			SET response_option_id = ?, created_at = ?
			WHERE id = ?
		`, optionID, time.Now(), existingID)
		return err
	} else if err != sql.ErrNoRows {
		// An error occurred other than "no rows found"
		return err
	}

	// User has not responded yet, create a new response
	_, err = db.DBInstance.DB.Exec(`
		INSERT INTO event_responses (event_id, user_id, response_option_id, created_at)
		VALUES (?, ?, ?, ?)
	`, eventID, userID, optionID, time.Now())

	return err
}
