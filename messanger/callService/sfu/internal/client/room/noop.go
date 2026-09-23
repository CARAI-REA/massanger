package room

import (
	"context"

	"sfu/internal/model"
)

type Noop struct{}

func NewNoop() *Noop { return &Noop{} }

func (n *Noop) AssertCanJoin(context.Context, string, string) error { return nil }

type Denied struct{}

func (d *Denied) AssertCanJoin(context.Context, string, string) error {
	return model.ErrPermissionDenied
}
