package services

import (
	"socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
	"time"
)

type GroupService struct{}

func NewGroupService() *GroupService {
	return &GroupService{}
}

// Update method names to match what's called in routes/groups.go
func (s *GroupService) CreateGroup(title, description string, userID int) (int, error) {
	return query.CreateGroup(title, description, userID)
}

func (s *GroupService) GetUserGroups(userID int) (interface{}, error) {
	return query.GetUserGroups(userID)
}

func (s *GroupService) GetGroupDetails(groupID, userID int) (interface{}, error) {
	return GetGroupDetailsService(groupID, userID) 
}

func (s *GroupService) InviteToGroup(groupID, inviterID int, userIDs []int) error {
	return InviteToGroupService(groupID, inviterID, userIDs) 
}

func (s *GroupService) GetGroupInvitations(userID int) (interface{}, error) {
	return query.GetGroupInvitations(userID)
}

func (s *GroupService) RequestToJoinGroup(groupID, userID int) error {
	return query.RequestToJoinGroup(groupID, userID)
}

func (s *GroupService) HandleGroupMembershipRequest(groupID, targetUserID, actorID int, action, requestType string) error {
	return query.HandleGroupMembershipRequest(groupID, targetUserID, action, requestType)
}

func (s *GroupService) CreateGroupPost(groupID, userID int, content, image string) (int, error) {
	return query.CreateGroupPost(groupID, userID, content, image)
}

func (s *GroupService) GetGroupPosts(groupID, userID int) (interface{}, error) {
	return query.GetGroupPosts(groupID)
}

func (s *GroupService) GetGroupPostComments(postID, userID int) (interface{}, error) {
	return query.GetGroupPostComments(postID)
}

func (s *GroupService) CreateGroupComment(postID, userID int, content string) (int, error) {
	return query.CreateGroupComment(postID, userID, content)
}

func (s *GroupService) CreateGroupEvent(groupID, userID int, title, description string, eventTime time.Time) (int, error) {
	return query.CreateGroupEvent(groupID, title, description, eventTime)
}

func (s *GroupService) GetGroupEvents(groupID, userID int) (interface{}, error) {
	return query.GetGroupEvents(groupID, userID)
}

func (s *GroupService) GetGroupEvent(eventID, userID int) (interface{}, interface{}, error) {
	event, err := query.GetGroupEvent(eventID, userID)
	if err != nil {
		return nil, nil, err
	}

	responses, err := query.GetEventResponses(eventID)
	if err != nil {
		responses = []models.EventResponse{}
	}

	return event, responses, nil
}

func (s *GroupService) RespondToEvent(eventID, userID, optionID int) error {
	return query.RespondToEvent(eventID, userID, optionID)
}
