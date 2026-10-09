package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func GetUserInfo(userID int) (models.User, error) {
	var user models.User
	query := `
	SELECT id,email,first_name,last_name,dob,avatar,nickname,about_me,isprivate,created_at
	FROM users
	WHERE id = ? AND is_suspended=0
	`
	err := db.DBInstance.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.DOB,
		&user.Avatar,
		&user.Nickname,
		&user.AboutMe,
		&user.IsPrivate,
		&user.CreatedAt,
	)
	return user, err
}
