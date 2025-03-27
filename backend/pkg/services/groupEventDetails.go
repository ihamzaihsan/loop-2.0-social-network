package services

import (
	"log"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

// GetGroupEventService returns a specific event by ID
func GetGroupEventService(eventID, userID int) (*models.GroupEvent, []models.EventResponse, error) {
	// Get event
	event, err := query.GetGroupEvent(eventID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get event: %v", err)
		return nil, nil, err
	}

	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(event.GroupID, userID)
	if err != nil || !isMember {
		return nil, nil, err
	}

	// Get responses
	responses, err := query.GetEventResponses(eventID)
	if err != nil {
		log.Printf("[ERROR] Failed to get event responses: %v", err)
		responses = []models.EventResponse{}
	}

	return event, responses, nil
}
