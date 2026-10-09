package routes

import (
	"database/sql"
	"net/http"
	"socialNetwork/pkg/access"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"strings"
)

func Blocks(w http.ResponseWriter, r *http.Request) {
	id := viewer(r)
	if r.Method == http.MethodGet {
		rows, err := db.DBInstance.DB.Query(`SELECT u.id,u.first_name,u.last_name FROM user_blocks b JOIN users u ON u.id=b.blocked_id WHERE b.blocker_id=? ORDER BY b.created_at DESC`, id)
		if err != nil {
			http.Error(w, "Unable to load blocked users", 500)
			return
		}
		defer rows.Close()
		result := make([]map[string]interface{}, 0)
		for rows.Next() {
			var target int
			var first, last string
			if rows.Scan(&target, &first, &last) != nil {
				http.Error(w, "Unable to load blocked users", 500)
				return
			}
			result = append(result, map[string]interface{}{"id": target, "name": first + " " + last})
		}
		respond(w, result)
		return
	}
	target := requestID(r)
	if target <= 0 || target == id {
		http.Error(w, "Invalid user", 400)
		return
	}
	if r.Method == http.MethodDelete {
		if _, err := db.DBInstance.DB.Exec(`DELETE FROM user_blocks WHERE blocker_id=? AND blocked_id=?`, id, target); err != nil {
			http.Error(w, "Unable to unblock", 500)
			return
		}
		respond(w, map[string]bool{"success": true})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		http.Error(w, "Unable to block", 500)
		return
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`INSERT OR IGNORE INTO user_blocks(blocker_id,blocked_id) VALUES (?,?)`, id, target); err != nil {
		http.Error(w, "User unavailable", 404)
		return
	}
	if _, err = tx.Exec(`DELETE FROM followers WHERE (follower_id=? AND following_id=?) OR (follower_id=? AND following_id=?)`, id, target, target, id); err != nil {
		http.Error(w, "Unable to block", 500)
		return
	}
	if _, err = tx.Exec(`DELETE FROM notifications WHERE (user_id=? AND from_user_id=?) OR (user_id=? AND from_user_id=?)`, id, target, target, id); err != nil {
		http.Error(w, "Unable to block", 500)
		return
	}
	if tx.Commit() != nil {
		http.Error(w, "Unable to block", 500)
		return
	}
	respond(w, map[string]bool{"success": true})
}

type targetInfo struct{ owner, group int }

func contentTarget(kind string, id int) (targetInfo, error) {
	var t targetInfo
	var statement string
	switch kind {
	case "user":
		statement = `SELECT id,0 FROM users WHERE id=?`
	case "post":
		statement = `SELECT user_id,0 FROM posts WHERE id=?`
	case "comment":
		statement = `SELECT user_id,0 FROM comments WHERE id=?`
	case "group_post":
		statement = `SELECT user_id,group_id FROM group_posts WHERE id=?`
	case "group_comment":
		statement = `SELECT c.user_id,p.group_id FROM group_comments c JOIN group_posts p ON p.id=c.post_id WHERE c.id=?`
	case "group_event":
		statement = `SELECT COALESCE(creator_id,(SELECT creator_id FROM groups WHERE id=group_id)),group_id FROM group_events WHERE id=?`
	default:
		return t, sql.ErrNoRows
	}
	err := db.DBInstance.DB.QueryRow(statement, id).Scan(&t.owner, &t.group)
	return t, err
}
func canReadTarget(kind string, id, user int, t targetInfo) bool {
	switch kind {
	case "user":
		return !access.Blocked(user, id)
	case "post":
		return access.CanViewPost(user, id)
	case "comment":
		var post int
		return db.DBInstance.DB.QueryRow(`SELECT post_id FROM comments WHERE id=?`, id).Scan(&post) == nil && access.CanViewPost(user, post)
	default:
		return access.GroupMember(t.group, user)
	}
}

