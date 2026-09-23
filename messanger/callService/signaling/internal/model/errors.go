package model

import "errors"

var (
	ErrUnauthenticated  = errors.New("unauthenticated")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNotFound         = errors.New("not found")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrRoomClosed       = errors.New("room closed")
	ErrSessionReplaced  = errors.New("session replaced")
	ErrRateLimited      = errors.New("rate limited")
	ErrInternal         = errors.New("internal")
)

const (
	CodeUnauthenticated  = "UNAUTHENTICATED"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeNotFound         = "NOT_FOUND"
	CodeInvalidArgument  = "INVALID_ARGUMENT"
	CodeRoomClosed       = "ROOM_CLOSED"
	CodeSessionReplaced  = "SESSION_REPLACED"
	CodeRateLimited      = "RATE_LIMITED"
	CodeInternal         = "INTERNAL"
)
