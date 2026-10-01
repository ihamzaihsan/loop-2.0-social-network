package services

import (
	"log"
	"socialNetwork/pkg/events"
)

// Initialize subscribes to events when the package is loaded
func init() {
	// Subscribe to session invalidation events
	events.Subscribe(events.SessionInvalidated, handleSessionInvalidation)
}

// handleSessionInvalidation processes session invalidation events
func handleSessionInvalidation(event events.Event) {
	userID := event.UserID

	message := Message{
		Type: "session_invalidated",
		Content: map[string]interface{}{
			"message": "Your session has been invalidated due to login from another device",
		},
	}

	// Send the invalidation message to the user
	SendToUser(userID, message)

	log.Printf("[INFO] Sent session invalidation message to user %d", userID)
}