func Reports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var input struct {
		Type   string `json:"type"`
		ID     int    `json:"id"`
		Reason string `json:"reason"`
	}
	if !decode(w, r, &input) {
		return
	}
	input.Reason = strings.TrimSpace(input.Reason)
	t, err := contentTarget(input.Type, input.ID)
	if err != nil || !canReadTarget(input.Type, input.ID, viewer(r), t) {
		http.Error(w, "Content unavailable", 404)
		return
	}
	if t.owner == viewer(r) || !validText(input.Reason, 1000, true) {
		http.Error(w, "Give a reason for reporting another user's content", 400)
		return
	}
	var open bool
	if db.DBInstance.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM reports WHERE reporter_id=? AND target_type=? AND target_id=? AND status='open')`, viewer(r), input.Type, input.ID).Scan(&open) != nil {
		http.Error(w, "Unable to submit report", 500)
		return
	}
	if open {
		http.Error(w, "You already have an open report for this content", 409)
		return
	}
	if _, err = db.DBInstance.DB.Exec(`INSERT INTO reports(reporter_id,target_type,target_id,reason) VALUES (?,?,?,?)`, viewer(r), input.Type, input.ID, input.Reason); err != nil {
		http.Error(w, "Unable to submit report", 500)
		return
	}
	respond(w, map[string]bool{"success": true})
}

func Moderation(w http.ResponseWriter, r *http.Request) {
	actor := viewer(r)
	if !auth.IsModerator(actor) {
		http.Error(w, "Moderator access required", 403)
		return
	}
	if r.Method == http.MethodGet {
		rows, err := db.DBInstance.DB.Query(`SELECT r.id,r.target_type,r.target_id,r.reason,r.status,r.resolution,r.created_at,COALESCE(u.first_name||' '||u.last_name,'Deleted account') FROM reports r LEFT JOIN users u ON u.id=r.reporter_id ORDER BY r.id DESC LIMIT 100`)
		if err != nil {
			http.Error(w, "Unable to load reports", 500)
			return
		}
		defer rows.Close()
		items := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, target int
			var kind, reason, status, resolution, created, reporter string
			if rows.Scan(&id, &kind, &target, &reason, &status, &resolution, &created, &reporter) != nil {
				http.Error(w, "Unable to load reports", 500)
				return
			}
			var preview string
			statements := map[string]string{"user": `SELECT first_name||' '||last_name FROM users WHERE id=?`, "post": `SELECT content FROM posts WHERE id=?`, "comment": `SELECT content FROM comments WHERE id=?`, "group_post": `SELECT content FROM group_posts WHERE id=?`, "group_comment": `SELECT content FROM group_comments WHERE id=?`, "group_event": `SELECT title||': '||description FROM group_events WHERE id=?`}
			if statement, ok := statements[kind]; ok {
				db.DBInstance.DB.QueryRow(statement, target).Scan(&preview)
			}
			items = append(items, map[string]interface{}{"id": id, "type": kind, "targetId": target, "reason": reason, "status": status, "resolution": resolution, "createdAt": created, "reporter": reporter, "preview": preview})
		}
		respond(w, items)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var input struct {
		ID     int    `json:"id"`
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if !decode(w, r, &input) {
		return
	}
	if !validText(input.Note, 1000, true) {
		http.Error(w, "Add a moderation note", 400)
		return
	}
	var kind string
	var target int
	if db.DBInstance.DB.QueryRow(`SELECT target_type,target_id FROM reports WHERE id=? AND status='open'`, input.ID).Scan(&kind, &target) != nil {
		http.Error(w, "Open report not found", 404)
		return
	}
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		http.Error(w, "Unable to resolve report", 500)
		return
	}
	defer tx.Rollback()
	status := "resolved"
	suspended := 0
	switch input.Action {
	case "dismiss":
		status = "dismissed"
	case "resolve":
	case "remove":
		if kind == "user" {
			http.Error(w, "Use suspend for accounts", 400)
			return
		}
		err = query.DeleteContentTx(tx, kind, target)
	case "suspend":
		t, e := contentTarget(kind, target)
		if e != nil {
			http.Error(w, "Target unavailable", 404)
			return
		}
		if t.owner == actor || auth.IsModerator(t.owner) {
			http.Error(w, "Cannot suspend a moderator", 403)
			return
		}
		suspended = t.owner
		_, err = tx.Exec(`UPDATE users SET is_suspended=1 WHERE id=?`, suspended)
	default:
		http.Error(w, "Invalid moderation action", 400)
		return
	}
	if err == nil {
		_, err = tx.Exec(`UPDATE reports SET status=?,resolution=?,resolved_by=?,resolved_at=CURRENT_TIMESTAMP WHERE id=? AND status='open'`, status, input.Action+": "+input.Note, actor, input.ID)
	}
	if err != nil || tx.Commit() != nil {
		http.Error(w, "Unable to resolve report", 500)
		return
	}
	if suspended > 0 {
		auth.RevokeUserSessions(suspended)
	}
	respond(w, map[string]bool{"success": true})
}
