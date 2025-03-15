package query

import (
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/models"
)

func GetUserInfo(userID int) (models.User, error) {
	var user models.User
	query := `
	SELECT id,email,first_name,last_name,dob,avatar,nickname,about_me,is_public,created_at
	FROM users
	WHERE id = ?
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
		&user.IsPublic,
		&user.CreatedAt,
	)
	return user, err
}
