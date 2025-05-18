package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"socialNetwork/pkg/auth"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

// ServeNotifications handles GET requests to retrieve a user's notifications
func ServeNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("[ERROR] Unauthorized: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get limit parameter, default to 20
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 20
		}
	}

	// Get notifications directly from the query package
	notifications, err := query.GetUserNotifications(userID, limit)
	if err != nil {
		log.Printf("[ERROR] Failed to get notifications: %v", err)
		http.Error(w, "Failed to retrieve notifications", http.StatusInternalServerError)
		return
	}

	// Get unread count
	unreadCount, err := query.GetUnreadNotificationCount(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get unread count: %v", err)
		unreadCount = 0
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"notifications": notifications,
		"unread_count":  unreadCount,
	})
}

// MarkNotificationAsRead handles POST requests to mark a notification as read
func MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("[ERROR] Unauthorized: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get notification ID from URL params
	notificationIDStr := r.URL.Query().Get("id")
	if notificationIDStr == "" {
		http.Error(w, "Missing notification ID", http.StatusBadRequest)
		return
	}

	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	// Mark notification as read using query package directly
	err = query.MarkNotificationAsRead(notificationID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to mark notification as read: %v", err)
		http.Error(w, "Failed to update notification", http.StatusInternalServerError)
		return
	}

	// Get updated unread count
	unreadCount, err := query.GetUnreadNotificationCount(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get unread count: %v", err)
		unreadCount = 0
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "Notification marked as read",
		"unread_count": unreadCount,
	})
}

// MarkAllNotificationsAsRead handles POST requests to mark all notifications as read
func MarkAllNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("[ERROR] Unauthorized: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Mark all notifications as read using query package directly
	err = query.MarkAllNotificationsAsRead(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to mark all notifications as read: %v", err)
		http.Error(w, "Failed to update notifications", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "All notifications marked as read",
		"unread_count": 0,
	})
}

// GetNotificationCount handles GET requests to retrieve a user's unread notification count
func GetNotificationCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("[ERROR] Unauthorized: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get unread notification count using query package directly
	unreadCount, err := query.GetUnreadNotificationCount(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get unread count: %v", err)
		http.Error(w, "Failed to retrieve notification count", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"unread_count": unreadCount,
	})
}

// HandleNotificationAction handles POST requests to perform actions on notifications (accept/reject)
func HandleNotificationAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("[ERROR] Unauthorized: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get notification ID from URL params
	notificationIDStr := r.URL.Query().Get("id")
	if notificationIDStr == "" {
		http.Error(w, "Missing notification ID", http.StatusBadRequest)
		return
	}

	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Action string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get notification to determine type
	notifications, err := query.GetUserNotifications(userID, 100)
	if err != nil {
		log.Printf("[ERROR] Failed to get notifications: %v", err)
		http.Error(w, "Failed to process action", http.StatusInternalServerError)
		return
	}

	// Find the notification
	var notification *models.Notification
	for i := range notifications {
		if notifications[i].ID == notificationID {
			notification = &notifications[i]
			break
		}
	}

	if notification == nil {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	// Process action based on notification type
	result := map[string]interface{}{
		"success": false,
		"message": "Action not supported for this notification type",
	}

	switch notification.Type {
	case "follow_request":
		result = handleFollowRequestAction(notification, userID, req.Action)
	case "group_invitation":
		result = handleGroupInvitationAction(notification, userID, req.Action)
	case "group_join_request":
		result = handleGroupJoinRequestAction(notification, userID, req.Action)
	}

	// Mark notification as read after processing action
	query.MarkNotificationAsRead(notificationID, userID)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Helper functions for notification actions

func handleFollowRequestAction(notification *models.Notification, userID int, action string) map[string]interface{} {
	log.Printf("[INFO] Handling follow request action: %s", action)

	requestID := notification.RelatedID
	followerID := notification.FromUserID

	result := map[string]interface{}{
		"success": false,
		"message": "Failed to process follow request",
	}

	if action == "accept" {
		_, err := query.AcceptFollowRequest(requestID, uint(userID))
		if err != nil {
			log.Printf("[ERROR] Failed to accept follow request from user %d: %v", followerID, err)
			return result
		}
		result["success"] = true
		result["message"] = "Follow request accepted"
	} else if action == "reject" {
		err := query.RejectFollowRequest(requestID, uint(userID))
		if err != nil {
			log.Printf("[ERROR] Failed to reject follow request: %v", err)
			return result
		}
		result["success"] = true
		result["message"] = "Follow request rejected"
	}

	return result
}

func handleGroupInvitationAction(notification *models.Notification, userID int, action string) map[string]interface{} {
	log.Printf("[INFO] Handling group invitation action: %s", action)

	groupID := notification.RelatedID

	result := map[string]interface{}{
		"success": false,
		"message": "Failed to process group invitation",
	}

	// Request type is "invitation" for group invitations
	err := query.HandleGroupMembershipRequest(groupID, userID, action, "invitation")
	if err != nil {
		log.Printf("[ERROR] Failed to handle group invitation: %v", err)
		return result
	}

	result["success"] = true
	if action == "accept" {
		result["message"] = "Group invitation accepted"
	} else {
		result["message"] = "Group invitation rejected"
	}

	return result
}

func handleGroupJoinRequestAction(notification *models.Notification, userID int, action string) map[string]interface{} {
	log.Printf("[INFO] Handling group join request action: %s", action)

	groupID := notification.RelatedID
	requesterID := notification.FromUserID

	result := map[string]interface{}{
		"success": false,
		"message": "Failed to process join request",
	}

	// Check if the user is the group creator
	isCreator, err := query.IsGroupCreator(groupID, userID)
	if err != nil || !isCreator {
		result["message"] = "Only the group creator can process join requests"
		return result
	}

	// Request type is "request" for join requests
	err = query.HandleGroupMembershipRequest(groupID, requesterID, action, "request")
	if err != nil {
		log.Printf("[ERROR] Failed to handle group join request: %v", err)
		return result
	}

	result["success"] = true
	if action == "accept" {
		result["message"] = "Join request accepted"
	} else {
		result["message"] = "Join request rejected"
	}

	return result
}
