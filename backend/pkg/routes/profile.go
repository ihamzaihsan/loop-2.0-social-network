package routes

import (
	"encoding/json"
	"log"
	"net/http"
	auth "socialNetwork/pkg/auth"
	query "socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
	"strings"
)

func Profile(w http.ResponseWriter, r *http.Request) {

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	UserID, err := auth.GetUserID(r)
	if err != nil {
		log.Printf("Auth error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("Fetching profile for user ID: %d", UserID)

	user, err := query.GetUserInfo(UserID)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	posts, err := query.GetPostsByUserID(UserID)
	if err != nil {
		log.Printf("Error getting user posts: %v", err)
		posts = []models.PostResponse{} // Default to empty array if there's an error
	}

	// Fix image paths in posts
	for i := range posts {
		if posts[i].Image != "" {
			// If the image path starts with "./uploads", convert it to "/uploads"
			if strings.HasPrefix(posts[i].Image, "./uploads") {
				posts[i].Image = strings.Replace(posts[i].Image, "./uploads", "/uploads", 1)
			}
		}

		// Ensure author data is properly set
		if posts[i].Author.Avatar != nil && *posts[i].Author.Avatar != "" {
			avatarStr := *posts[i].Author.Avatar
			if strings.HasPrefix(avatarStr, "./uploads") {
				avatarStr = strings.Replace(avatarStr, "./uploads", "/uploads", 1)
				posts[i].Author.Avatar = &avatarStr
			}
		}
	}

	postsCount, err := query.GetUserPostsCount(UserID)
	if err != nil {
		log.Printf("Error getting posts count: %v", err)
		postsCount = 0 // Default to 0 if there's an error
	}

	var followers []models.User = []models.User{}
	var followersCount int = 0
	var following []models.User = []models.User{}
	var followingCount int = 0

	followersResult, followersCountResult, err := query.GetFollowers(uint(UserID))
	if err == nil {
		for _, f := range followersResult {
			var avatar *string
			if f.Avatar != "" {
				avatarStr := f.Avatar
				avatar = &avatarStr
			}

			user := models.User{
				ID:        int(f.FollowerID),
				FirstName: "",
				LastName:  "",
				Avatar:    avatar,
			}
			followers = append(followers, user)
		}
		followersCount = followersCountResult
	} else {
		log.Printf("Warning: Could not get followers: %v", err)
	}

	followingResult, followingCountResult, err := query.GetFollowing(uint(UserID))
	if err == nil {
		for _, f := range followingResult {
			var avatar *string
			if f.Avatar != "" {
				avatarStr := f.Avatar
				avatar = &avatarStr
			}

			user := models.User{
				ID:        int(f.FollowedID),
				FirstName: "",
				LastName:  "",
				Avatar:    avatar,
			}
			following = append(following, user)
		}
		followingCount = followingCountResult
	} else {
		log.Printf("Warning: Could not get following: %v", err)
	}

	profile := map[string]interface{}{
		"user":           user,
		"posts":          posts,
		"postsCount":     postsCount,
		"followersCount": followersCount,
		"followingCount": followingCount,
		"followers":      followers,
		"following":      following,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Printf("Error encoding profile to JSON: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
