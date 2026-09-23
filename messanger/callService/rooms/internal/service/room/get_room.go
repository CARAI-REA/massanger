package room

import (
	"context"
	"errors"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/cache"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"go.uber.org/zap"

	"rooms/internal/model"
)

func (s *service) GetRoom(ctx context.Context, req model.GetRoomRequest) (model.GetRoomResponse, error) {
	if cached, err := s.roomCache.GetRoom(ctx, req.RoomUUID); err == nil {
		return model.GetRoomResponse{Room: cached}, nil
	} else if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return model.GetRoomResponse{}, err
	}

	room, err := s.roomRepository.GetRoom(ctx, req)
	if err != nil {
		logger.Error(ctx, "Ошибка при получении", zap.Error(err))
		return model.GetRoomResponse{}, err
	}

	_ = s.roomCache.SetRoom(ctx, room.Room, s.cacheTTL)

	return room, nil
}
