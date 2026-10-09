package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

// Store keeps the existing SQLite API usable locally while adapting parameters
// and generated IDs to PostgreSQL. Application queries remain parameterized.
type Store struct {
	*sql.DB
	postgres bool
}

type Tx struct {
	*sql.Tx
	postgres bool
}

func IsPostgres() bool { return DBInstance.DB != nil && DBInstance.DB.postgres }

func prepare(query string, args []any, postgres bool) (string, []any) {
	if !postgres {
		return query, args
	}
	var result strings.Builder
	quoted := false
	parameter := 0
	previousWord := ""
	for i := 0; i < len(query); i++ {
		ch := query[i]
		if !quoted && ((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_') {
			end := i + 1
			for end < len(query) && ((query[end] >= 'a' && query[end] <= 'z') || (query[end] >= 'A' && query[end] <= 'Z') || query[end] == '_' || (query[end] >= '0' && query[end] <= '9')) {
				end++
			}
			word := query[i:end]
			lower := strings.ToLower(word)
			tablePosition := previousWord == "FROM" || previousWord == "JOIN" || previousWord == "INTO" || previousWord == "UPDATE"
			if (tablePosition && applicationTables[lower]) || lower == "instr" || lower == "julianday" || lower == "take_rate_limit" {
				result.WriteString("loop.")
			}
			result.WriteString(word)
			previousWord = strings.ToUpper(word)
			i = end - 1
			continue
		}
		if ch == '\'' {
			if quoted && i+1 < len(query) && query[i+1] == '\'' {
				result.WriteString("''")
				i++
				continue
			}
			quoted = !quoted
		}
		if ch == '?' && !quoted {
			parameter++
			fmt.Fprintf(&result, "$%d", parameter)
		} else {
			result.WriteByte(ch)
		}
	}
	values := append([]any(nil), args...)
	for i, value := range values {
		switch v := value.(type) {
		case bool:
			if v {
				values[i] = 1
			} else {
				values[i] = 0
			}
		case []byte:
			values[i] = string(v)
		}
	}
	return result.String(), values
}

// These tables use generated IDs. PostgreSQL returns them on the same pooled
// connection that performed the insert; SELECT lastval() would be unsafe.
var insertTable = regexp.MustCompile(`(?i)^\s*INSERT\s+INTO\s+(?:loop\.)?([a-z_]+)`)
var generatedTables = map[string]bool{
	"user_ids": true, "users": true, "followers": true, "posts": true, "comments": true,
	"likes": true, "groups": true, "group_members": true, "group_posts": true,
	"group_events": true, "event_response_options": true, "event_responses": true,
	"chats": true, "messages": true, "group_chat_messages": true, "notifications": true,
	"post_viewers": true, "group_invitations": true, "group_join_requests": true,
	"group_comments": true, "reports": true,
}

var applicationTables = func() map[string]bool {
	tables := map[string]bool{}
	for table := range generatedTables {
		tables[table] = true
	}
	for _, table := range []string{"sessions", "media_uploads", "password_resets", "user_blocks", "oauth_identities", "oauth_flows", "rate_limits", "presence", "realtime_outbox"} {
		tables[table] = true
	}
	return tables
}()

type insertResult struct{ id, count int64 }

func (r insertResult) LastInsertId() (int64, error) { return r.id, nil }
func (r insertResult) RowsAffected() (int64, error) { return r.count, nil }

type executor interface {
	Exec(string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
}

func execute(e executor, postgres bool, query string, args ...any) (sql.Result, error) {
	query, args = prepare(query, args, postgres)
	match := insertTable.FindStringSubmatch(query)
	if !postgres || len(match) == 0 || !generatedTables[strings.ToLower(match[1])] || strings.Contains(strings.ToUpper(query), "RETURNING") {
		return e.Exec(query, args...)
	}
	rows, err := e.Query(strings.TrimSuffix(strings.TrimSpace(query), ";")+" RETURNING id", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := insertResult{}
	for rows.Next() {
		if err = rows.Scan(&result.id); err != nil {
			return nil, err
		}
		result.count++
	}
	return result, rows.Err()
}

func (s *Store) Exec(q string, args ...any) (sql.Result, error) {
	return execute(s.DB, s.postgres, q, args...)
}
func (s *Store) Query(q string, args ...any) (*sql.Rows, error) {
	q, args = prepare(q, args, s.postgres)
	return s.DB.Query(q, args...)
}
func (s *Store) QueryRow(q string, args ...any) *sql.Row {
	q, args = prepare(q, args, s.postgres)
	return s.DB.QueryRow(q, args...)
}
func (s *Store) Begin() (*Tx, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, postgres: s.postgres}, nil
}
func (t *Tx) Exec(q string, args ...any) (sql.Result, error) {
	return execute(t.Tx, t.postgres, q, args...)
}
func (t *Tx) Query(q string, args ...any) (*sql.Rows, error) {
	q, args = prepare(q, args, t.postgres)
	return t.Tx.Query(q, args...)
}
func (t *Tx) QueryRow(q string, args ...any) *sql.Row {
	q, args = prepare(q, args, t.postgres)
	return t.Tx.QueryRow(q, args...)
}

// Serialize rotation across instances as well as across local goroutines.
func (t *Tx) LockUser(id int) error {
	if !t.postgres {
		return nil
	}
	var found int
	return t.QueryRow(`SELECT id FROM users WHERE id=? FOR UPDATE`, id).Scan(&found)
}
