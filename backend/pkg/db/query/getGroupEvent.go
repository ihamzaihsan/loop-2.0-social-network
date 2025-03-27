package query

import (
	"database/sql"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetGroupEvent returns a specific event by ID
func GetGroupEvent(eventID, userID int) (*models.GroupEvent, error) {
	var event models.GroupEvent
	var userResponseID sql.NullInt64
	var userResponse sql.NullString

	err := db.DBInstance.DB.QueryRow(`
		SELECT ge.id, ge.group_id, ge.title, ge.description, ge.event_time, ge.created_at,
			   (SELECT er.response_option_id 
				FROM event_responses er 
				JOIN event_response_options ero ON er.response_option_id = ero.id
				WHERE er.event_id = ge.id AND er.user_id = ?) as user_response_id,
			   (SELECT ero.option_text 
				FROM event_responses er 
				JOIN event_response_options ero ON er.response_option_id = ero.id
				WHERE er.event_id = ge.id AND er.user_id = ?) as user_response
		FROM group_events ge
		WHERE ge.id = ?
	`, userID, userID, eventID).Scan(
		&event.ID,
		&event.GroupID,
		&event.Title,
		&event.Description,
		&event.EventTime,
		&event.CreatedAt,
		&userResponseID,
		&userResponse,
	)

	if err != nil {
		return nil, err
	}

	if userResponse.Valid {
		event.UserResponse = userResponse.String
	}

	optionRows, err := db.DBInstance.DB.Query(`
		SELECT ero.id, ero.event_id, ero.option_text,
			   (SELECT COUNT(*) FROM event_responses WHERE response_option_id = ero.id) as response_count
		FROM event_response_options ero
		WHERE ero.event_id = ?
	`, event.ID)

	if err == nil {
		defer optionRows.Close()

		for optionRows.Next() {
			var option models.EventResponseOption

			if err := optionRows.Scan(
				&option.ID,
				&option.EventID,
				&option.OptionText,
				&option.ResponseCount,
			); err == nil {
				event.ResponseOptions = append(event.ResponseOptions, option)
			}
		}
	}

	return &event, nil
}
