package room

import (
	"time"

	"rooms/internal/kafka"
	"rooms/internal/repository"
	def "rooms/internal/service"
)

var _ def.RoomService = (*service)(nil)

type service struct {
	roomRepository repository.RoomRepository
	eventProducer  kafka.EventProducer
	roomCache      repository.RoomCache
	cacheTTL       time.Duration
}

func NewService(
	roomRepository repository.RoomRepository,
	eventProducer kafka.EventProducer,
	roomCache repository.RoomCache,
	cacheTTL time.Duration,
) *service {
	return &service{
		roomRepository: roomRepository,
		eventProducer:  eventProducer,
		roomCache:      roomCache,
		cacheTTL:       cacheTTL,
	}
}
