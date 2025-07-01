package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"socialNetwork/pkg/auth"
	query "socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
	"socialNetwork/pkg/utils"
	"strconv"
	"strings"
	"time"
)

// Global service instance
var GroupServiceImpl models.GroupService

// SetGroupService allows setting the service implementation from outside
func SetGroupService(service models.GroupService) {
	GroupServiceImpl = service
}

// CreateGroup handles the creation of a new group
// In routes/groups.go
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("[ERROR] Failed to get user ID: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to parse request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Title == "" {
		http.Error(w, "Group title is required", http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, "Group description is required", http.StatusBadRequest)
		return
	}

	// Validate field lengths
	if len(req.Title) > 100 {
		http.Error(w, "Group title exceeds maximum length of 100 characters", http.StatusBadRequest)
		return
	}

	if len(req.Description) > 100 {
		http.Error(w, "Group description exceeds maximum length of 100 characters", http.StatusBadRequest)
		return
	}
	log.Printf("[INFO] Creating group with title: %s, description: %s, userID: %d", req.Title, req.Description, userID)

	// Create group using service
	groupID, err := GroupServiceImpl.CreateGroup(req.Title, req.Description, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to create group: %v", err)
		http.Error(w, "Failed to create group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"group_id": groupID,
		"message":  "Group created successfully",
	})
}

// GetUserGroups returns all groups a user is a member of
func GetUserGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get groups using service
	groups, err := GroupServiceImpl.GetUserGroups(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user groups: %v", err)
		http.Error(w, "Failed to retrieve groups", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"groups":  groups,
	})
}

// GetGroupDetails returns detailed information about a specific group
func GetGroupDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("[ERROR] Failed to get user ID: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get group ID from URL
	groupIDStr := r.URL.Query().Get("id")
	if groupIDStr == "" {
		groupIDStr = r.URL.Path
		// Extract group ID from path like /groups/1
		if len(groupIDStr) > 0 {
			groupIDStr = groupIDStr[len("/groups/"):] // Remove prefix
		}
	}

	// Remove any extra characters after the ID if present
	if idx := strings.IndexAny(groupIDStr, ":"); idx != -1 {
		groupIDStr = groupIDStr[:idx]
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		log.Printf("[ERROR] Invalid group ID: %v", err)
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(groupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to check group membership: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !isMember {
		log.Printf("[INFO] User %d attempted to access group %d without permission", userID, groupID)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get group details using service
	group, err := GroupServiceImpl.GetGroupDetails(groupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group details: %v", err)
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    group,
	})
}

// InviteToGroup handles inviting users to a group
func InviteToGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		GroupID int   `json:"group_id"`
		UserIDs []int `json:"user_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Invite users using service
	err = GroupServiceImpl.InviteToGroup(req.GroupID, userID, req.UserIDs)
	if err != nil {
		switch err {
		case query.ErrUserAlreadyInvited:
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "One or more users already have pending invitations",
			})
			return
		case query.ErrUserAlreadyMember:
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "One or more users are already members of the group",
			})
			return
		case sql.ErrNoRows:
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "All selected users either already have pending invitations or are already members",
			})
			return
		default:
			http.Error(w, "Failed to send invitations", http.StatusInternalServerError)
			return
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Invitations sent successfully",
	})
}

// GetGroupInvitations returns all pending group invitations for a user
func GetGroupInvitations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get invitations using service
	invitations, err := GroupServiceImpl.GetGroupInvitations(userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group invitations: %v", err)
		http.Error(w, "Failed to retrieve invitations", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"invitations": invitations,
	})
}

// RequestToJoinGroup handles user requests to join a group
func RequestToJoinGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		GroupID int `json:"group_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Request to join using service
	err = GroupServiceImpl.RequestToJoinGroup(req.GroupID, userID)
	if err != nil {
		http.Error(w, "Failed to request joining group", http.StatusInternalServerError)
		return
	}

	// CREATE NOTIFICATION FOR GROUP OWNER
	go func() {
		// Get group details to find the creator
		group, err := query.GetGroupByID(req.GroupID, userID)
		if err != nil {
			log.Printf("[ERROR] Failed to get group details for notification: %v", err)
			return
		}

		log.Printf("[INFO] Creating join request notification for group owner %d", group.CreatorID)

		// Create notification directly using query package
		notificationID, err := query.CreateNotification(
			group.CreatorID, // recipient (group owner)
			userID,          // sender (user requesting to join)
			"group_join_request",
			req.GroupID, // related_id (group ID)
			fmt.Sprintf("User wants to join your group '%s'", group.Title),
		)

		if err != nil {
			log.Printf("[ERROR] Failed to create join request notification: %v", err)
			return
		}

		log.Printf("[INFO] Created join request notification with ID: %d", notificationID)

		// Send real-time WebSocket notification
		// Get sender details for rich notification
		var senderFirstName, senderLastName string
		err = query.GetUserNames(userID, &senderFirstName, &senderLastName)
		if err != nil {
			log.Printf("[ERROR] Failed to get sender name: %v", err)
			senderFirstName = "Someone"
			senderLastName = ""
		}

		// Create notification object for WebSocket
		notification := map[string]interface{}{
			"id":           notificationID,
			"user_id":      group.CreatorID,
			"from_user_id": userID,
			"type":         "group_join_request",
			"related_id":   req.GroupID,
			"content":      fmt.Sprintf("User wants to join your group '%s'", group.Title),
			"status":       "unread",
			"sender_name":  fmt.Sprintf("%s %s", senderFirstName, senderLastName),
			"group_title":  group.Title,
			"actions":      []string{"accept", "reject"},
			"created_at":   time.Now(),
		}

		// Send WebSocket notification to group owner
		success := SendToUser(group.CreatorID, Message{
			Type:    "notification",
			Content: notification,
		})

		if success {
			log.Printf("[INFO] Sent real-time notification to group owner %d", group.CreatorID)
		} else {
			log.Printf("[INFO] Group owner %d is offline, notification stored for later", group.CreatorID)
		}
	}()

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Join request sent successfully",
	})
}

