package models

type Profile struct {
	User           User   `json:"user"`
	Posts          []Post `json:"posts"`
	PostsCount     int    `json:"postsCount"`
	FollowersCount int    `json:"followersCount"`
	FollowingCount int    `json:"followingCount"`
	Followers      []User `json:"followers"`
	Following      []User `json:"following"`
}
