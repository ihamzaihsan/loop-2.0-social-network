package query

import "socialNetwork/pkg/db"

// IsGroupCreator checks if a user is the creator of a group
func IsGroupCreator(groupID, userID int) (bool, error) {
	var isCreator bool
	err := db.DBInstance.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM groups 
			WHERE id = ? AND creator_id = ?
		)
	`, groupID, userID).Scan(&isCreator)

	return isCreator, err
}