// HandleGroupMembershipRequest processes a group invitation or join request
func HandleGroupMembershipRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	// Parse request body
	var req struct {
		GroupID     int    `json:"group_id"`
		UserID      int    `json:"user_id,omitempty"`
		Action      string `json:"action"`
		RequestType string `json:"request_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Handling %s %s: group_id=%d, user_id=%d",
		req.Action, req.RequestType, req.GroupID, userID)

	targetUserID := userID
	if req.RequestType == "request" && req.UserID > 0 {
		targetUserID = req.UserID
	}

	// Handle the request
	err = query.HandleGroupMembershipRequest(req.GroupID, targetUserID, req.Action, req.RequestType)
	if err != nil {
		log.Printf("[ERROR] Failed to handle group membership request: %v", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// If request was accepted, broadcast membership update
	if req.Action == "accept" {
		// Get updated group info
		group, err := query.GetGroupByID(req.GroupID, userID)
		if err == nil {
			// Get updated ACTIVE members list only
			members, err := query.GetActiveGroupMembers(req.GroupID)
			if err == nil {
				// Get active members count
				activeCount := len(members)

				// Update group member count
				group.MemberCount = activeCount

				// Create update message
				updateData := map[string]interface{}{
					"group_id":      req.GroupID,
					"action":        "member_joined",
					"member_count":  activeCount,
					"members":       members,
					"group":         group,
					"new_member_id": targetUserID,
				}

				// Broadcast to all group members
				broadcastToGroupMembers(req.GroupID, 0, Message{
					Type:    "group_membership_update",
					Content: updateData,
				})
			}
		}

		// Send notification if request was accepted
		recipientID := targetUserID
		if req.RequestType == "invitation" {
			// Get group creator ID to notify them
			if group.CreatorID > 0 {
				recipientID = group.CreatorID

				// Send WebSocket notification to group creator
				SendToUser(recipientID, Message{
					Type: "group_member_joined",
					Content: map[string]interface{}{
						"group_id": req.GroupID,
						"user_id":  userID,
						"message":  "A user has accepted your invitation to join the group",
					},
				})
			}
		} else if req.RequestType == "request" {
			// Send WebSocket notification to the user who requested to join
			SendToUser(targetUserID, Message{
				Type: "group_join_approved",
				Content: map[string]interface{}{
					"group_id": req.GroupID,
					"message":  "Your request to join the group has been approved",
				},
			})
		}
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// CreateGroupPost handles creating a post in a group
func CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to parse as multipart form first (for file uploads)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// If parsing multipart form fails, try to parse as JSON
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			var jsonRequest struct {
				GroupID int    `json:"group_id"`
				Content string `json:"content"`
				Image   string `json:"image,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&jsonRequest); err != nil {
				http.Error(w, "Invalid request format", http.StatusBadRequest)
				return
			}

			// Validate content length
			if len(jsonRequest.Content) > 100 {
				http.Error(w, "Post content exceeds maximum length of 100 characters", http.StatusBadRequest)
				return
			}

			// Validate that post has content
			if strings.TrimSpace(jsonRequest.Content) == "" {
				http.Error(w, "Post must contain text content", http.StatusBadRequest)
				return
			}

			// Check if user is a member of the group
			isMember, err := query.IsGroupMember(jsonRequest.GroupID, userID)
			if err != nil {
				log.Printf("[ERROR] Failed to check group membership: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !isMember {
				log.Printf("[INFO] User %d attempted to post in group %d without permission", userID, jsonRequest.GroupID)
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			// Create post using service
			postID, err := GroupServiceImpl.CreateGroupPost(jsonRequest.GroupID, userID, jsonRequest.Content, jsonRequest.Image)
			if err != nil {
				http.Error(w, "Failed to create post", http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"post_id": postID,
				"image":   jsonRequest.Image,
			})
			return
		}

		http.Error(w, "Invalid request format or file too large", http.StatusBadRequest)
		return
	}

	// Handle multipart form (with potential file upload)
	var req struct {
		GroupID int    `json:"group_id"`
		Content string `json:"content"`
		Image   string `json:"image,omitempty"`
	}

	// Parse form values
	groupIDStr := r.FormValue("group_id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	req.GroupID = groupID
	req.Content = r.FormValue("content")

	// Validate content length
	if len(req.Content) > 100 {
		http.Error(w, "Post content exceeds maximum length of 100 characters", http.StatusBadRequest)
		return
	}

	// Check if image will be provided
	_, _, imageErr := r.FormFile("image")
	hasImage := imageErr == nil

	// Validate that post has content or image
	if strings.TrimSpace(req.Content) == "" && !hasImage {
		http.Error(w, "Post must contain either text content or an image", http.StatusBadRequest)
		return
	}

	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(req.GroupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to check group membership: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !isMember {
		log.Printf("[INFO] User %d attempted to post in group %d without permission", userID, req.GroupID)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Handle image upload if present
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imagePath, err := utils.HandleImageUpload(file, header)
		if err != nil {
			log.Printf("[ERROR] Failed to upload image: %v", err)
			http.Error(w, fmt.Sprintf("Failed to upload image: %v", err), http.StatusInternalServerError)
			return
		}
		req.Image = imagePath
	} else if err != http.ErrMissingFile {
		// Only log if it's not just a missing file
		log.Printf("[ERROR] Error processing image upload: %v", err)
		http.Error(w, "Error processing image upload", http.StatusBadRequest)
		return
	}

	// Create post using service
	postID, err := GroupServiceImpl.CreateGroupPost(req.GroupID, userID, req.Content, req.Image)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"post_id": postID,
		"image":   req.Image,
		"message": "Post created successfully",
	})
}

