package services

import (
	"log"
	"socialNetwork/pkg/db/query"
)

// RespondToEventService records a user's response to an event
func RespondToEventService(eventID, userID, optionID int) error {
	// Get event to check group membership
	event, err := query.GetGroupEvent(eventID, userID)
	if err != nil {
		return err
	}

	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(event.GroupID, userID)
	if err != nil || !isMember {
		return err
	}

	// Record response
	err = query.RespondToEvent(eventID, userID, optionID)
	if err != nil {
		log.Printf("[ERROR] Failed to record event response: %v", err)
		return err
	}

	return nil
}
