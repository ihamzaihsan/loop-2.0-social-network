package routes

import (
	"net/http"
	"socialNetwork/pkg/db"
	"socialNetwork/pkg/db/query"
	"strconv"
	"strings"
)

func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	actor := viewer(r)
	term := strings.TrimSpace(r.URL.Query().Get("q"))
	kind := r.URL.Query().Get("type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	if page > 100000 || !validText(term, 100, true) {
		http.Error(w, "Enter 1–100 search characters", 400)
		return
	}
	if kind == "posts" {
		posts, total, err := query.VisiblePosts(actor, 0, 10, (page-1)*10, term)
		if err != nil {
			http.Error(w, "Search unavailable", 500)
			return
		}
		respond(w, map[string]interface{}{"items": posts, "total": total})
		return
	}
	var statement string
	switch kind {
	case "users":
		statement = `SELECT u.id,u.first_name||' '||u.last_name,COALESCE(u.nickname,'') FROM users u WHERE u.is_suspended=0 AND u.id!=? AND (instr(lower(u.first_name||' '||u.last_name),lower(?))>0 OR instr(lower(COALESCE(u.nickname,'')),lower(?))>0) AND NOT EXISTS(SELECT 1 FROM user_blocks b WHERE (b.blocker_id=? AND b.blocked_id=u.id) OR (b.blocker_id=u.id AND b.blocked_id=?)) ORDER BY u.first_name,u.id LIMIT 21 OFFSET ?`
	case "groups":
		statement = `SELECT g.id,g.title,COALESCE(g.description,'') FROM groups g JOIN users u ON u.id=g.creator_id WHERE u.is_suspended=0 AND ? > 0 AND (instr(lower(g.title),lower(?))>0 OR instr(lower(COALESCE(g.description,'')),lower(?))>0) AND NOT EXISTS(SELECT 1 FROM user_blocks b WHERE (b.blocker_id=? AND b.blocked_id=g.creator_id) OR (b.blocker_id=g.creator_id AND b.blocked_id=?)) ORDER BY g.id DESC LIMIT 21 OFFSET ?`
	default:
		http.Error(w, "Choose users, posts or groups", 400)
		return
	}
	rows, err := db.DBInstance.DB.Query(statement, actor, term, term, actor, actor, (page-1)*20)
	if err != nil {
		http.Error(w, "Search unavailable", 500)
		return
	}
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var title, description string
		if rows.Scan(&id, &title, &description) != nil {
			http.Error(w, "Search unavailable", 500)
			return
		}
		items = append(items, map[string]interface{}{"id": id, "title": title, "description": description})
	}
	more := len(items) > 20
	if more {
		items = items[:20]
	}
	respond(w, map[string]interface{}{"items": items, "hasMore": more})
}
