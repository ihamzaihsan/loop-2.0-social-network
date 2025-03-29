package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetGroupByID returns a specific group by ID
func GetGroupByID(groupID, userID int) (*models.Group, error) {
	var group models.Group
	var memberCount int

	err := db.DBInstance.DB.QueryRow(`
		SELECT g.id, g.title, g.description, g.creator_id, g.created_at,
			   (SELECT COUNT(*) FROM group_members WHERE group_id = g.id AND status = 'active') as member_count
		FROM groups g
		WHERE g.id = ?
	`, groupID).Scan(
		&group.ID,
		&group.Title,
		&group.Description,
		&group.CreatorID,
		&group.CreatedAt,
		&memberCount,
	)

	if err != nil {
		return nil, err
	}

	group.MemberCount = memberCount
	group.IsCreator = group.CreatorID == userID

	return &group, nil
}