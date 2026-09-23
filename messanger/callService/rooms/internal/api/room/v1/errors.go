package v1

import (
	"errors"

	"rooms/internal/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, model.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, model.ErrRoomNotFound), errors.Is(err, model.ErrParticipantNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, model.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, model.ErrResourceExhausted):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, model.ErrParticipantConflict), errors.Is(err, model.ErrRoomConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, model.ErrRoomInactive), errors.Is(err, model.ErrRoomNoContent):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return err
	}
}
