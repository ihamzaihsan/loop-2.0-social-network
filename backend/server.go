package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	auth "socialNetwork/pkg/auth"
	"socialNetwork/pkg/cloud"
	db "socialNetwork/pkg/db"
	routes "socialNetwork/pkg/routes"
	services "socialNetwork/pkg/services"
	ws "socialNetwork/pkg/websocket"
	"strings"
	"time"
)

func main() {
	if err := cloud.Validate(); err != nil {
		log.Fatal(err)
	}
	if os.Getenv("VERCEL") != "" && !strings.HasPrefix(os.Getenv("FRONTEND_URL"), "https://") {
		log.Fatal("FRONTEND_URL must be the HTTPS deployment origin")
	}
	// Initialize database
	err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database initialized successfully!")
	if err := auth.BootstrapModerator(); err != nil {
		log.Fatalf("Moderator setup failed: %v", err)
	}
	if err := auth.InitSessions(); err != nil {
		log.Fatalf("Failed to load sessions: %v", err)
	}

	routes.SetGroupService(services.NewGroupService())
	relayContext, cancelRelay := context.WithCancel(context.Background())
	defer cancelRelay()
	ws.StartRelay(relayContext)

	// Defer database closure
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Serve authorized media from local storage or a private Supabase bucket.
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
	handler := auth.SecurityMiddleware(routes.MutationEvents(http.DefaultServeMux))
	// Vercel Services forwards the original /api path. Nginx strips it locally.
	if os.Getenv("VERCEL") != "" {
		prefix := strings.TrimRight(os.Getenv("PUBLIC_API_PATH"), "/")
		if prefix != "/api" {
			log.Fatal("PUBLIC_API_PATH must be /api on Vercel")
		}
		handler = http.StripPrefix(prefix, handler)
	}
	server := &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
