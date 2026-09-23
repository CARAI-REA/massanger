package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"rooms/internal/model"
)

type ParticipantRepository struct {
	mock.Mock
}

func (m *ParticipantRepository) AddParticipant(ctx context.Context, req model.AddParticipantRequest) (model.AddParticipantResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.AddParticipantResponse), args.Error(1)
}

func (m *ParticipantRepository) RemoveParticipant(ctx context.Context, req model.RemoveParticipantRequest) (model.RemoveParticipantResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.RemoveParticipantResponse), args.Error(1)
}

func (m *ParticipantRepository) GetParticipant(ctx context.Context, req model.GetParticipantsRequest) (model.GetParticipantsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.GetParticipantsResponse), args.Error(1)
}

func (m *ParticipantRepository) IsParticipant(ctx context.Context, req model.IsParticipantRequest) (model.IsParticipantResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.IsParticipantResponse), args.Error(1)
}

func (m *ParticipantRepository) UpdateParticipantMetadata(ctx context.Context, req model.UpdateParticipantMetadataRequest) (model.UpdateParticipantMetadataResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.UpdateParticipantMetadataResponse), args.Error(1)
}

func NewParticipantRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *ParticipantRepository {
	m := &ParticipantRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}
