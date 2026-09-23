package kafka

import "context"

// EventProducer публикует доменные события в Kafka (room_task.md §13.5, без outbox).
type EventProducer interface {
	PublishRoomCreated(ctx context.Context, roomID, ownerID string) error
	PublishRoomUpdated(ctx context.Context, roomID string) error
	PublishRoomDeleted(ctx context.Context, roomID string) error
	PublishRoomEnded(ctx context.Context, roomID string) error
	PublishParticipantJoined(ctx context.Context, roomID, userID string, metadata []byte) error
	PublishParticipantLeft(ctx context.Context, roomID, userID string) error
	Close() error
}
