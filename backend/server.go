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

	// Follow routes
	http.HandleFunc("/follow", auth.CorsMiddleware(auth.AuthMiddleware(routes.FollowUser)))
	http.HandleFunc("/unfollow", auth.CorsMiddleware(auth.AuthMiddleware(routes.UnfollowUser)))
	http.HandleFunc("/users", auth.CorsMiddleware(auth.AuthMiddleware(routes.GetAllUsers)))
	
	
	


	fmt.Println("Server is running on http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
