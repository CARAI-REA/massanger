package authctx

import (
	"context"

	"rooms/internal/model"
)

type userIDKey struct{}
type serviceNameKey struct{}

// WithUserID кладёт идентификатор пользователя из JWT в контекст (см. room_task.md §7.1).
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserID возвращает user_id из контекста, если он задан и не пустой.
func UserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey{}).(string)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}

// MustUserID как UserID, но для ошибки использует model.ErrUnauthenticated.
func MustUserID(ctx context.Context) (string, error) {
	id, ok := UserID(ctx)
	if !ok {
		return "", model.ErrUnauthenticated
	}
	return id, nil
}

// WithServiceName stores the calling service name from a service JWT.
func WithServiceName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, serviceNameKey{}, name)
}

// ServiceName returns the service claim from context when present.
func ServiceName(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(serviceNameKey{}).(string)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}
