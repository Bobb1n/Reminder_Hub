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

	rawEmails := make([]models.RawEmail, 0, len(messages.RawEmail))
	for _, msg := range messages.RawEmail {
		// Проверяем и форматируем дату, если нужно
		dateReceived := msg.Date
		if dateReceived == "" {
			dateReceived = time.Now().Format(time.RFC3339)
		}

		syncTimestamp := msg.TimeStamp
		if syncTimestamp == "" {
			syncTimestamp = time.Now().Format(time.RFC3339)
		}

		rawEmail := models.RawEmail{
			EmailID:   msg.EmailID,
			UserID:    msg.UserID,
			MessageID: msg.MessageID,
			From:      msg.From,
			Subject:   msg.Subject,
			Text:      msg.Text,
			Date:      dateReceived,
			TimeStamp: syncTimestamp,
		}
		rawEmails = append(rawEmails, rawEmail)
	}

	// Публикуем в формате RawEmails, который ожидает analyzer-service
	rawEmailsMessage := &models.RawEmails{
		RawEmail: rawEmails,
	}

	return p.publisher.PublishMessage(rawEmailsMessage)
}
