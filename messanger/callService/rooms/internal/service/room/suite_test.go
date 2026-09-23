package room

import (
	"context"
	"testing"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/stretchr/testify/suite"

	"rooms/internal/kafka"
	"rooms/internal/repository/mocks"
	"rooms/internal/repository/roomcache"
)

type ServiceSuite struct {
	suite.Suite

	ctx context.Context

	roomRepository *mocks.RoomRepository

	service *service
}

func (s *ServiceSuite) SetupTest() {
	logger.SetNopLogger()

	s.roomRepository = mocks.NewRoomRepository(s.T())

	s.service = NewService(s.roomRepository, kafka.NoopProducer{}, roomcache.NewNoop(), time.Hour)
}

func (s *ServiceSuite) TearDownTest() {
}

func TestServiceIntegration(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
