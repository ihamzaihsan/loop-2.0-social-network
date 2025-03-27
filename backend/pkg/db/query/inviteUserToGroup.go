package query

import (
	"socialNetwork/pkg/db"
	"time"
)

// InviteUserToGroup creates an invitation for a user to join a group
func InviteUserToGroup(groupID, inviterID, inviteeID int) error {
	var exists bool
	err := db.DBInstance.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)
	`, groupID, inviteeID).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// Add the invitation
	_, err = db.DBInstance.DB.Exec(`
		INSERT INTO group_members (group_id, user_id, role, status, created_at) 
		VALUES (?, ?, ?, ?, ?)
	`, groupID, inviteeID, "member", "invited", time.Now())

	return err
}
