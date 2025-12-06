package rabbitmq

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type EmailMessage struct {
	EmailID       string `json:"email_id"`
	UserID        string `json:"user_id"`
	MessageID     string `json:"message_id"`
	FromAddress   string `json:"from_address"`
	Subject       string `json:"subject"`
	BodyText      string `json:"body_text"`
	DateReceived  string `json:"date_received"`
	SyncTimestamp string `json:"sync_timestamp"`
}

type EmailBatchMessage struct {
	Emails        []*EmailMessage `json:"emails"`
	BatchSize     int             `json:"batch_size"`
	SyncTimestamp string          `json:"sync_timestamp"`
}

func NewProducer(url string) (*Producer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = channel.ExchangeDeclare(
		"email.raw",
		"topic",
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
		"email.raw.process",
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
		"email.raw.*",
		"email.raw",
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("RabbitMQ producer initialized successfully")

	return &Producer{
		conn:    conn,
		channel: channel,
	}, nil
}

func (p *Producer) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *Producer) PublishEmail(message *EmailMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal email message")
		return err
	}

	err = p.channel.Publish(
		"email.raw",
		"email.raw.process",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to publish email to RabbitMQ")
		return err
	}

	log.Debug().Msgf("Published email: %s", message.EmailID)
	return nil
}

func (p *Producer) PublishEmailBatch(messages []*EmailMessage) error {
	if len(messages) == 0 {
		return nil
	}

	batchMessage := &EmailBatchMessage{
		Emails:        messages,
		BatchSize:     len(messages),
		SyncTimestamp: time.Now().Format(time.RFC3339),
	}

	body, err := json.Marshal(batchMessage)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal batch message")
		return err
	}

	err = p.channel.Publish(
		"email.raw",
		"email.raw.batch",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to publish batch to RabbitMQ")
		return err
	}

	log.Info().Msgf("Published batch of %d emails", len(messages))
	return nil
}
