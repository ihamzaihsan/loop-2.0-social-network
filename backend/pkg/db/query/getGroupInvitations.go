package query

import (
	"log"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

// GetGroupInvitations returns all pending invitations for a user
func GetGroupInvitations(userID int) ([]models.GroupInvitation, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT gm.id, gm.group_id, g.title, gm.created_at,
			   u.id as inviter_id, u.first_name, u.last_name
		FROM group_members gm
		JOIN groups g ON gm.group_id = g.id
		JOIN users u ON g.creator_id = u.id
		WHERE gm.user_id = ? AND gm.status = 'invited'
		ORDER BY gm.created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []models.GroupInvitation
	for rows.Next() {
		var invitation models.GroupInvitation
		var firstName, lastName string

		if err := rows.Scan(
			&invitation.ID,
			&invitation.GroupID,
			&invitation.GroupTitle,
			&invitation.CreatedAt,
			&invitation.InviterID,
			&firstName,
			&lastName,
		); err != nil {
			log.Printf("Error scanning group invitation row: %v", err)
			continue
		}

		invitation.InviteeID = userID
		invitation.Status = "pending"
		invitation.InviterName = firstName + " " + lastName

		invitations = append(invitations, invitation)
	}

	return invitations, nil
}
