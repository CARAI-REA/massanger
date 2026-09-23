package env

import "github.com/caarlos0/env/v11"

type kafkaEnvConfig struct {
	Brokers                []string `env:"KAFKA_BROKERS" envDefault:""`
	RoomEventsTopic        string   `env:"KAFKA_ROOM_EVENTS_TOPIC" envDefault:"room-events"`
	ParticipantEventsTopic string   `env:"KAFKA_PARTICIPANT_EVENTS_TOPIC" envDefault:"participant-events"`
	ConsumerGroup          string   `env:"KAFKA_CONSUMER_GROUP" envDefault:"signaling"`
	Enabled                bool     `env:"KAFKA_CONSUMER_ENABLED" envDefault:"true"`
}

type kafkaConfig struct {
	raw kafkaEnvConfig
}

func NewKafkaConfig() (*kafkaConfig, error) {
	var raw kafkaEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &kafkaConfig{raw: raw}, nil
}

func (cfg *kafkaConfig) Brokers() []string              { return cfg.raw.Brokers }
func (cfg *kafkaConfig) RoomEventsTopic() string        { return cfg.raw.RoomEventsTopic }
func (cfg *kafkaConfig) ParticipantEventsTopic() string { return cfg.raw.ParticipantEventsTopic }
func (cfg *kafkaConfig) ConsumerGroup() string          { return cfg.raw.ConsumerGroup }
func (cfg *kafkaConfig) Enabled() bool                  { return cfg.raw.Enabled }
