package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetGroupJoinRequests returns all pending join requests for a group
func GetGroupJoinRequests(groupID int) ([]models.GroupJoinRequest, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT gm.id, gm.group_id, gm.user_id, gm.created_at,
			   u.first_name, u.last_name, u.avatar
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'requested'
		ORDER BY gm.created_at DESC
	`, groupID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.GroupJoinRequest
	for rows.Next() {
		var request models.GroupJoinRequest
		var avatar sql.NullString

		if err := rows.Scan(
			&request.ID,
			&request.GroupID,
			&request.UserID,
			&request.CreatedAt,
			&request.FirstName,
			&request.LastName,
			&avatar,
		); err != nil {
			log.Printf("Error scanning group join request row: %v", err)
			continue
		}

		request.Status = "pending"
		if avatar.Valid {
			request.Avatar = &avatar.String
		}

		requests = append(requests, request)
	}

	return requests, nil
}
