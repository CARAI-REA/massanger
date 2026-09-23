package participant

import (
	"rooms/internal/config"
	"rooms/internal/kafka"
	"rooms/internal/repository"
	def "rooms/internal/service"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
)

var _ def.ParticipantService = (*service)(nil)

type service struct {
	roomRepository        repository.RoomRepository
	participantRepository repository.ParticipantRepository
	eventProducer         kafka.EventProducer
	roomCache             repository.RoomCache
	joinTokenService      tokens.JoinTokenService
	turnConfig            config.TURNConfig
}

func NewService(
	roomRepository repository.RoomRepository,
	participantRepository repository.ParticipantRepository,
	eventProducer kafka.EventProducer,
	roomCache repository.RoomCache,
	joinTokenService tokens.JoinTokenService,
	turnConfig config.TURNConfig,
) *service {
	return &service{
		roomRepository:        roomRepository,
		participantRepository: participantRepository,
		eventProducer:         eventProducer,
		roomCache:             roomCache,
		joinTokenService:      joinTokenService,
		turnConfig:            turnConfig,
	}
}

type joinInfo struct {
	userUUID string
	roomUUID string
}

func (j joinInfo) GetUserUUID() string { return j.userUUID }
func (j joinInfo) GetRoomUUID() string { return j.roomUUID }
