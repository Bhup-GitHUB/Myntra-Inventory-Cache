package events

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	channel *amqp.Channel
}

func NewConsumer(ctx context.Context, conn *amqp.Connection) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := DeclareTopology(ctx, ch); err != nil {
		_ = ch.Close()
		return nil, err
	}
	if err := ch.Qos(10, 0, false); err != nil {
		_ = ch.Close()
		return nil, err
	}
	return &Consumer{channel: ch}, nil
}

func (c *Consumer) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	return c.channel.ConsumeWithContext(ctx, QueueCacheInvalidator, "", false, false, false, false, nil)
}

func DecodeInventoryUpdated(delivery amqp.Delivery) (InventoryUpdatedEvent, error) {
	var event InventoryUpdatedEvent
	err := json.Unmarshal(delivery.Body, &event)
	return event, err
}

func (c *Consumer) Close() error {
	return c.channel.Close()
}
