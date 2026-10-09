package query

import (
	"database/sql"
	"socialNetwork/pkg/db"
)

// GetAllUsers retrieves all users from the database with follow status
func GetAllUsers(currentUserID int) ([]map[string]interface{}, error) {
	rows, err := db.DBInstance.DB.Query(`
		SELECT u.id, u.first_name, u.last_name, u.email, u.nickname,
               EXISTS(SELECT 1 FROM followers 
                      WHERE follower_id = ? AND following_id = u.id AND status = 'accept') as is_following,
 EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = u.id AND status = 'pending'), u.isprivate
		FROM users u
		WHERE u.id != ? AND u.is_suspended=0 AND NOT EXISTS(SELECT 1 FROM user_blocks WHERE (blocker_id=? AND blocked_id=u.id) OR (blocker_id=u.id AND blocked_id=?))
	`, currentUserID, currentUserID, currentUserID, currentUserID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var firstName, lastName, email string
		var nicknameNull sql.NullString
		var isFollowing, isPending, isPrivate bool

		err := rows.Scan(&id, &firstName, &lastName, &email, &nicknameNull, &isFollowing, &isPending, &isPrivate)
		if err != nil {
			return nil, err
		}

		nickname := ""
		if nicknameNull.Valid {
			nickname = nicknameNull.String
		}

		user := map[string]interface{}{
			"id":            id,
			"firstName":     firstName,
			"lastName":      lastName,
			"email":         email,
			"nickname":      nickname,
			"following":     isFollowing,
			"pendingFollow": isPending,
			"isPrivate":     isPrivate,
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
