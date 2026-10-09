package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSessionFromCookie(r)
		if err != nil || session == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Unauthorized - Please login",
			})
			return
		}

		if !session.IsActive {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Session expired - Please login again",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}

func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow multiple frontend origins for both Docker (3000) and direct execution (3001)
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
		}

		if configured := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/"); configured != "" {
			allowedOrigins = append(allowedOrigins, configured)
		}
		w.Header().Add("Vary", "Origin")
		allowed := origin == ""
		// Check if the origin is in our allowed list
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		if !allowed {
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}
		// If no origin header (like direct API calls), allow localhost:3000 as default
		if origin == "" {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func ProtectUsersEndpoint(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		session, err := GetSessionFromCookie(r)
		if err != nil || session == nil || !session.IsActive {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Unauthorized - Please login",
			})
			return
		}

		accept := r.Header.Get("Accept")
		isAPIRequest := strings.Contains(accept, "application/json")

		if !isAPIRequest {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Page not found",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}
