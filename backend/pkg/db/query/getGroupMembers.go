package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetGroupMembers returns all active members of a group
func GetGroupMembers(groupID int) ([]models.GroupMember, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.created_at,
			   u.first_name, u.last_name, u.avatar
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'active'
		ORDER BY 
			CASE 
				WHEN gm.role = 'creator' THEN 1
				WHEN gm.role = 'admin' THEN 2
				ELSE 3
			END,
			u.first_name, u.last_name
	`, groupID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.GroupMember
	for rows.Next() {
		var member models.GroupMember
		var avatar sql.NullString

		if err := rows.Scan(
			&member.ID,
			&member.GroupID,
			&member.UserID,
			&member.Role,
			&member.Status,
			&member.CreatedAt,
			&member.FirstName,
			&member.LastName,
			&avatar,
		); err != nil {
			log.Printf("Error scanning group member row: %v", err)
			continue
		}

		if avatar.Valid {
			member.Avatar = &avatar.String
		}

		members = append(members, member)
	}

	return members, nil
}
