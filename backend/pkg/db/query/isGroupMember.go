package query

import "socialNetwork/pkg/db"

// IsGroupMember checks if a user is an active member of a group
func IsGroupMember(groupID, userID int) (bool, error) {
	var isMember bool
	err := db.DBInstance.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ? AND status = 'active'
		)
	`, groupID, userID).Scan(&isMember)

	return isMember, err
}
