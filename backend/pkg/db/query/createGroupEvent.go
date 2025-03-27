package query

import (
	"socialNetwork/pkg/db"
	"time"
)

// CreateGroupEvent creates a new event in a group
func CreateGroupEvent(groupID int, title, description string, eventTime time.Time) (int, error) {
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(`
		INSERT INTO group_events (group_id, title, description, event_time, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, groupID, title, description, eventTime, time.Now())

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = tx.Exec(`
		INSERT INTO event_response_options (event_id, option_text)
		VALUES (?, ?), (?, ?)
	`, eventID, "Going", eventID, "Not Going")

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(eventID), nil
}
