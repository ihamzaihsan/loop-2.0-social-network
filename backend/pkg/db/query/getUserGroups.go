package query

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetUserGroups returns all groups a user is a member of
func GetUserGroups(userID int) ([]models.Group, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT g.id, g.title, g.description, g.creator_id, g.created_at,
			   (SELECT COUNT(*) FROM group_members WHERE group_id = g.id AND status = 'active') as member_count
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ? AND gm.status = 'active'
		ORDER BY g.created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		var memberCount int

		if err := rows.Scan(
			&group.ID,
			&group.Title,
			&group.Description,
			&group.CreatorID,
			&group.CreatedAt,
			&memberCount,
		); err != nil {
			log.Printf("Error scanning group row: %v", err)
			continue
		}

		group.MemberCount = memberCount
		group.IsCreator = group.CreatorID == userID

		groups = append(groups, group)
	}

	return groups, nil
}