// GetGroupPosts returns all posts in a group
func GetGroupPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get group ID from query parameters
	groupIDStr := r.URL.Query().Get("group_id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get posts using service
	posts, err := GroupServiceImpl.GetGroupPosts(groupID, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"posts":   posts,
	})
}

// GetGroupPostComments returns all comments for a post
func GetGroupPostComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get post ID from query parameters
	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get comments using service
	comments, err := GroupServiceImpl.GetGroupPostComments(postID, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"comments": comments,
	})
}

// CreateGroupComment adds a comment to a group post
func CreateGroupComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		PostID  int    `json:"post_id"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create comment using service
	commentID, err := GroupServiceImpl.CreateGroupComment(req.PostID, userID, req.Content)
	if err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"comment_id": commentID,
		"message":    "Comment created successfully",
	})
}

// CreateGroupEvent creates a new event in a group
func CreateGroupEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		GroupID     int       `json:"group_id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		EventTime   time.Time `json:"event_time"`
		Options     []string  `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Title == "" {
		http.Error(w, "Event title is required", http.StatusBadRequest)
		return
	}

	if len(req.Options) < 2 {
		http.Error(w, "At least two response options are required", http.StatusBadRequest)
		return
	}

	// Create event using service
	eventID, err := GroupServiceImpl.CreateGroupEvent(req.GroupID, userID, req.Title, req.Description, req.EventTime)
	if err != nil {
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"event_id": eventID,
		"message":  "Event created successfully",
	})
}

// GetGroupEvents returns all events in a group
func GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get group ID from query parameters
	groupIDStr := r.URL.Query().Get("group_id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get events using service
	events, err := GroupServiceImpl.GetGroupEvents(groupID, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve events", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"events":  events,
	})
}

// GetGroupEvent returns a specific event by ID
func GetGroupEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get event ID from query parameters
	eventIDStr := r.URL.Query().Get("event_id")
	if eventIDStr == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get event using service
	event, responses, err := GroupServiceImpl.GetGroupEvent(eventID, userID)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"event":     event,
		"responses": responses,
	})
}

// RespondToEvent records a user's response to an event
// RespondToEvent records a user's response to an event
func RespondToEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Method not allowed: %s", r.Method)
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

	// Parse request body
	var req struct {
		EventID  int `json:"event_id"`
		OptionID int `json:"option_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Responding to event: event_id=%d, user_id=%d, option_id=%d", req.EventID, userID, req.OptionID)

	// Record response using service
	err = GroupServiceImpl.RespondToEvent(req.EventID, userID, req.OptionID)
	if err != nil {
		log.Printf("[ERROR] Failed to record response: %v", err)
		http.Error(w, "Failed to record response", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] Response recorded successfully")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Response recorded successfully",
	})
}

// GetGroupMessages returns all messages for a specific group
func GetGroupMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get group ID from query parameters
	groupIDStr := r.URL.Query().Get("id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if user is a member of the group
	isMember, err := GroupServiceImpl.IsGroupMember(groupID, userID)
	if err != nil {
		http.Error(w, "Failed to check group membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		http.Error(w, "You are not a member of this group", http.StatusForbidden)
		return
	}

	// Get messages using service
	messages, err := GroupServiceImpl.GetGroupMessages(groupID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group messages: %v", err)
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}
