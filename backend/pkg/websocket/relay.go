package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"os"
	"socialNetwork/pkg/cloud"
	"socialNetwork/pkg/db"
	"sync"
	"time"

	"github.com/google/uuid"
	gorilla "github.com/gorilla/websocket"
)

// Only server credentials can join this private channel. Browser connections
// keep the existing HttpOnly cookie authentication and Go access checks.
const relayTopic = "loop:server-events"

var instanceID = uuid.NewString()
var relayState struct {
	sync.RWMutex
	ready bool
}

type relayEvent struct {
	Origin  string  `json:"origin"`
	UserID  int     `json:"user_id"`
	Message Message `json:"message"`
}

func WaitForRelay(ctx context.Context) error {
	if !cloud.Enabled() {
		return nil
	}
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		relayState.RLock()
		ready := relayState.ready
		relayState.RUnlock()
		if ready {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return errors.New("realtime service unavailable")
		case <-ticker.C:
		}
	}
}

func StartRelay(ctx context.Context) {
	if !cloud.Enabled() {
		return
	}
	go func() {
		for ctx.Err() == nil {
			if err := subscribeRelay(ctx); err != nil {
				log.Print("Supabase realtime disconnected; reconnecting")
			}
			relayState.Lock()
			relayState.ready = false
			relayState.Unlock()
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
}

func subscribeRelay(ctx context.Context) error {
	endpoint, err := url.Parse(os.Getenv("SUPABASE_URL"))
	if err != nil {
		return err
	}
	endpoint.Scheme = "wss"
	endpoint.Path = "/realtime/v1/websocket"
	params := url.Values{"apikey": {os.Getenv("SUPABASE_SERVICE_ROLE_KEY")}, "vsn": {"1.0.0"}}
	endpoint.RawQuery = params.Encode()
	conn, _, err := gorilla.DefaultDialer.DialContext(ctx, endpoint.String(), nil)
	if err != nil {
		return errors.New("realtime connection failed")
	}
	defer conn.Close()
	conn.SetReadLimit(2 << 20)
	topic := "realtime:" + relayTopic
	join := map[string]any{"topic": topic, "event": "phx_join", "ref": "1", "payload": map[string]any{
		"access_token": os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		"config":       map[string]any{"private": true, "broadcast": map[string]any{"ack": false, "self": true}, "presence": map[string]any{"enabled": false}, "postgres_changes": []any{}},
	}}
	if err = conn.WriteJSON(join); err != nil {
		return err
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				conn.Close()
				return
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if conn.WriteJSON(map[string]any{"topic": "phoenix", "event": "heartbeat", "payload": map[string]any{}, "ref": uuid.NewString()}) != nil {
					conn.Close()
					return
				}
				flushOutbox()
			}
		}
	}()
	for {
		conn.SetReadDeadline(time.Now().Add(70 * time.Second))
		var frame struct {
			Event, Ref string
			Payload    json.RawMessage
		}
		if err = conn.ReadJSON(&frame); err != nil {
			return err
		}
		if frame.Event == "phx_reply" && frame.Ref == "1" {
			var reply struct{ Status string }
			_ = json.Unmarshal(frame.Payload, &reply)
			if reply.Status != "ok" {
				return errors.New("private channel authorization failed")
			}
			relayState.Lock()
			relayState.ready = true
			relayState.Unlock()
			// State is durable; refresh pages whenever their relay subscription returns.
			broadcastLocal(Message{Type: "resource_changed", Content: map[string]string{"resource": "all"}})
		}
		if frame.Event == "phx_error" || frame.Event == "phx_close" {
			return errors.New("realtime channel closed")
		}
		if frame.Event != "broadcast" {
			continue
		}
		var broadcast struct {
			Event   string
			Payload relayEvent
		}
		if json.Unmarshal(frame.Payload, &broadcast) != nil || broadcast.Event != "loop_event" || broadcast.Payload.Origin == instanceID {
			continue
		}
		event := broadcast.Payload
		if event.UserID == 0 {
			broadcastLocal(event.Message)
		} else {
			sendLocal(event.UserID, event.Message)
		}
	}
}

func publishRelay(userID int, message Message) {
	if !cloud.Enabled() {
		return
	}
	payload, err := json.Marshal(relayEvent{Origin: instanceID, UserID: userID, Message: message})
	if err != nil {
		return
	}
	if _, err = db.DBInstance.DB.Exec(`INSERT INTO realtime_outbox(payload) VALUES(?)`, string(payload)); err != nil {
		log.Print("Failed to queue realtime update")
		return
	}
	flushOutbox()
}

// The transactional outbox survives an instance shutdown or a provider outage.
// SKIP LOCKED lets multiple instances drain it without sending the same row.
func flushOutbox() {
	tx, err := db.DBInstance.DB.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	rows, err := tx.Query(`SELECT id,payload FROM realtime_outbox ORDER BY id LIMIT 20 FOR UPDATE SKIP LOCKED`)
	if err != nil {
		return
	}
	var ids []int64
	var messages []any
	for rows.Next() {
		var id int64
		var payload string
		if rows.Scan(&id, &payload) != nil {
			rows.Close()
			return
		}
		ids = append(ids, id)
		messages = append(messages, map[string]any{"topic": relayTopic, "event": "loop_event", "payload": json.RawMessage(payload), "private": true})
	}
	err = rows.Err()
	rows.Close()
	if err != nil || len(ids) == 0 {
		return
	}
	body, err := json.Marshal(map[string]any{"messages": messages})
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := cloud.Request(ctx, "POST", "/realtime/v1/api/broadcast", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Print("Realtime delivery pending; queued update will retry")
		return
	}
	response.Body.Close()
	for _, id := range ids {
		if _, err = tx.Exec(`DELETE FROM realtime_outbox WHERE id=?`, id); err != nil {
			return
		}
	}
	_ = tx.Commit()
}
