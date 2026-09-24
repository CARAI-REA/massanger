package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"signaling/internal/metrics"
	"signaling/internal/model"
	"signaling/internal/service"
)

type Handler struct {
	sessions        service.SessionService
	verifier        tokens.JoinTokenVerifier
	wsPath          string
	origins         map[string]struct{}
	pingEvery       time.Duration
	pongWait        time.Duration
	maxBytes        int64
	maxPerSec       int
	allowQueryToken bool
	upgrader        websocket.Upgrader
}

func NewHandler(
	sessions service.SessionService,
	verifier tokens.JoinTokenVerifier,
	wsPath string,
	origins []string,
	pingEvery, pongWait time.Duration,
	maxBytes int64,
	maxPerSec int,
	allowQueryToken bool,
) *Handler {
	originSet := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		originSet[o] = struct{}{}
	}
	h := &Handler{
		sessions:        sessions,
		verifier:        verifier,
		wsPath:          wsPath,
		origins:         originSet,
		pingEvery:       pingEvery,
		pongWait:        pongWait,
		maxBytes:        maxBytes,
		maxPerSec:       maxPerSec,
		allowQueryToken: allowQueryToken,
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	return h
}

func (h *Handler) checkOrigin(r *http.Request) bool {
	if len(h.origins) == 0 {
		return true
	}
	origin := r.Header.Get("Origin")
	_, ok := h.origins[origin]
	return ok
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(h.origins) > 0 && !h.checkOrigin(r) {
		metrics.HandshakeFailuresTotal.WithLabelValues("origin").Inc()
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}

	token := extractToken(r, h.allowQueryToken)
	if token == "" {
		metrics.HandshakeFailuresTotal.WithLabelValues("missing_token").Inc()
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	roomQuery := extractRoomUUID(r)
	if roomQuery == "" {
		metrics.HandshakeFailuresTotal.WithLabelValues("missing_room").Inc()
		http.Error(w, "missing room_uuid", http.StatusBadRequest)
		return
	}

	claims, err := h.verifier.VerifyJoinToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, tokens.ErrTokenExpired) {
			metrics.HandshakeFailuresTotal.WithLabelValues("token_expired").Inc()
			ws, upErr := h.upgrader.Upgrade(w, r, nil)
			if upErr != nil {
				http.Error(w, "token expired", http.StatusUnauthorized)
				return
			}
			_ = ws.WriteJSON(map[string]any{
				"type": "error",
				"payload": map[string]string{
					"code":    model.CodeTokenExpired,
					"message": "join token expired",
				},
			})
			_ = ws.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(model.CloseTokenExpired, model.CodeTokenExpired))
			_ = ws.Close()
			return
		}
		metrics.HandshakeFailuresTotal.WithLabelValues("invalid_token").Inc()
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if claims.RoomUUID != roomQuery {
		metrics.HandshakeFailuresTotal.WithLabelValues("room_mismatch").Inc()
		http.Error(w, "room mismatch", http.StatusUnauthorized)
		return
	}

	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		metrics.HandshakeFailuresTotal.WithLabelValues("upgrade").Inc()
		logger.Error(r.Context(), "ws upgrade failed", zap.Error(err))
		return
	}

	conn := NewClientConn(
		ws,
		claims.RoomUUID,
		claims.UserUUID,
		h.pingEvery,
		h.pongWait,
		h.maxBytes,
		h.maxPerSec,
		h.sessions,
	)

	ctx := logger.ContextWithUserID(r.Context(), claims.UserUUID)
	if err := h.sessions.Connect(ctx, conn, claims.RoomUUID, claims.UserUUID); err != nil {
		reason := "room_check"
		switch {
		case errors.Is(err, model.ErrUnauthenticated):
			reason = "unauthenticated"
		case errors.Is(err, model.ErrNotFound):
			reason = "not_found"
		case errors.Is(err, model.ErrRoomClosed):
			reason = "room_closed"
		case errors.Is(err, model.ErrPermissionDenied):
			reason = "permission_denied"
		}
		metrics.HandshakeFailuresTotal.WithLabelValues(reason).Inc()
		_ = conn.Close(websocket.ClosePolicyViolation, reason)
		return
	}

	conn.Run(ctx)
}
