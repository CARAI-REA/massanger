package env

import "github.com/caarlos0/env/v11"

type kafkaEnvConfig struct {
	Brokers                []string `env:"KAFKA_BROKERS,required"`
	RoomEventsTopic        string   `env:"ROOM_EVENTS_TOPIC,required"`
	ParticipantEventsTopic string   `env:"PARTICIPANT_EVENTS_TOPIC,required"`
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

func (cfg *kafkaConfig) Brokers() []string {
	return cfg.raw.Brokers
}

func (cfg *kafkaConfig) RoomEventsTopic() string {
	return cfg.raw.RoomEventsTopic
}

func (cfg *kafkaConfig) ParticipantEventsTopic() string {
	return cfg.raw.ParticipantEventsTopic
}
