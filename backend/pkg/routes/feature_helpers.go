package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"socialNetwork/pkg/auth"
	"strconv"
	"strings"
	"unicode/utf8"
)

func respond(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}
func decode(w http.ResponseWriter, r *http.Request, value interface{}) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if decoder.Decode(value) != nil || decoder.Decode(new(interface{})) != io.EOF {
		http.Error(w, "Invalid request body", 400)
		return false
	}
	return true
}
func requestID(r *http.Request) int { id, _ := strconv.Atoi(r.URL.Query().Get("id")); return id }
func viewer(r *http.Request) int    { id, _ := auth.GetUserID(r); return id }
func validText(value string, max int, required bool) bool {
	return utf8.RuneCountInString(value) <= max && (!required || strings.TrimSpace(value) != "")
}
