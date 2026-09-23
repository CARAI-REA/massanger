package env

import "github.com/caarlos0/env/v11"

type kafkaEnvConfig struct {
	Brokers                []string `env:"KAFKA_BROKERS" envDefault:""`
	RoomEventsTopic        string   `env:"KAFKA_ROOM_EVENTS_TOPIC" envDefault:"room-events"`
	ParticipantEventsTopic string   `env:"KAFKA_PARTICIPANT_EVENTS_TOPIC" envDefault:"participant-events"`
	ConsumerGroup          string   `env:"KAFKA_CONSUMER_GROUP" envDefault:"sfu"`
	Enabled                bool     `env:"KAFKA_CONSUMER_ENABLED" envDefault:"true"`
}

type kafkaConfig struct{ raw kafkaEnvConfig }

func NewKafkaConfig() (*kafkaConfig, error) {
	var raw kafkaEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &kafkaConfig{raw: raw}, nil
}

func (c *kafkaConfig) Brokers() []string              { return c.raw.Brokers }
func (c *kafkaConfig) RoomEventsTopic() string        { return c.raw.RoomEventsTopic }
func (c *kafkaConfig) ParticipantEventsTopic() string { return c.raw.ParticipantEventsTopic }
func (c *kafkaConfig) ConsumerGroup() string          { return c.raw.ConsumerGroup }
func (c *kafkaConfig) Enabled() bool                  { return c.raw.Enabled }
