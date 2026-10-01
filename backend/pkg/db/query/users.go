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
                      WHERE follower_id = ? AND following_id = u.id AND status = 'accept') as is_following
		FROM users u
		WHERE u.id != ?
	`, currentUserID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var firstName, lastName, email string
		var nicknameNull sql.NullString
		var isFollowing bool

		err := rows.Scan(&id, &firstName, &lastName, &email, &nicknameNull, &isFollowing)
		if err != nil {
			return nil, err
		}

		nickname := ""
		if nicknameNull.Valid {
			nickname = nicknameNull.String
		}

		user := map[string]interface{}{
			"id":        id,
			"firstName": firstName,
			"lastName":  lastName,
			"email":     email,
			"nickname":  nickname,
			"following": isFollowing,
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
