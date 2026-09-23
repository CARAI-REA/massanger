package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"

	kafkapkg "github.com/CARAI-REA/messanger/callService/platform/pkg/kafka"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/kafka/producer"

	"rooms/internal/metrics"
)

const (
	eventRoomCreated         = "room_created"
	eventRoomUpdated         = "room_updated"
	eventRoomDeleted         = "room_deleted"
	eventRoomEnded           = "room_ended"
	eventParticipantJoined   = "participant_joined"
	eventParticipantLeft     = "participant_left"
)

type roomEnvelope struct {
	EventType string `json:"event_type"`
	RoomID    string `json:"room_id"`
	OwnerID   string `json:"owner_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

type participantEnvelope struct {
	EventType string          `json:"event_type"`
	RoomID    string          `json:"room_id"`
	UserID    string          `json:"user_id"`
	Timestamp string          `json:"timestamp"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

type eventProducer struct {
	sync              sarama.SyncProducer
	roomEventsTopic   kafkapkg.Producer
	participantEvents kafkapkg.Producer
}

// NewSyncProducerConfig конфиг Sarama: sync, acks=all, ретраи (room_task.md §7.4).
func NewSyncProducerConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_6_0_0
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Return.Successes = true
	return cfg
}

// NewEventProducer создаёт продюсер двух топиков на одном sync producer.
func NewEventProducer(
	brokers []string,
	roomEventsTopic, participantEventsTopic string,
	log producer.Logger,
) (*eventProducer, error) {
	cfg := NewSyncProducerConfig()
	sp, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}
	return &eventProducer{
		sync:              sp,
		roomEventsTopic:   producer.NewProducer(sp, roomEventsTopic, log),
		participantEvents: producer.NewProducer(sp, participantEventsTopic, log),
	}, nil
}

func (e *eventProducer) Close() error {
	return e.sync.Close()
}

func (e *eventProducer) PublishRoomCreated(ctx context.Context, roomID, ownerID string) error {
	return e.sendRoom(ctx, eventRoomCreated, roomID, ownerID)
}

func (e *eventProducer) PublishRoomUpdated(ctx context.Context, roomID string) error {
	return e.sendRoom(ctx, eventRoomUpdated, roomID, "")
}

func (e *eventProducer) PublishRoomDeleted(ctx context.Context, roomID string) error {
	return e.sendRoom(ctx, eventRoomDeleted, roomID, "")
}

func (e *eventProducer) PublishRoomEnded(ctx context.Context, roomID string) error {
	return e.sendRoom(ctx, eventRoomEnded, roomID, "")
}

func (e *eventProducer) sendRoom(ctx context.Context, eventType, roomID, ownerID string) error {
	body, err := json.Marshal(roomEnvelope{
		EventType: eventType,
		RoomID:    roomID,
		OwnerID:   ownerID,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		metrics.KafkaErrorsTotal.Inc()
		return err
	}
	key := []byte(roomID)
	if err := e.roomEventsTopic.Send(ctx, key, body); err != nil {
		metrics.KafkaErrorsTotal.Inc()
		return err
	}
	return nil
}

func (e *eventProducer) PublishParticipantJoined(ctx context.Context, roomID, userID string, metadata []byte) error {
	var raw json.RawMessage
	if len(metadata) > 0 {
		raw = json.RawMessage(metadata)
	}
	body, err := json.Marshal(participantEnvelope{
		EventType: eventParticipantJoined,
		RoomID:    roomID,
		UserID:    userID,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Metadata:  raw,
	})
	if err != nil {
		metrics.KafkaErrorsTotal.Inc()
		return err
	}
	key := []byte(roomID + ":" + userID)
	if err := e.participantEvents.Send(ctx, key, body); err != nil {
		metrics.KafkaErrorsTotal.Inc()
		return err
	}
	return nil
}

func (e *eventProducer) PublishParticipantLeft(ctx context.Context, roomID, userID string) error {
	body, err := json.Marshal(participantEnvelope{
		EventType: eventParticipantLeft,
		RoomID:    roomID,
		UserID:    userID,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		metrics.KafkaErrorsTotal.Inc()
		return err
	}
	key := []byte(roomID + ":" + userID)
	if err := e.participantEvents.Send(ctx, key, body); err != nil {
		metrics.KafkaErrorsTotal.Inc()
		return err
	}
	return nil
}
