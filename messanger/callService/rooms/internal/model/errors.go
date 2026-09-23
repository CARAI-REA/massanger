package model

import "errors"

var ErrRoomNotFound = errors.New("room not found")

var ErrRoomConflict = errors.New("conflict")
var ErrRoomNoContent = errors.New("No content")

var ErrRoomInactive = errors.New("room inactive")

var ErrParticipantNotFound = errors.New("participant not found")
var ErrParticipantConflict = errors.New("participant conflict")

var ErrPermissionDenied = errors.New("permission denied")

var ErrUnauthenticated = errors.New("unauthenticated")

var ErrInvalidArgument = errors.New("invalid argument")

var ErrResourceExhausted = errors.New("resource exhausted")

var ErrKafkaPublish = errors.New("kafka publish failed")
