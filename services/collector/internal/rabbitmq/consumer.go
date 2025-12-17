package rabbitmq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	handler MessageHandler
}

type MessageHandler func(ctx context.Context, body []byte) error

type ParsedEmail struct {
	UserID      string    `json:"user_id"`
	EmailID     string    `json:"email_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
}

func NewConsumer(url, queueName string, handler MessageHandler) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = channel.ExchangeDeclare(
		"parsed_emails",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = channel.QueueBind(
		queue.Name,
		"parsed_emails",
		"parsed_emails",
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = channel.Qos(
		10,
		0,
		false,
	)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("RabbitMQ consumer initialized successfully")

	return &Consumer{
		conn:    conn,
		channel: channel,
		queue:   queue.Name,
		handler: handler,
	}, nil
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Consumer) Start(ctx context.Context) error {

	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming messages")
		return err
	}

	log.Info().Str("queue", c.queue).Msg("Started consuming messages")

	for {
		select {
		case <-ctx.Done():

			log.Info().Msg("Stopping consumer...")
			return ctx.Err()

		case msg, ok := <-msgs:
			if !ok {

				log.Error().Msg("Message channel closed")
				return ErrChannelClosed
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	messageID := msg.MessageId
	if messageID == "" {
		messageID = "unknown"
	}

	log.Debug().Msgf("Processing message: %s", messageID)

	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := c.handler(processCtx, msg.Body)

	if err != nil {

		log.Error().Err(err).Msgf("Failed to process message: %s", messageID)

		if nackErr := msg.Nack(false, true); nackErr != nil {
			log.Error().Err(nackErr).Msg("Failed to nack message")
		}
		return
	}

	if ackErr := msg.Ack(false); ackErr != nil {
		log.Error().Err(ackErr).Msg("Failed to ack message")
	}

	log.Debug().Msgf("Message processed successfully: %s", messageID)
}
