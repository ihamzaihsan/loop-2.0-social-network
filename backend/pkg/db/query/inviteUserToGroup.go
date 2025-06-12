package query

import (
	"database/sql"
	"errors"
	"socialNetwork/pkg/db"
	"time"
)

var (
	ErrUserAlreadyInvited = errors.New("user already has a pending invitation")
	ErrUserAlreadyMember  = errors.New("user is already a member of the group")
)

// InviteUserToGroup creates an invitation for a user to join a group
func InviteUserToGroup(groupID, inviterID, inviteeID int) error {
	// First check if user is already a member
	var memberStatus string
	err := db.DBInstance.DB.QueryRow(`
		SELECT status FROM group_members 
		WHERE group_id = ? AND user_id = ?
	`, groupID, inviteeID).Scan(&memberStatus)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// If user exists in group_members table
	if err == nil {
		switch memberStatus {
		case "active":
			return ErrUserAlreadyMember
		case "invited":
			return ErrUserAlreadyInvited
		case "rejected":
			// If previously rejected, update the invitation
			_, err = db.DBInstance.DB.Exec(`
				UPDATE group_members 
				SET status = 'invited', created_at = ?
				WHERE group_id = ? AND user_id = ?
			`, time.Now(), groupID, inviteeID)
			return err
		}
	}

	// If user doesn't exist in group_members table, create new invitation
	_, err = db.DBInstance.DB.Exec(`
		INSERT INTO group_members (group_id, user_id, role, status, created_at) 
		VALUES (?, ?, ?, ?, ?)
	`, groupID, inviteeID, "member", "invited", time.Now())

	return err
}
