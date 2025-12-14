package rabbitmq

import (
	"fmt"
)

func errFailedToConnect(err error) error {
	return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
}

func errFailedToOpenChannel(err error) error {
	return fmt.Errorf("failed to open RabbitMQ channel: %w", err)
}

func errFailedToDeclareExchange(err error) error {
	return fmt.Errorf("failed to declare exchange: %w", err)
}

func errFailedToDeclareQueue(err error) error {
	return fmt.Errorf("failed to declare queue: %w", err)
}

func errFailedToBindQueue(err error) error {
	return fmt.Errorf("failed to bind queue: %w", err)
}

func errFailedToDeclareDLQ(err error) error {
	return fmt.Errorf("failed to declare dead letter queue: %w", err)
}

func errFailedToBindDLQ(err error) error {
	return fmt.Errorf("failed to bind dead letter queue: %w", err)
}

func errFailedToMarshalMessage(err error) error {
	return fmt.Errorf("failed to marshal message: %w", err)
}

func errFailedToPublishMessage(err error) error {
	return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
}
