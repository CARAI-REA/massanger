package kafka

import "context"

// NoopProducer заглушка для тестов и окружений без Kafka.
type NoopProducer struct{}

func (NoopProducer) PublishRoomCreated(context.Context, string, string) error         { return nil }
func (NoopProducer) PublishRoomUpdated(context.Context, string) error                 { return nil }
func (NoopProducer) PublishRoomDeleted(context.Context, string) error                   { return nil }
func (NoopProducer) PublishRoomEnded(context.Context, string) error                    { return nil }
func (NoopProducer) PublishParticipantJoined(context.Context, string, string, []byte) error {
	return nil
}
func (NoopProducer) PublishParticipantLeft(context.Context, string, string) error { return nil }
func (NoopProducer) Close() error                                                 { return nil }
