package main

import (
	"fmt"
	"log"
	"net/http"
	auth "socialNetwork/pkg/auth"
	db "socialNetwork/pkg/db"
	routes "socialNetwork/pkg/routes"
)

func main() {
	// Initialize database
	err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database initialized successfully!")

	// Defer database closure
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	//routes
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
	http.HandleFunc("/chat/followed", auth.CorsMiddleware(auth.AuthMiddleware(routes.ServeFollowedUsers)))
	http.HandleFunc("/follow", auth.CorsMiddleware(auth.AuthMiddleware(routes.FollowUser)))
	http.HandleFunc("/unfollow", auth.CorsMiddleware(auth.AuthMiddleware(routes.UnfollowUser)))
	http.HandleFunc("/users", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetAllUsers)))
	http.HandleFunc("/groups/create", auth.CorsMiddleware(auth.AuthMiddleware(routes.CreateGroup)))
	http.HandleFunc("/groups/user", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetUserGroups)))
	http.HandleFunc("/groups/details", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupDetails)))
	http.HandleFunc("/groups/invite", auth.CorsMiddleware(auth.AuthMiddleware(routes.InviteToGroup)))
	http.HandleFunc("/groups/invitations", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupInvitations)))
	http.HandleFunc("/groups/join/request", auth.CorsMiddleware(auth.AuthMiddleware(routes.RequestToJoinGroup)))
	http.HandleFunc("/groups/membership/handle", auth.CorsMiddleware(auth.AuthMiddleware(routes.HandleGroupMembershipRequest)))
	http.HandleFunc("/groups/posts/create", auth.CorsMiddleware(auth.AuthMiddleware(routes.CreateGroupPost)))
	http.HandleFunc("/groups/posts", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupPosts)))
	http.HandleFunc("/groups/posts/comments", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupPostComments)))
	http.HandleFunc("/groups/posts/comments/create", auth.CorsMiddleware(auth.AuthMiddleware(routes.CreateGroupComment)))
	http.HandleFunc("/groups/events/create", auth.CorsMiddleware(auth.AuthMiddleware(routes.CreateGroupEvent)))
	http.HandleFunc("/groups/events", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupEvents)))
	http.HandleFunc("/groups/events/details", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetGroupEvent)))
	http.HandleFunc("/groups/events/respond", auth.CorsMiddleware(auth.AuthMiddleware(routes.RespondToEvent)))

	fmt.Println("Server is running on http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
