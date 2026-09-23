package roomcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/cache"

	"rooms/internal/model"
	"rooms/internal/repository"
)

const (
	keyRoomInfoFmt        = "room:%s:info"
	keyRoomParticipantsFmt = "room:%s:participants"
)

// Redis реализует repository.RoomCache (clean_arch: api → service → cache repo → redis client).
type Redis struct {
	client cache.RedisClient
}

func NewRedis(client cache.RedisClient) *Redis {
	return &Redis{client: client}
}

var _ repository.RoomCache = (*Redis)(nil)

func roomInfoKey(roomUUID string) string {
	return fmt.Sprintf(keyRoomInfoFmt, roomUUID)
}

func participantsKey(roomUUID string) string {
	return fmt.Sprintf(keyRoomParticipantsFmt, roomUUID)
}

func (r *Redis) SetRoom(ctx context.Context, room model.Room, ttl time.Duration) error {
	b, err := json.Marshal(room)
	if err != nil {
		return err
	}
	return r.client.SetWithTTL(ctx, roomInfoKey(room.RoomUUID), b, ttl)
}

func (r *Redis) GetRoom(ctx context.Context, roomUUID string) (model.Room, error) {
	b, err := r.client.Get(ctx, roomInfoKey(roomUUID))
	if err != nil {
		if errors.Is(err, cache.ErrCacheMiss) {
			return model.Room{}, cache.ErrCacheMiss
		}
		return model.Room{}, err
	}
	var room model.Room
	if err := json.Unmarshal(b, &room); err != nil {
		return model.Room{}, err
	}
	return room, nil
}

func (r *Redis) InvalidateRoomInfo(ctx context.Context, roomUUID string) error {
	return r.client.Del(ctx, roomInfoKey(roomUUID))
}

func (r *Redis) PurgeRoom(ctx context.Context, roomUUID string) error {
	return r.client.Del(ctx, roomInfoKey(roomUUID), participantsKey(roomUUID))
}

func (r *Redis) AddActiveParticipant(ctx context.Context, roomUUID, userUUID string) error {
	return r.client.SAdd(ctx, participantsKey(roomUUID), userUUID)
}

func (r *Redis) RemoveActiveParticipant(ctx context.Context, roomUUID, userUUID string) error {
	return r.client.SRem(ctx, participantsKey(roomUUID), userUUID)
}

func (r *Redis) ActiveParticipantCount(ctx context.Context, roomUUID string) (int64, error) {
	return r.client.SCard(ctx, participantsKey(roomUUID))
}
