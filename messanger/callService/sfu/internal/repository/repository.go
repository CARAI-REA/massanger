package repository

import (
	"context"
	"time"
)

type AffinityRepository interface {
	Claim(ctx context.Context, roomUUID, instanceID, publicURL string, ttl time.Duration) (claimed bool, ownerURL string, err error)
	Refresh(ctx context.Context, roomUUID, instanceID string, ttl time.Duration) error
	Release(ctx context.Context, roomUUID, instanceID string) error
}
