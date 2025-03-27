package services

import (
	"log"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

// GetGroupEventsService returns all events in a group
func GetGroupEventsService(groupID, userID int) ([]models.GroupEvent, error) {
	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(groupID, userID)
	if err != nil || !isMember {
		return nil, err
	}

	// Get events
	events, err := query.GetGroupEvents(groupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group events: %v", err)
		return nil, err
	}

	return events, nil
}
