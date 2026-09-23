package service

import (
	"context"

	"sfu/internal/model"
)

type Conn interface {
	Send(ctx context.Context, msg model.Envelope) error
	SendAndClose(ctx context.Context, msg model.Envelope, code int, reason string) error
	Close(code int, reason string) error
	RemoteAddr() string
}

type SessionService interface {
	Connect(ctx context.Context, conn Conn, roomUUID, userUUID string) error
	Handle(ctx context.Context, roomUUID, userUUID string, msg model.Envelope) error
	Disconnect(ctx context.Context, conn Conn, roomUUID, userUUID string, reason string) error
	CloseRoom(ctx context.Context, roomUUID string) error
	CloseUser(ctx context.Context, roomUUID, userUUID string) error
}
