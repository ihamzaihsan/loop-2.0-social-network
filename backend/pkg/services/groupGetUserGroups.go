package services

import (
	"log"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

// GetUserGroupsService returns all groups a user is a member of
func GetUserGroupsService(userID int) ([]models.Group, error) {
	groups, err := query.GetUserGroups(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user groups: %v", err)
		return nil, err
	}

	return groups, nil
}
