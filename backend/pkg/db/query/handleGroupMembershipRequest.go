package query

import (
	"database/sql"
	"socialNetwork/pkg/db"
)

// HandleGroupMembershipRequest processes a group invitation or join request
func HandleGroupMembershipRequest(groupID, userID int, action, requestType string) error {
	if action != "accept" && action != "reject" {
		return sql.ErrNoRows
	}

	var status string
	if requestType == "invitation" {
		status = "invited"
	} else if requestType == "request" {
		status = "requested"
	} else {
		return sql.ErrNoRows
	}

	if action == "accept" {
		_, err := db.DBInstance.DB.Exec(`
			UPDATE group_members 
			SET status = 'active' 
			WHERE group_id = ? AND user_id = ? AND status = ?
		`, groupID, userID, status)
		return err
	} else {
		_, err := db.DBInstance.DB.Exec(`
			DELETE FROM group_members 
			WHERE group_id = ? AND user_id = ? AND status = ?
		`, groupID, userID, status)
		return err
	}
}
