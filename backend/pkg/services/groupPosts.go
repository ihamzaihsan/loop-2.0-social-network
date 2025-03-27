package services

import (
	"log"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

// GetGroupPostsService returns all posts in a group
func GetGroupPostsService(groupID, userID int) ([]models.GroupPost, error) {
	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(groupID, userID)
	if err != nil || !isMember {
		return nil, err
	}

	// Get posts
	posts, err := query.GetGroupPosts(groupID)
	if err != nil {
		log.Printf("[ERROR] Failed to get group posts: %v", err)
		return nil, err
	}

	return posts, nil
}
