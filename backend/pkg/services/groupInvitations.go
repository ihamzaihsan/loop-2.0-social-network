package services

import (
	"log"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

// GetGroupInvitationsService returns all pending group invitations for a user
func GetGroupInvitationsService(userID int) ([]models.GroupInvitation, error) {
	invitations, err := query.GetGroupInvitations(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group invitations: %v", err)
		return nil, err
	}

	return invitations, nil
}
