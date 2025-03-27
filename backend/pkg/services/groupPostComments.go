package services

import (
	"log"
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

// GetGroupPostCommentsService returns all comments for a post
func GetGroupPostCommentsService(postID, userID int) ([]models.GroupComment, error) {
	// Get post to check group membership
	post, err := query.GetGroupPost(postID)
	if err != nil {
		return nil, err
	}

	// Check if user is a member of the group
	isMember, err := query.IsGroupMember(post.GroupID, userID)
	if err != nil || !isMember {
		return nil, err
	}

	// Get comments
	comments, err := query.GetGroupPostComments(postID)
	if err != nil {
		log.Printf("[ERROR] Failed to get post comments: %v", err)
		return nil, err
	}

	return comments, nil
}
