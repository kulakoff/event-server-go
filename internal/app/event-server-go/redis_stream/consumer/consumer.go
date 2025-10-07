package consumer

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

type StreamMessage struct {
	ID     string                 `json:"id"`
	Values map[string]interface{} `json:"values"`
}

type EventProcessor interface {
	ProcessEvent(ctx context.Context, message *StreamMessage)
}
type ConsumerConfig struct {
	StreamName    string
	ConsumerGroup string
	ConsumerName  string
	BatchSize     int64
	BlockTime     time.Duration
}

type RedisStreamConsumer struct {
	client    *redis.Client
	config    ConsumerConfig
	logger    *slog.Logger
	processor EventProcessor
}

func NewRedisStreamConsumer(client *redis.Client, config ConsumerConfig, logger *slog.Logger, processor EventProcessor) *RedisStreamConsumer {
	return &RedisStreamConsumer{client, config, logger, processor}
}

func (c *RedisStreamConsumer) Start(ctx context.Context) error {
	// TODO: implement me
	c.logger.Info("Start RedisStreamConsumer")
	err := c.createConsumerGroup(ctx)
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	// start process
	return c.processMessages(ctx)
}

func (c *RedisStreamConsumer) createConsumerGroup(ctx context.Context) error {
	// TODO: implement me
	c.logger.Info("Start createConsumerGroup")
	//err := c.client.XGroupCreateMkStream(ctx, c.config.StreamName, c.config.ConsumerGroup, "0").Err()
	err := c.client.XGroupCreateMkStream(ctx, c.config.StreamName, c.config.ConsumerGroup, "$").Err()
	if err != nil {
		// group exist
		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			c.logger.Debug("consumer group already exists", "group", c.config.ConsumerGroup)
			return nil
		}
		return err
	}

	c.logger.Info("consumer group created", "group", c.config.ConsumerGroup)
	return nil
}

func (c *RedisStreamConsumer) processMessages(ctx context.Context) error {
	// TODO: implement me
	c.logger.Info("Start processMessages")

	return nil
}

func (c *RedisStreamConsumer) readAndProcessBatch() error {
	// TODO: implement me
	c.logger.Info("Start readAndProcessBatch")
	return nil
}
