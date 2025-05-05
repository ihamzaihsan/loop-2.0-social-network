package events

import "sync"

type EventType string

const (
	SessionInvalidated EventType = "session_invalidated"
)

type Event struct {
	Type    EventType
	UserID  int
	Payload interface{}
}

var (
	listeners   = make(map[EventType][]func(Event))
	listenersMu sync.Mutex
)

// Subscribe registers a listener for a specific event type
func Subscribe(eventType EventType, listener func(Event)) {
	listenersMu.Lock()
	defer listenersMu.Unlock()
	
	listeners[eventType] = append(listeners[eventType], listener)
}

// Publish sends an event to all registered listeners
func Publish(event Event) {
	listenersMu.Lock()
	eventListeners := listeners[event.Type]
	listenersMu.Unlock()
	
	for _, listener := range eventListeners {
		go listener(event)
	}
}