package query

import "socialNetwork/pkg/db"

// GetUserNames retrieves the first name and last name of a user by ID
func GetUserNames(userID int, firstName, lastName *string) error {
	return db.DBInstance.DB.QueryRow(`
		SELECT first_name, last_name FROM users WHERE id = ?
	`, userID).Scan(firstName, lastName)
}
