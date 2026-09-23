package participant

import (
	"context"
	"fmt"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/turncred"

	"rooms/internal/model"
)

func (s *service) GetTURNCredentials(ctx context.Context, req model.GetTURNCredentialsRequest) (model.GetTURNCredentialsResponse, error) {
	if req.RoomUUID == "" || req.UserUUID == "" {
		return model.GetTURNCredentialsResponse{}, model.ErrInvalidArgument
	}

	partRes, err := s.participantRepository.IsParticipant(ctx, model.IsParticipantRequest{
		RoomUUID: req.RoomUUID,
		UserUUID: req.UserUUID,
	})
	if err != nil {
		return model.GetTURNCredentialsResponse{}, err
	}
	if !partRes.IsActive {
		return model.GetTURNCredentialsResponse{}, model.ErrPermissionDenied
	}

	if s.turnConfig == nil {
		return model.GetTURNCredentialsResponse{}, fmt.Errorf("turn config is not configured")
	}

	cred, err := turncred.Generate(s.turnConfig.SharedSecret(), req.UserUUID, s.turnConfig.TTL())
	if err != nil {
		return model.GetTURNCredentialsResponse{}, err
	}

	urls := s.turnConfig.URLs()
	servers := make([]model.ICEServer, 0, 1)
	if len(urls) > 0 {
		servers = append(servers, model.ICEServer{
			URLs:       urls,
			Username:   cred.Username,
			Credential: cred.Password,
		})
	}

	return model.GetTURNCredentialsResponse{
		ICEServers: servers,
		TTLSeconds: int64(cred.TTL.Seconds()),
	}, nil
}
