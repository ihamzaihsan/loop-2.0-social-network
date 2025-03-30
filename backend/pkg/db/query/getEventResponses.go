package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetEventResponses returns all responses for an event
func GetEventResponses(eventID int) ([]models.EventResponse, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT er.id, er.event_id, er.user_id, er.response_option_id, er.created_at,
			   u.first_name, u.last_name, u.avatar,
			   ero.option_text
		FROM event_responses er
		JOIN users u ON er.user_id = u.id
		JOIN event_response_options ero ON er.response_option_id = ero.id
		WHERE er.event_id = ?
		ORDER BY er.created_at DESC
	`, eventID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []models.EventResponse
	for rows.Next() {
		var response models.EventResponse
		var avatar sql.NullString

		if err := rows.Scan(
			&response.ID,
			&response.EventID,
			&response.UserID,
			&response.ResponseOptionID,
			&response.CreatedAt,
			&response.FirstName,
			&response.LastName,
			&avatar,
			&response.OptionText,
		); err != nil {
			log.Printf("Error scanning event response row: %v", err)
			continue
		}

		if avatar.Valid {
			response.Avatar = &avatar.String
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetUserEventResponse returns a user's response to a specific event
func GetUserEventResponse(eventID, userID int) (int, error) {
    var responseOptionID int
    err := db.DBInstance.DB.QueryRow(`
        SELECT response_option_id
        FROM event_responses
        WHERE event_id = ? AND user_id = ?
    `, eventID, userID).Scan(&responseOptionID)

    if err == sql.ErrNoRows {
        return 0, nil // User hasn't responded yet
    }

    return responseOptionID, err
}




