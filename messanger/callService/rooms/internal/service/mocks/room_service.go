package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"rooms/internal/model"
)

type RoomService struct {
	mock.Mock
}

func NewRoomService(t interface {
	mock.TestingT
	Cleanup(func())
}) *RoomService {
	m := &RoomService{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *RoomService) CreateRoom(ctx context.Context, req model.CreateRoomRequest) (model.CreateRoomResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.CreateRoomResponse), args.Error(1)
}

func (m *RoomService) GetRoom(ctx context.Context, req model.GetRoomRequest) (model.GetRoomResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.GetRoomResponse), args.Error(1)
}

func (m *RoomService) UpdateRoom(ctx context.Context, req model.UpdateRoomRequest) (model.UpdateRoomResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.UpdateRoomResponse), args.Error(1)
}

func (m *RoomService) DeleteRoom(ctx context.Context, req model.DeleteRoomRequest) (model.DeleteRoomResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.DeleteRoomResponse), args.Error(1)
}

func (m *RoomService) ListRooms(ctx context.Context, req model.ListRoomRequest) (model.ListRoomResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.ListRoomResponse), args.Error(1)
}

func (m *RoomService) EndRoom(ctx context.Context, req model.EndRoomRequest) (model.EndRoomResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.EndRoomResponse), args.Error(1)
}
