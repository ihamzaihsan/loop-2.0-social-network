package routes

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"socialNetwork/pkg/access"
	"socialNetwork/pkg/cloud"
	"socialNetwork/pkg/db"
	"strings"
)

func ServeMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", 405)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/uploads/")
	if name == "" || filepath.Base(name) != name || strings.ContainsAny(name, "\\/") {
		http.NotFound(w, r)
		return
	}
	path := "/uploads/" + name
	actor := viewer(r)
	allowed := false
	statements := []string{
		`SELECT EXISTS(SELECT 1 FROM media_uploads WHERE path=? AND owner_id=?)`,
		`SELECT EXISTS(SELECT 1 FROM users u WHERE u.avatar=? AND u.is_suspended=0 AND NOT EXISTS(SELECT 1 FROM user_blocks WHERE (blocker_id=? AND blocked_id=u.id) OR (blocker_id=u.id AND blocked_id=?)))`,
		`SELECT EXISTS(SELECT 1 FROM posts p JOIN users u ON u.id=p.user_id WHERE p.image=? AND ` + access.PostAudience(actor) + `)`,
		`SELECT EXISTS(SELECT 1 FROM comments c JOIN posts p ON p.id=c.post_id JOIN users u ON u.id=p.user_id WHERE c.image=? AND ` + access.PostAudience(actor) + `)`,
		`SELECT EXISTS(SELECT 1 FROM group_posts p JOIN group_members m ON m.group_id=p.group_id WHERE p.image=? AND m.user_id=? AND m.status='active')`,
		`SELECT EXISTS(SELECT 1 FROM messages m JOIN chats c ON c.id=m.chat_id WHERE instr(m.content,?)>0 AND (c.user1_id=? OR c.user2_id=?))`,
		`SELECT EXISTS(SELECT 1 FROM group_chat_messages c JOIN group_members m ON m.group_id=c.group_id WHERE instr(c.content,?)>0 AND m.user_id=? AND m.status='active')`,
	}
	for _, s := range statements {
		args := []interface{}{path}
		for i := strings.Count(s, "?"); len(args) < i; {
			args = append(args, actor)
		}
		var yes bool
		if db.DBInstance.DB.QueryRow(s, args...).Scan(&yes) == nil && yes {
			allowed = true
			break
		}
	}
	if !allowed {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if cloud.Enabled() {
		response, err := cloud.Read(r.Context(), name)
		if err != nil {
			var providerError *cloud.HTTPError
			if errors.As(err, &providerError) && providerError.Status == 404 {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "Media unavailable", http.StatusBadGateway)
			return
		}
		defer response.Body.Close()
		for _, header := range []string{"Content-Type", "Content-Length"} {
			if value := response.Header.Get(header); value != "" {
				w.Header().Set(header, value)
			}
		}
		if r.Method == http.MethodGet {
			_, _ = io.Copy(w, response.Body)
		}
		return
	}
	http.ServeFile(w, r, filepath.Join("uploads", name))
}
