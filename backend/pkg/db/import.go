package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// ImportSQLite copies durable data into an empty PostgreSQL application schema.
// Local sessions and one-time recovery/OAuth flows deliberately expire here.
// The source is opened read-only; a failed import rolls back the entire copy.
func ImportSQLite(source string) error {
	if !IsPostgres() {
		return fmt.Errorf("import requires DATABASE_URL")
	}
	absolute, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	path := filepath.ToSlash(absolute)
	if filepath.VolumeName(absolute) != "" {
		path = "/" + path
	}
	uri := url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}
	local, err := sql.Open("sqlite3", uri.String())
	if err != nil {
		return err
	}
	defer local.Close()
	snapshot, err := local.Begin()
	if err != nil {
		return err
	}
	defer snapshot.Rollback()
	tables := []string{"user_ids", "users", "followers", "posts", "comments", "likes", "groups", "group_members", "group_posts", "group_events", "event_response_options", "event_responses", "chats", "messages", "group_chat_messages", "notifications", "post_viewers", "group_invitations", "group_join_requests", "group_comments", "media_uploads", "user_blocks", "reports", "oauth_identities"}
	tx, err := DBInstance.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range tables {
		var count int
		if err = tx.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("target table %s is not empty; refusing to overwrite data", table)
		}
	}
	for _, table := range tables {
		rows, queryErr := snapshot.Query(`SELECT * FROM ` + table)
		if queryErr != nil {
			return fmt.Errorf("source schema must be migrated: %w", queryErr)
		}
		columns, queryErr := rows.Columns()
		if queryErr != nil {
			rows.Close()
			return queryErr
		}
		names := make([]string, len(columns))
		placeholders := make([]string, len(columns))
		for i, col := range columns {
			names[i] = `"` + strings.ReplaceAll(col, `"`, `""`) + `"`
			placeholders[i] = "?"
		}
		insert := `INSERT INTO ` + table + `(` + strings.Join(names, ",") + `) VALUES(` + strings.Join(placeholders, ",") + `)`
		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range values {
				pointers[i] = &values[i]
			}
			if err = rows.Scan(pointers...); err != nil {
				rows.Close()
				return err
			}
			if _, err = tx.Exec(insert, values...); err != nil {
				rows.Close()
				return fmt.Errorf("copy %s: %w", table, err)
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		if generatedTables[table] {
			// Names come from the fixed list above, never from user input.
			if _, err = tx.Tx.Exec(`SELECT setval(pg_get_serial_sequence('loop.` + table + `','id'),COALESCE(MAX(id),1),COUNT(*)>0) FROM loop.` + table); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
