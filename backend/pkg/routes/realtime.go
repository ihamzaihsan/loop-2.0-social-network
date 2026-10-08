package routes

import (
	"encoding/json"
	"net/http"
	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db"
	ws "socialNetwork/pkg/websocket"
)

var mutationResources = map[string][]string{
	"/posts": {"posts"}, "/create-post": {"posts"}, "/comments": {"comments"},
	"/profile/privacy": {"profiles", "users", "posts"}, "/register": {"users"},
	"/follow": {"social"}, "/unfollow": {"social"}, "/follow-request": {"social"},
	"/notifications/action": {"social", "groups", "notifications"},
	"/groups/create":        {"groups"}, "/groups/invite": {"groups"},
	"/groups/join/request": {"groups"}, "/groups/membership/handle": {"groups"},
	"/groups/posts/create": {"groups"}, "/groups/posts/comments/create": {"groups"},
	"/groups/events/create": {"groups"}, "/groups/events/respond": {"groups"},
}

// MutationEvents publishes invalidations after successful writes. Events carry
// no content; each viewer reloads through the normal authenticated read routes.
func MutationEvents(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete && r.Method != http.MethodPatch {
			next.ServeHTTP(w, r)
			return
		}

		changes := mutationResources[r.URL.Path]
		if len(changes) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		recorder := &mutationResponse{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		if recorder.status >= 200 && recorder.status < 300 {
			for _, resource := range changes {
				ws.PublishChange(resource)
			}
		}
	})
}

type mutationResponse struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *mutationResponse) WriteHeader(status int) {
	if !w.wrote {
		w.status, w.wrote = status, true
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *mutationResponse) Write(body []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

// MarkMessagesRead marks only messages this reader has actually fetched. The
// upper bound prevents messages arriving during the request from being read.
func MarkMessagesRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	readerID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var input struct {
		ContactID int `json:"contact_id"`
		ThroughID int `json:"through_id"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil || input.ContactID <= 0 || input.ThroughID <= 0 {
		http.Error(w, "Invalid read boundary", http.StatusBadRequest)
		return
	}
	result, err := db.DBInstance.DB.Exec(`UPDATE messages SET is_read = 1
		WHERE sender_id = ? AND id <= ? AND is_read = 0 AND chat_id IN
		(SELECT id FROM chats WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?))`,
		input.ContactID, input.ThroughID, readerID, input.ContactID, input.ContactID, readerID)
	if err != nil {
		http.Error(w, "Failed to mark messages read", http.StatusInternalServerError)
		return
	}
	if count, _ := result.RowsAffected(); count > 0 {
		message := ws.Message{Type: "messages_read", Content: map[string]int{"reader_id": readerID, "contact_id": input.ContactID, "through_id": input.ThroughID}}
		ws.SendToUser(readerID, message)
		ws.SendToUser(input.ContactID, message)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
