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
	http.HandleFunc("/register", auth.Register)
	http.HandleFunc("/login", auth.Login)
	http.HandleFunc("/logout", auth.Logout)
	http.HandleFunc("/profile", auth.AuthMiddleware(routes.Profile))

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
