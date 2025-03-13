package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func GetFollowing(userID int) ([]models.User, int, error) {
	var following []models.User
	var count int

	countQuery := `SELECT COUNT(*) FROM followers WHERE follower_id = ?`
	err := db.DBInstance.DB.QueryRow(countQuery, userID).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT u.id, u.email, u.first_name, u.last_name, u.dob, u.avatar, u.nickname, u.about_me, u.is_public, u.created_at 
		FROM users u 
		INNER JOIN followers f ON u.id = f.following_id 
		WHERE f.follower_id = ?
	`

	rows, err := db.DBInstance.DB.Query(query, userID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.DOB,
			&user.Avatar,
			&user.Nickname,
			&user.AboutMe,
			&user.IsPublic,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		following = append(following, user)
	}

	return following, count, nil
}
