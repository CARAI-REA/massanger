package model

import "errors"

var (
	ErrUnauthenticated     = errors.New("unauthenticated")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrNotFound            = errors.New("not found")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrRoomClosed          = errors.New("room closed")
	ErrSessionReplaced     = errors.New("session replaced")
	ErrRateLimited         = errors.New("rate limited")
	ErrRoomOnOtherInstance = errors.New("room on other instance")
	ErrTURNUnavailable     = errors.New("turn credentials unavailable")
	ErrInstanceDraining    = errors.New("instance draining")
	ErrInternal            = errors.New("internal")
)

const (
	CodeUnauthenticated     = "UNAUTHENTICATED"
	CodePermissionDenied    = "PERMISSION_DENIED"
	CodeNotFound            = "NOT_FOUND"
	CodeInvalidArgument     = "INVALID_ARGUMENT"
	CodeRoomClosed          = "ROOM_CLOSED"
	CodeSessionReplaced     = "SESSION_REPLACED"
	CodeRateLimited         = "RATE_LIMITED"
	CodeRoomOnOtherInstance = "ROOM_ON_OTHER_INSTANCE"
	CodeTokenExpired        = "TOKEN_EXPIRED"
	CodeInstanceDraining    = "INSTANCE_DRAINING"
	CodeTURNUnavailable     = "TURN_UNAVAILABLE"
	CodeInternal            = "INTERNAL"

	// CloseTokenExpired is the WS close code when join JWT is expired.
	CloseTokenExpired = 4001
	// CloseInstanceDraining is the WS close code when SFU rejects new sessions during drain.
	CloseInstanceDraining = 4002
)
