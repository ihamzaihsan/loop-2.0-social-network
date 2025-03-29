package query

import (
	"database/sql"
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetAllGroups returns all available groups with membership status for the given user
func GetAllGroups(userID int) ([]models.Group, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT g.id, g.title, g.description, g.creator_id, g.created_at,
			   (SELECT COUNT(*) FROM group_members WHERE group_id = g.id AND status = 'active') as member_count,
			   gm.role, gm.status
		FROM groups g
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.user_id = ?
		ORDER BY g.created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		var role, status sql.NullString

		if err := rows.Scan(
			&group.ID,
			&group.Title,
			&group.Description,
			&group.CreatorID,
			&group.CreatedAt,
			&group.MemberCount,
			&role,
			&status,
		); err != nil {
			log.Printf("Error scanning group row: %v", err)
			continue
		}

		if role.Valid {
			group.Role = role.String
		}
		if status.Valid {
			group.Status = status.String
		}

		groups = append(groups, group)
	}

	return groups, nil
}
