package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/kafka"
	platconsumer "github.com/CARAI-REA/messanger/callService/platform/pkg/kafka/consumer"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"go.uber.org/zap"

	"sfu/internal/service"
)

type Consumer struct {
	group   sarama.ConsumerGroup
	inner   kafka.Consumer
	handler *Handler
	cancel  context.CancelFunc
}

func NewConsumer(brokers []string, groupID string, topics []string, sessions service.SessionService) (*Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_6_0_0
	cfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		group:   group,
		inner:   platconsumer.NewConsumer(group, topics, logger.Logger()),
		handler: NewHandler(sessions),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	go func() {
		err := c.inner.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
			return c.handler.Handle(ctx, msg.Value)
		})
		if err != nil && ctx.Err() == nil {
			logger.Error(ctx, "kafka consumer stopped", zap.Error(err))
		}
	}()
}

func (c *Consumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return c.group.Close()
}
