package query

import (
	"socialNetwork/pkg/db"
	"time"
)

// RequestToJoinGroup creates a request for a user to join a group
func RequestToJoinGroup(groupID, userID int) error {
	var exists bool
	err := db.DBInstance.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)
	`, groupID, userID).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		return nil 
	}

	_, err = db.DBInstance.DB.Exec(`
		INSERT INTO group_members (group_id, user_id, role, status, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, groupID, userID, "member", "requested", time.Now())

	return err
}
