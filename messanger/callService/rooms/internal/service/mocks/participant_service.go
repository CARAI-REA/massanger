package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"rooms/internal/model"
)

type ParticipantService struct {
	mock.Mock
}

func NewParticipantService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ParticipantService {
	m := &ParticipantService{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *ParticipantService) AddParticipant(ctx context.Context, req model.AddParticipantRequest) (model.AddParticipantResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.AddParticipantResponse), args.Error(1)
}

func (m *ParticipantService) RemoveParticipant(ctx context.Context, req model.RemoveParticipantRequest) (model.RemoveParticipantResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.RemoveParticipantResponse), args.Error(1)
}

func (m *ParticipantService) GetParticipant(ctx context.Context, req model.GetParticipantsRequest) (model.GetParticipantsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.GetParticipantsResponse), args.Error(1)
}

func (m *ParticipantService) IsParticipant(ctx context.Context, req model.IsParticipantRequest) (model.IsParticipantResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.IsParticipantResponse), args.Error(1)
}

func (m *ParticipantService) UpdateParticipantMetadata(
	ctx context.Context,
	req model.UpdateParticipantMetadataRequest,
) (model.UpdateParticipantMetadataResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.UpdateParticipantMetadataResponse), args.Error(1)
}

func (m *ParticipantService) AssertCanJoin(ctx context.Context, req model.AssertCanJoinRequest) (model.AssertCanJoinResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.AssertCanJoinResponse), args.Error(1)
}

func (m *ParticipantService) RefreshJoinToken(ctx context.Context, req model.RefreshJoinTokenRequest) (model.RefreshJoinTokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.RefreshJoinTokenResponse), args.Error(1)
}

func (m *ParticipantService) GetTURNCredentials(ctx context.Context, req model.GetTURNCredentialsRequest) (model.GetTURNCredentialsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(model.GetTURNCredentialsResponse), args.Error(1)
}
