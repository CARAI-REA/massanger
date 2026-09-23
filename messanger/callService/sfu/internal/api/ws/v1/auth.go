package v1

import (
	"net/http"
	"strings"
)

func extractToken(r *http.Request, allowQueryToken bool) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	if !allowQueryToken {
		return ""
	}
	return strings.TrimSpace(r.URL.Query().Get("token"))
}

func extractRoomUUID(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("room_uuid"))
}
