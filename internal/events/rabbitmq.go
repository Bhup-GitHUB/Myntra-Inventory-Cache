package events

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Dial(url string) (*amqp.Connection, error) {
	return amqp.Dial(url)
}

func DeclareTopology(ctx context.Context, ch *amqp.Channel) error {
	_ = ctx
	if err := ch.ExchangeDeclare(ExchangeInventoryEvents, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(QueueCacheInvalidator, true, false, false, false, nil); err != nil {
		return err
	}
	return ch.QueueBind(QueueCacheInvalidator, RoutingInventoryUpdated, ExchangeInventoryEvents, false, nil)
}
