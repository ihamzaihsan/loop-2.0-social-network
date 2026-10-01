package models

import "time"

type Group struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatorID   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	MemberCount int       `json:"member_count,omitempty"`
	IsCreator   bool      `json:"is_creator,omitempty"`
	Role        string    `json:"role,omitempty"`
	Status      string    `json:"status,omitempty"`
}

type GroupMember struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"group_id"`
	UserID    int       `json:"user_id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Avatar    *string   `json:"avatar,omitempty"`
}

type GroupInvitation struct {
	ID          int       `json:"id"`
	GroupID     int       `json:"group_id"`
	InviterID   int       `json:"inviter_id"`
	InviteeID   int       `json:"invitee_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	GroupTitle  string    `json:"group_title,omitempty"`
	InviterName string    `json:"inviter_name,omitempty"`
}

type GroupJoinRequest struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"group_id"`
	Title     string    `json:"title"`
	UserID    int       `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Avatar    *string   `json:"avatar,omitempty"`
}

type GroupPost struct {
	ID           int       `json:"id"`
	GroupID      int       `json:"group_id"`
	UserID       int       `json:"user_id"`
	Content      string    `json:"content"`
	Image        string    `json:"image,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	FirstName    string    `json:"first_name,omitempty"`
	LastName     string    `json:"last_name,omitempty"`
	Avatar       *string   `json:"avatar,omitempty"`
	CommentCount int       `json:"comment_count,omitempty"`
}

type GroupComment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Avatar    *string   `json:"avatar,omitempty"`
}

type GroupEvent struct {
	ID              int                   `json:"id"`
	GroupID         int                   `json:"group_id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	EventTime       time.Time             `json:"event_time"`
	CreatedAt       time.Time             `json:"created_at"`
	UserResponse    string                `json:"user_response,omitempty"`
	GoingCount      int                   `json:"going_count"`
	NotGoingCount   int                   `json:"not_going_count"`
	ResponseOptions []EventResponseOption `json:"response_options,omitempty"`
}
type UserEventResponse struct {
	ID         int    `json:"id"`
	OptionID   int    `json:"option_id"`
	OptionText string `json:"option_text"`
}
type EventResponseOption struct {
	ID            int    `json:"id"`
	EventID       int    `json:"event_id"`
	OptionText    string `json:"option_text"`
	ResponseCount int    `json:"response_count,omitempty"`
}

type EventResponse struct {
	ID               int       `json:"id"`
	EventID          int       `json:"event_id"`
	UserID           int       `json:"user_id"`
	ResponseOptionID int       `json:"response_option_id"`
	CreatedAt        time.Time `json:"created_at"`
	FirstName        string    `json:"first_name,omitempty"`
	LastName         string    `json:"last_name,omitempty"`
	Avatar           *string   `json:"avatar,omitempty"`
	OptionText       string    `json:"option_text,omitempty"`
}

type GroupMessage struct {
	ID        int       `json:"id"`
	SenderID  int       `json:"sender_id"`
	GroupID   int       `json:"group_id,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Sender    struct {
		ID        int    `json:"id"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Avatar    string `json:"avatar,omitempty"`
	} `json:"sender"`
}

type GroupService interface {
	CreateGroup(title, description string, userID int) (int, error)
	GetUserGroups(userID int) (interface{}, error)
	GetGroupDetails(groupID, userID int) (interface{}, error)
	InviteToGroup(groupID, inviterID int, userIDs []int) error
	GetGroupInvitations(userID int) (interface{}, error)
	RequestToJoinGroup(groupID, userID int) error
	HandleGroupMembershipRequest(groupID, targetUserID, actorID int, action, requestType string) error
	CreateGroupPost(groupID, userID int, content, image string) (int, error)
	GetGroupPosts(groupID, userID int) (interface{}, error)
	GetGroupPostComments(postID, userID int) (interface{}, error)
	CreateGroupComment(postID, userID int, content string) (int, error)
	CreateGroupEvent(groupID, userID int, title, description string, eventTime time.Time) (int, error)
	GetGroupEvents(groupID, userID int) (interface{}, error)
	GetGroupEvent(eventID, userID int) (interface{}, interface{}, error)
	RespondToEvent(eventID, userID, optionID int) error
	IsGroupMember(groupID, userID int) (bool, error)
	GetGroupMessages(groupID int) ([]GroupMessage, error)
}
