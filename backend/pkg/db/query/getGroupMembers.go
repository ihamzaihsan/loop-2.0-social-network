package query

import (
	"database/sql"

	"socialNetwork/pkg/db"
)

type GroupMember struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Avatar    *string `json:"avatar,omitempty"`
	Role      string  `json:"role"`
	Status    string  `json:"status"`
}

// GetGroupMembers returns all members (including pending)
func GetGroupMembers(groupID int) ([]GroupMember, error) {
	rows, err := db.DBInstance.DB.Query(`


		SELECT gm.id, gm.user_id, u.first_name, u.last_name, u.avatar, gm.role, gm.status
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ?
		ORDER BY gm.created_at ASC
	`, groupID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []GroupMember
	for rows.Next() {

		var member GroupMember
		var avatar sql.NullString

		if err := rows.Scan(
			&member.ID,
			&member.UserID,
			&member.FirstName,
			&member.LastName,
			&avatar,
			&member.Role,
			&member.Status,
		); err != nil {
			continue
		}

		if avatar.Valid {

			avatarStr := avatar.String
			member.Avatar = &avatarStr
		}

		members = append(members, member)
	}

	return members, nil
}

// GetActiveGroupMembers returns only active members
func GetActiveGroupMembers(groupID int) ([]GroupMember, error) {
	rows, err := db.DBInstance.DB.Query(`


		SELECT gm.id, gm.user_id, u.first_name, u.last_name, u.avatar, gm.role, gm.status
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'active'
		ORDER BY gm.created_at ASC
	`, groupID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []GroupMember
	for rows.Next() {

		var member GroupMember
		var avatar sql.NullString

		if err := rows.Scan(
			&member.ID,
			&member.UserID,
			&member.FirstName,
			&member.LastName,
			&avatar,
			&member.Role,
			&member.Status,
		); err != nil {
			continue
		}

		if avatar.Valid {

			avatarStr := avatar.String
			member.Avatar = &avatarStr
		}

		members = append(members, member)
	}

	return members, nil
}

// GetActiveGroupMembersCount returns count of only active members
func GetActiveGroupMembersCount(groupID int) (int, error) {
	var count int
	err := db.DBInstance.DB.QueryRow(`
        SELECT COUNT(*) FROM group_members 
        WHERE group_id = ? AND status = 'active'
    `, groupID).Scan(&count)

	return count, err
}
