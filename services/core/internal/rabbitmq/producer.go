package rabbitmq

import (
	"context"
	"time"

	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/models"
	pkgrabbitmq "reminder-hub/pkg/rabbitmq"

	"github.com/streadway/amqp"
)

type EmailBatchMessage struct {
	Emails        *models.RawEmails `json:"emails"`
	BatchSize     int               `json:"batch_size"`
	SyncTimestamp string            `json:"sync_timestamp"`
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

func (p *Producer) PublishEmail(message *models.RawEmail) error {
	return p.publisher.PublishMessage(message)
}

func (p *Producer) PublishEmailBatch(messages *models.RawEmails) error {
	if len(messages.RawEmail) == 0 {
		return nil
	}

	batchMessage := &EmailBatchMessage{
		Emails:        messages,
		BatchSize:     len(messages.RawEmail),
		SyncTimestamp: time.Now().Format(time.RFC3339),
	}

	return p.publisher.PublishMessage(batchMessage)
}
