package query

import (
	"socialNetwork/pkg/db"
	"time"
)

// CreateGroup creates a new group and adds the creator as a member
func CreateGroup(title, description string, creatorID int) (int, error) {
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(
		"INSERT INTO groups (title, description, creator_id, created_at) VALUES (?, ?, ?, ?)",
		title, description, creatorID, time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = tx.Exec(
		"INSERT INTO group_members (group_id, user_id, role, status, created_at) VALUES (?, ?, ?, ?, ?)",
		groupID, creatorID, "creator", "active", time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(groupID), nil
}
