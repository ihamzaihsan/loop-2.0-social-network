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

	log.Printf("[INFO] Processing notification action: notification_id=%d, user_id=%d, action=%s", notificationID, userID, req.Action)

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
		log.Printf("[ERROR] Notification not found: %d", notificationID)
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	log.Printf("[INFO] Found notification: type=%s, related_id=%d", notification.Type, notification.RelatedID)

	// Process action based on notification type
	result := map[string]interface{}{
		"success": false,
		"message": "Action not supported for this notification type",
	}

	switch notification.Type {
	case "follow_request":
		result = handleFollowRequestAction(notification, userID, req.Action)
	case "group_invitation":
		log.Printf("[INFO] Processing group invitation action")
		result = handleGroupInvitationAction(notification, userID, req.Action)
	case "group_join_request":
		result = handleGroupJoinRequestAction(notification, userID, req.Action)
	}

	log.Printf("[INFO] Action result: %+v", result)

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

		// ADD THIS: Broadcast group membership update when invitation is accepted
		go func() {
			// Get updated group info
			group, err := query.GetGroupByID(groupID, userID)
			if err != nil {
				log.Printf("[ERROR] Failed to get group info for broadcast: %v", err)
				return
			}

			// Get updated ACTIVE members list only
			members, err := query.GetActiveGroupMembers(groupID)
			if err != nil {
				log.Printf("[ERROR] Failed to get active group members for broadcast: %v", err)
				return
			}

			// Get active members count
			activeCount := len(members)

			// Update group member count
			group.MemberCount = activeCount

			// Create update message
			updateData := map[string]interface{}{
				"group_id":      groupID,
				"action":        "member_joined",
				"member_count":  activeCount,
				"members":       members,
				"group":         group,
				"new_member_id": userID,
			}

			log.Printf("[INFO] Broadcasting group membership update for group %d after invitation acceptance", groupID)

			// Import the websocket package at the top of the file if not already imported
			// Broadcast to all group members
			broadcastToGroupMembers(groupID, 0, Message{
				Type:    "group_membership_update",
				Content: updateData,
			})
		}()
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
