package services

import (
	"log"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/websocket"
)

// CreateGroupPostService handles creating a post in a group
func CreateGroupPostService(groupID, userID int, content, image string) (int, error) {
	// Check if user is an active member of the group
	isMember, err := query.IsGroupMember(groupID, userID)
	if err != nil || !isMember {
		return 0, err
	}

	// Create post
	postID, err := query.CreateGroupPost(groupID, userID, content, image)
	if err != nil {
		log.Printf("[ERROR] Failed to create group post: %v", err)
		return 0, err
	}

	// Broadcast complete post data to all group members via WebSocket
	websocket.BroadcastGroupPost(groupID, userID, int64(postID), content, image)

	return postID, nil
}
