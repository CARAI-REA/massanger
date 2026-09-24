package room

import (
	"context"

	"sfu/internal/model"
)

type Noop struct{}

func NewNoop() *Noop { return &Noop{} }

func (n *Noop) AssertCanJoin(context.Context, string, string) (JoinCheck, error) {
	return JoinCheck{OK: true}, nil
}

func (n *Noop) GetTURNCredentials(context.Context, string, string) ([]model.ICEServer, error) {
	return nil, nil
}

type Denied struct{}

func (d *Denied) AssertCanJoin(context.Context, string, string) (JoinCheck, error) {
	return JoinCheck{}, model.ErrPermissionDenied
}

func (d *Denied) GetTURNCredentials(context.Context, string, string) ([]model.ICEServer, error) {
	return nil, model.ErrPermissionDenied
}
