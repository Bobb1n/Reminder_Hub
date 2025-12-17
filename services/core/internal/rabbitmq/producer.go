package rabbitmq

import (
	"context"
	"time"

	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/models"
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

	// Преобразуем EmailMessage в RawEmail для совместимости с analyzer-service
	rawEmails := make([]models.RawEmail, 0, len(messages))
	for _, msg := range messages {
		dateReceived, err := time.Parse(time.RFC3339, msg.DateReceived)
		if err != nil {
			// Если не удалось распарсить, используем текущее время
			dateReceived = time.Now()
		}
		
		syncTimestamp, err := time.Parse(time.RFC3339, msg.SyncTimestamp)
		if err != nil {
			// Если не удалось распарсить, используем текущее время
			syncTimestamp = time.Now()
		}

		rawEmail := models.RawEmail{
			EmailID:   msg.EmailID,
			UserID:    msg.UserID,
			MessageID: msg.MessageID,
			From:      msg.FromAddress,
			Subject:   msg.Subject,
			Text:      msg.BodyText,
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
