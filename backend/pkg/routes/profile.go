package routes

import (
	"encoding/json"
	"net/http"
	auth "socialNetwork/pkg/auth"
	query "socialNetwork/pkg/db/query"
	"socialNetwork/pkg/models"
)

func Profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := query.GetUserInfo(UserID)
	if err != nil {
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	posts, postsCount, err := query.GetUserPosts(UserID)
	if err != nil {
		http.Error(w, "Failed to get user posts", http.StatusInternalServerError)
		return
	}

	followers, followersCount, err := query.GetFollowers(UserID)
	if err != nil {
		http.Error(w, "Failed to get user followers", http.StatusInternalServerError)
		return
	}

	following, followingCount, err := query.GetFollowing(UserID)
	if err != nil {
		http.Error(w, "Failed to get user following", http.StatusInternalServerError)
		return
	}

	profile := models.Profile{
		User:           user,
		Posts:          posts,
		PostsCount:     postsCount,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
		Followers:      followers,
		Following:      following,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
