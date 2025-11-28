package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/models"
	"github.com/streadway/amqp"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQProducer(rabbitURL string) (*Producer, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange for raw emails
	err = channel.ExchangeDeclare(
		"raw_emails", // exchange name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Println("✅ Connected to RabbitMQ producer")
	return &Producer{conn: conn, channel: channel}, nil
}

func (p *Producer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

func (p *Producer) Channel() (*amqp.Channel, error) {
	if p.channel == nil {
		return nil, fmt.Errorf("channel is not available")
	}
	return p.channel, nil
}

func (p *Producer) PublishRawEmail(message *models.RabbitMQEmailMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.channel.Publish(
		"raw_emails", // exchange
		"raw",        // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("📤 Published raw email to RabbitMQ: %s", message.MessageID)
	return nil
}
