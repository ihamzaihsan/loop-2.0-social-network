package routes

import "net/http"

func ServeMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.Redirect(w, r, "http://localhost:3000", http.StatusSeeOther)
}
