package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"rooms/internal/service/mocks"
)

type APISuite struct {
	suite.Suite

	ctx context.Context

	roomService        *mocks.RoomService
	participantService *mocks.ParticipantService

	api *api
}

func (s *APISuite) SetupTest() {
	s.ctx = context.Background()

	s.roomService = mocks.NewRoomService(s.T())
	s.participantService = mocks.NewParticipantService(s.T())

	s.api = NewAPI(s.roomService, s.participantService)
}

func (s *APISuite) TearDownTest() {
}

func TestAPIIntegration(t *testing.T) {
	suite.Run(t, new(APISuite))
}
