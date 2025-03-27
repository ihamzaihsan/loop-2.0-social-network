package services

import (
	"log"
	"socialNetwork/pkg/db/query"
)

// GetGroupDetailsService returns detailed information about a specific group
func GetGroupDetailsService(groupID, userID int) (map[string]interface{}, error) {
	// Get group details
	group, err := query.GetGroupByID(groupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group details: %v", err)
		return nil, err
	}

	// Check if user is a member
	isMember, err := query.IsGroupMember(groupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to check group membership: %v", err)
		return nil, err
	}

	result := map[string]interface{}{
		"group":     group,
		"is_member": isMember,
	}

	// If not a member, only return basic info
	if !isMember {
		return result, nil
	}

	// Get members
	members, err := query.GetGroupMembers(groupID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group members: %v", err)
	} else {
		result["members"] = members
	}

	// Get posts
	posts, err := query.GetGroupPosts(groupID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group posts: %v", err)
	} else {
		result["posts"] = posts
	}

	// Get events
	events, err := query.GetGroupEvents(groupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group events: %v", err)
	} else {
		result["events"] = events
	}

	// Check if user is creator
	isCreator, err := query.IsGroupCreator(groupID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to check if user is creator: %v", err)
	} else {
		result["is_creator"] = isCreator
	}

	// If user is creator, get join requests
	if isCreator {
		joinRequests, err := query.GetGroupJoinRequests(groupID)
		if err != nil {
			log.Printf("[ERROR] Failed to get join requests: %v", err)
		} else {
			result["join_requests"] = joinRequests
		}
	}

	return result, nil
}
