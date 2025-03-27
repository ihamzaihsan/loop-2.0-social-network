package services

import (
	"log"
	"socialNetwork/pkg/db/query"
)

// CreateGroupService handles the logic for creating a new group
func CreateGroupService(title, description string, creatorID int) (int, error) {
	groupID, err := query.CreateGroup(title, description, creatorID)
	if err != nil {
		log.Printf("[ERROR] Failed to create group: %v", err)
		return 0, err
	}

	return groupID, nil
}
