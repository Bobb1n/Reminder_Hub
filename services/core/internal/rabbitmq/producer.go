package rabbitmq

import (
	"context"
	"time"

	"reminder-hub/pkg/logger"
	pkgrabbitmq "reminder-hub/pkg/rabbitmq"

	"github.com/streadway/amqp"
)

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

type Producer struct {
	publisher pkgrabbitmq.IPublisher
}

func NewProducerWithConn(conn *amqp.Connection, cfg *pkgrabbitmq.RabbitMQConfig, log *logger.CurrentLogger, ctx context.Context) (*Producer, error) {
	publisher := pkgrabbitmq.NewPublisher(ctx, cfg, conn, log)
	return &Producer{
		publisher: publisher,
	}, nil
}

func (p *Producer) Close() error {
	return nil
}

func (p *Producer) PublishEmail(message *EmailMessage) error {
	return p.publisher.PublishMessage(message)
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

	return p.publisher.PublishMessage(batchMessage)
}
