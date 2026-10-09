package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	auth "socialNetwork/pkg/auth"
	db "socialNetwork/pkg/db"
	routes "socialNetwork/pkg/routes"
	services "socialNetwork/pkg/services"
)

func main() {
	// Initialize database
	err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database initialized successfully!")
	if err := auth.BootstrapModerator(); err != nil {
		log.Fatalf("Moderator setup failed: %v", err)
	}
	if err := auth.InitSessionStore(); err != nil {
		log.Fatalf("Failed to load sessions: %v", err)
	}

	routes.SetGroupService(services.NewGroupService())

	// Defer database closure
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Serve uploaded media from the local runtime directory.
	http.HandleFunc("/uploads/", auth.CorsMiddleware(auth.AuthMiddleware(routes.ServeMedia)))

	http.HandleFunc("/auth/google/config", auth.CorsMiddleware(auth.GoogleConfig))
	http.HandleFunc("/auth/google/start", auth.CorsMiddleware(auth.GoogleStart))
	http.HandleFunc("/auth/google/callback", auth.GoogleCallback)
	http.HandleFunc("/auth/google/registration", auth.CorsMiddleware(auth.GoogleRegistration))
	http.HandleFunc("/forgot-password", auth.CorsMiddleware(auth.ForgotPassword))
	http.HandleFunc("/reset-password", auth.CorsMiddleware(auth.ResetPassword))
	http.HandleFunc("/account", auth.CorsMiddleware(auth.AuthMiddleware(routes.AccountSettings)))
	http.HandleFunc("/blocks", auth.CorsMiddleware(auth.AuthMiddleware(routes.Blocks)))
	http.HandleFunc("/reports", auth.CorsMiddleware(auth.AuthMiddleware(routes.Reports)))
	http.HandleFunc("/moderation", auth.CorsMiddleware(auth.AuthMiddleware(routes.Moderation)))
	http.HandleFunc("/search", auth.CorsMiddleware(auth.AuthMiddleware(routes.Search)))
	http.HandleFunc("/content/manage", auth.CorsMiddleware(auth.AuthMiddleware(routes.ManageContent)))
	http.HandleFunc("/posts/like", auth.CorsMiddleware(auth.AuthMiddleware(routes.PostLike)))
	http.HandleFunc("/groups/manage", auth.CorsMiddleware(auth.AuthMiddleware(routes.ManageGroup)))
	// Routes
	http.HandleFunc("/", routes.ServeMain)
	http.HandleFunc("/register", auth.CorsMiddleware(auth.Register))
	http.HandleFunc("/login", auth.CorsMiddleware(auth.Login))
	http.HandleFunc("/logout", auth.CorsMiddleware(auth.Logout))
	http.HandleFunc("/profile", auth.CorsMiddleware(auth.AuthMiddleware(routes.Profile)))
	http.HandleFunc("/posts", auth.CorsMiddleware(auth.AuthMiddleware(routes.HandlePosts)))
	http.HandleFunc("/comments", auth.CorsMiddleware(auth.AuthMiddleware(routes.HandleComments)))
	http.HandleFunc("/create-post", auth.CorsMiddleware(auth.AuthMiddleware(routes.HandlePosts)))
	http.HandleFunc("/ws", auth.CorsMiddleware(auth.AuthMiddleware(routes.HandleWebSocket)))
	http.HandleFunc("/messages", auth.CorsMiddleware(auth.AuthMiddleware(routes.SendMessage)))
	http.HandleFunc("/messages/", auth.CorsMiddleware(auth.AuthMiddleware(routes.ServeMessages)))
	http.HandleFunc("/chat/contacts", auth.CorsMiddleware(auth.AuthMiddleware(routes.ServeChatContacts)))
	http.HandleFunc("/chat/read", auth.CorsMiddleware(auth.AuthMiddleware(routes.MarkMessagesRead)))
	http.HandleFunc("/chat/followed", auth.CorsMiddleware(auth.AuthMiddleware(routes.ServeFollowedUsers)))
	http.HandleFunc("/follow", auth.CorsMiddleware(auth.AuthMiddleware(routes.FollowUser)))
	http.HandleFunc("/unfollow", auth.CorsMiddleware(auth.AuthMiddleware(routes.UnfollowUser)))
	http.HandleFunc("/follow-requests", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetFollowRequests)))
	http.HandleFunc("/follow-request", auth.CorsMiddleware(auth.AuthMiddleware(routes.HandleFollowRequest)))
	http.HandleFunc("/users", auth.CorsMiddleware(auth.ProtectUsersEndpoint(routes.GetAllUsers)))
	http.HandleFunc("/groups/create", auth.CorsMiddleware(auth.AuthMiddleware(routes.CreateGroup)))
	http.HandleFunc("/groups/user", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetUserGroups)))
	http.HandleFunc("/groups/messages", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupMessages)))
	http.HandleFunc("/groups/join/request", auth.CorsMiddleware(auth.AuthMiddleware(routes.RequestToJoinGroup)))
	http.HandleFunc("/groups/details", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupDetails)))
	http.HandleFunc("/groups/invite", auth.CorsMiddleware(auth.AuthMiddleware(routes.InviteToGroup)))
	http.HandleFunc("/groups/invitations", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupInvitations)))
	http.HandleFunc("/groups/join/requests", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetJoinRequestsForCreator)))
	http.HandleFunc("/groups/membership/handle", auth.CorsMiddleware(auth.AuthMiddleware(routes.HandleGroupMembershipRequest)))
	http.HandleFunc("/groups/posts/create", auth.CorsMiddleware(auth.AuthMiddleware(routes.CreateGroupPost)))
	http.HandleFunc("/groups/posts", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupPosts)))
	http.HandleFunc("/groups/posts/comments", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupPostComments)))
	http.HandleFunc("/groups/posts/comments/create", auth.CorsMiddleware(auth.AuthMiddleware(routes.CreateGroupComment)))
	http.HandleFunc("/groups/events/create", auth.CorsMiddleware(auth.AuthMiddleware(routes.CreateGroupEvent)))
	http.HandleFunc("/groups/events", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupEvents)))
	http.HandleFunc("/groups/events/details", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupEvent)))
	http.HandleFunc("/groups/events/respond", auth.CorsMiddleware(auth.AuthMiddleware(routes.RespondToEvent)))
	http.HandleFunc("/groups/all", auth.CorsMiddleware(auth.AuthMiddleware(services.GetAllGroups)))
	http.HandleFunc("/chat/upload-image", auth.CorsMiddleware(auth.AuthMiddleware(routes.UploadChatImage)))
	http.HandleFunc("/api/friends", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetFriends)))
	http.HandleFunc("/profile/", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetUserProfile)))
	http.HandleFunc("/user/", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetUserConnections)))
	// http.HandleFunc("/api/following", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetUserFollowing)))
	http.HandleFunc("/api/following", auth.CorsMiddleware(auth.AuthMiddleware(routes.ServeFollowingUsers)))

	http.HandleFunc("/profile/privacy", auth.CorsMiddleware(auth.AuthMiddleware(routes.UpdatePrivacy)))

	// Notification routes
	http.HandleFunc("/notifications", auth.CorsMiddleware(auth.AuthMiddleware(routes.ServeNotifications)))
	http.HandleFunc("/notifications/count", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetNotificationCount)))
	http.HandleFunc("/notifications/read", auth.CorsMiddleware(auth.AuthMiddleware(routes.MarkNotificationAsRead)))
	http.HandleFunc("/notifications/read-all", auth.CorsMiddleware(auth.AuthMiddleware(routes.MarkAllNotificationsAsRead)))
	http.HandleFunc("/notifications/action", auth.CorsMiddleware(auth.AuthMiddleware(routes.HandleNotificationAction)))

	http.HandleFunc("/user/info", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetUserInfo)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	address := ":" + port
	fmt.Printf("Server is running on http://0.0.0.0:%s\n", port)
	err = http.ListenAndServe(address, routes.MutationEvents(http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
