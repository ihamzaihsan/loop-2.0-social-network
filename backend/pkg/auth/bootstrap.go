package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"socialNetwork/pkg/db"
	"strings"
)

// BootstrapModerator provisions a trusted owner once, before accepting requests.
// Rebuilds never reset credentials or grant roles to existing email addresses.
func BootstrapModerator() error {
	email := strings.TrimSpace(os.Getenv("BOOTSTRAP_MODERATOR_EMAIL"))
	password := os.Getenv("BOOTSTRAP_MODERATOR_PASSWORD")
	if email == "" && password == "" {
		return nil
	}
	var count int
	if err := db.DBInstance.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if !ValidEmail(email) || !ValidPassword(password) {
		return fmt.Errorf("fresh database requires a valid BOOTSTRAP_MODERATOR_EMAIL and BOOTSTRAP_MODERATOR_PASSWORD of 8-72 bytes")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = tx.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	result, err := tx.Exec(`INSERT INTO user_ids DEFAULT VALUES`)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO users(id,email,password,first_name,last_name,dob,isprivate,is_moderator) VALUES (?,?,?,'Site','Moderator','1970-01-01',1,1)`, id, email, hashed)
	if err != nil {
		return err
	}
	return tx.Commit()
}
