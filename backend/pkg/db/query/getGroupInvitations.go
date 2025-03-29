package query

import (
	"database/sql"
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

// GetJoinRequestsForCreator returns all join requests for groups where the user is the creator
func GetJoinRequestsForCreator(creatorID int) ([]models.GroupJoinRequest, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT gm.id, gm.group_id, g.title as group_title, gm.user_id, u.first_name, u.last_name, u.avatar, gm.created_at
		FROM group_members gm
		JOIN groups g ON gm.group_id = g.id
		JOIN users u ON gm.user_id = u.id
		WHERE g.creator_id = ? AND gm.status = 'requested'
		ORDER BY gm.created_at DESC
	`, creatorID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.GroupJoinRequest
	for rows.Next() {
		var request models.GroupJoinRequest
		var avatar sql.NullString
		var groupTitle string

		if err := rows.Scan(
			&request.ID,
			&request.GroupID,
			&groupTitle,
			&request.UserID,
			&request.FirstName,
			&request.LastName,
			&avatar,
			&request.CreatedAt,
		); err != nil {
			log.Printf("Error scanning join request row: %v", err)
			continue
		}

		request.Status = "pending"
		request.Title = groupTitle
		if avatar.Valid {
			request.Avatar = &avatar.String
		}

		requests = append(requests, request)
	}

	return requests, nil
}
