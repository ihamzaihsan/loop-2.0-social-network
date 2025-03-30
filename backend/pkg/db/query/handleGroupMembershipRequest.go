package query

import (
	"database/sql"
	"log"
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

	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		log.Printf("[ERROR] Failed to begin transaction: %v", err)
		return err
	}

	if action == "accept" {
		// Update the status to active
		_, err := tx.Exec(`
			UPDATE group_members 
			SET status = 'active' 
			WHERE group_id = ? AND user_id = ? AND status = ?
		`, groupID, userID, status)
		
		if err != nil {
			tx.Rollback()
			log.Printf("[ERROR] Failed to update group member status: %v", err)
			return err
		}
		
		// Check if the update was successful (affected rows)
		var count int
		err = tx.QueryRow("SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ? AND status = 'active'", 
			groupID, userID).Scan(&count)
		
		if err != nil || count == 0 {
			tx.Rollback()
			log.Printf("[ERROR] Failed to verify group member update: %v", err)
			return err
		}
		
		return tx.Commit()
	} else {
		// For rejection, delete the record
		_, err := tx.Exec(`
			DELETE FROM group_members 
			WHERE group_id = ? AND user_id = ? AND status = ?
		`, groupID, userID, status)
		
		if err != nil {
			tx.Rollback()
			log.Printf("[ERROR] Failed to delete group member record: %v", err)
			return err
		}
		
		return tx.Commit()
	}
}
