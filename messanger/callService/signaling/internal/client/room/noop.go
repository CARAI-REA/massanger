package room

import (
	"context"

	"signaling/internal/model"
)

type Noop struct{}

func NewNoop() *Noop { return &Noop{} }

func (n *Noop) AssertCanJoin(context.Context, string, string) error {
	return nil
}

// Denied always denies — useful in tests.
type Denied struct{}

func (d *Denied) AssertCanJoin(context.Context, string, string) error {
	return model.ErrPermissionDenied
}
