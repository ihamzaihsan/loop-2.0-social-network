package routes

import (
	"encoding/json"
	"log"
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