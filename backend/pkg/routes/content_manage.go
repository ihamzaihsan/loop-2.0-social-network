package routes

import (
	"net/http"
	"socialNetwork/pkg/access"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"time"
)

func ManageContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", 405)
		return
	}
	kind, id := r.URL.Query().Get("type"), requestID(r)
	t, err := contentTarget(kind, id)
	actor := viewer(r)
	if err != nil || kind == "user" || !canReadTarget(kind, id, actor, t) {
		http.Error(w, "Content unavailable", 404)
		return
	}
	if t.owner != actor && !(t.group > 0 && access.GroupCreator(t.group, actor)) {
		http.Error(w, "Only the author or group owner can manage this content", 403)
		return
	}
	if r.Method == http.MethodDelete {
		tx, err := db.DBInstance.DB.Begin()
		if err != nil {
			http.Error(w, "Unable to delete content", 500)
			return
		}
		defer tx.Rollback()
		if query.DeleteContentTx(tx, kind, id) != nil || tx.Commit() != nil {
			http.Error(w, "Unable to delete content", 500)
			return
		}
		respond(w, map[string]bool{"success": true})
		return
	}
	var input struct {
		Content     string    `json:"content"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		EventTime   time.Time `json:"eventTime"`
	}
	if !decode(w, r, &input) {
		return
	}
	switch kind {
	case "comment", "group_comment", "group_post":
		if !validText(input.Content, 100, true) {
			http.Error(w, "Content must contain 1–100 characters", 400)
			return
		}
		table := map[string]string{"comment": "comments", "group_comment": "group_comments", "group_post": "group_posts"}[kind]
		_, err = db.DBInstance.DB.Exec(`UPDATE `+table+` SET content=? WHERE id=?`, input.Content, id)
	case "group_event":
		if !validText(input.Title, 100, true) || !validText(input.Description, 500, false) || input.EventTime.IsZero() {
			http.Error(w, "Check event title, description and time", 400)
			return
		}
		_, err = db.DBInstance.DB.Exec(`UPDATE group_events SET title=?,description=?,event_time=? WHERE id=?`, input.Title, input.Description, input.EventTime, id)
	default:
		http.Error(w, "Unsupported content type", 400)
		return
	}
	if err != nil {
		http.Error(w, "Unable to save content", 500)
		return
	}
	respond(w, map[string]bool{"success": true})
}

func PostLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", 405)
		return
	}
	id, actor := requestID(r), viewer(r)
	if !access.CanViewPost(actor, id) {
		http.Error(w, "Post unavailable", 404)
		return
	}
	var err error
	if r.Method == http.MethodPut {
		_, err = db.DBInstance.DB.Exec(`INSERT OR IGNORE INTO likes(post_id,user_id) VALUES (?,?)`, id, actor)
	} else {
		_, err = db.DBInstance.DB.Exec(`DELETE FROM likes WHERE post_id=? AND user_id=?`, id, actor)
	}
	if err != nil {
		http.Error(w, "Unable to update like", 500)
		return
	}
	var count int
	var liked bool
	if db.DBInstance.DB.QueryRow(`SELECT COUNT(*),EXISTS(SELECT 1 FROM likes WHERE post_id=? AND user_id=?) FROM likes WHERE post_id=?`, id, actor, id).Scan(&count, &liked) != nil {
		http.Error(w, "Unable to load likes", 500)
		return
	}
	respond(w, map[string]interface{}{"likeCount": count, "isLiked": liked})
}

func ManageGroup(w http.ResponseWriter, r *http.Request) {
	id, actor := requestID(r), viewer(r)
	if !access.GroupMember(id, actor) {
		http.Error(w, "Group unavailable", 404)
		return
	}
	owner := access.GroupCreator(id, actor)
	if r.Method == http.MethodPut {
		if !owner {
			http.Error(w, "Group owner access required", 403)
			return
		}
		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if !decode(w, r, &input) {
			return
		}
		if !validText(input.Title, 100, true) || !validText(input.Description, 100, false) {
			http.Error(w, "Check group title and description", 400)
			return
		}
		if _, err := db.DBInstance.DB.Exec(`UPDATE groups SET title=?,description=? WHERE id=?`, input.Title, input.Description, id); err != nil {
			http.Error(w, "Unable to save group", 500)
			return
		}
		respond(w, map[string]bool{"success": true})
		return
	}
	if r.Method == http.MethodDelete {
		if !owner {
			http.Error(w, "Group owner access required", 403)
			return
		}
		tx, err := db.DBInstance.DB.Begin()
		if err != nil {
			http.Error(w, "Unable to delete group", 500)
			return
		}
		defer tx.Rollback()
		if query.DeleteGroupTx(tx, id) != nil || tx.Commit() != nil {
			http.Error(w, "Unable to delete group", 500)
			return
		}
		respond(w, map[string]bool{"success": true})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var input struct {
		Action string `json:"action"`
		UserID int    `json:"userId"`
	}
	if !decode(w, r, &input) {
		return
	}
	target := input.UserID
	if input.Action == "leave" {
		target = actor
		if owner {
			http.Error(w, "Group owners must delete the group instead of leaving", 400)
			return
		}
	} else if input.Action != "remove" || !owner || target == actor || target <= 0 {
		http.Error(w, "Invalid member action", 403)
		return
	}
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		http.Error(w, "Unable to update membership", 500)
		return
	}
	defer tx.Rollback()
	for _, s := range []string{`DELETE FROM group_members WHERE group_id=? AND user_id=?`, `DELETE FROM group_join_requests WHERE group_id=? AND user_id=?`, `DELETE FROM group_invitations WHERE group_id=? AND invitee_id=?`, `DELETE FROM notifications WHERE related_id=? AND user_id=? AND type='group_invitation'`} {
		if _, err = tx.Exec(s, id, target); err != nil {
			http.Error(w, "Unable to update membership", 500)
			return
		}
	}
	if tx.Commit() != nil {
		http.Error(w, "Unable to update membership", 500)
		return
	}
	respond(w, map[string]bool{"success": true})
}
