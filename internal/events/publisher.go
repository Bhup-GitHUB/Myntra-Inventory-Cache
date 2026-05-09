package events

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	channel *amqp.Channel
}

func NewPublisher(ctx context.Context, conn *amqp.Connection) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := DeclareTopology(ctx, ch); err != nil {
		_ = ch.Close()
		return nil, err
	}
	return &Publisher{channel: ch}, nil
}

func (p *Publisher) PublishInventoryUpdated(ctx context.Context, event InventoryUpdatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.channel.PublishWithContext(ctx, ExchangeInventoryEvents, RoutingInventoryUpdated, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (p *Publisher) Close() error {
	return p.channel.Close()
}
