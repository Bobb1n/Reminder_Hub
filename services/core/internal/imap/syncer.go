package imap

import (
	"context"
	"time"

	"core/internal/database"
	"core/internal/rabbitmq"
	"core/internal/security"
	"core/internal/util"
	"reminder-hub/pkg/logger"
)

const maxBatchSize = 7

type Syncer struct {
	db        *database.DB
	rabbit    *rabbitmq.Producer
	encryptor security.Encryptor
	timeout   time.Duration
	log       *logger.CurrentLogger
}

func NewSyncer(db *database.DB, rabbit *rabbitmq.Producer, enc security.Encryptor, timeout time.Duration, log *logger.CurrentLogger) *Syncer {
	return &Syncer{db: db, rabbit: rabbit, encryptor: enc, timeout: timeout, log: log}
}

func (s *Syncer) SyncIntegration(integration *database.EmailIntegration) error {
	ctx := context.Background()
	ctx = logger.WithRequestID(ctx, integration.ID)

	pass, err := s.encryptor.Decrypt(integration.Password)
	if err != nil {
		return errDecryptPassword(integration.ID, err)
	}

	imapClient, err := NewIMAPClient(integration.ImapHost, integration.ImapPort, integration.UseSSL, s.timeout)
	if err != nil {
		return errCreateIMAPClient(integration.ImapHost, integration.ImapPort, err)
	}
	defer func() {
		if err := imapClient.Logout(); err != nil {
			s.log.Warn(ctx, "Logout failed", "error", err)
		}
	}()

	if err := imapClient.Login(integration.EmailAddress, pass); err != nil {
		return errLoginToIMAP(integration.ImapHost, integration.EmailAddress, err)
	}

	var since *time.Time
	if integration.LastSyncAt != nil {
		since = integration.LastSyncAt
	}

	msgs, err := imapClient.GetUnseenMessages(since)
	if err != nil {
		return errGetMessages(integration.EmailAddress, err)
	}

	s.log.Info(ctx, "Messages found", "count", len(msgs), "email", integration.EmailAddress)

	if len(msgs) == 0 {
		if err := s.db.UpdateLastSync(ctx, integration.ID); err != nil {
			return errUpdateLastSync(integration.ID, err)
		}
		s.log.Info(ctx, "No messages to process")
		return nil
	}

	var currentBatch []*rabbitmq.EmailMessage
	var processed int

	for _, msg := range msgs {
		rabbitMsg, err := s.processMessage(ctx, integration, msg)
		if err != nil {
			s.log.Warn(ctx, "Process failed", "error", err, "msg_id", msg.MessageID)
			continue
		}
		if rabbitMsg != nil {
			currentBatch = append(currentBatch, rabbitMsg)
			processed++

			if len(currentBatch) >= maxBatchSize {
				if err := s.rabbit.PublishEmailBatch(currentBatch); err != nil {
					return errPublishEmail(integration.ID, err)
				}
				s.log.Info(ctx, "Batch published", "batch_size", len(currentBatch))
				currentBatch = nil
			}
		}
	}

	if len(currentBatch) > 0 {
		if err := s.rabbit.PublishEmailBatch(currentBatch); err != nil {
			return errPublishEmail(integration.ID, err)
		}
		s.log.Info(ctx, "Final batch published", "batch_size", len(currentBatch))
	}

	if err := s.db.UpdateLastSync(ctx, integration.ID); err != nil {
		return errUpdateLastSync(integration.ID, err)
	}

	s.log.Info(ctx, "Sync done", "processed", processed)
	return nil
}

func (s *Syncer) processMessage(ctx context.Context, integration *database.EmailIntegration, msg *EmailMessage) (*rabbitmq.EmailMessage, error) {

	exists, err := s.db.EmailExists(ctx, integration.UserID, msg.MessageID)
	if err != nil {
		return nil, errCheckEmailExistence(msg.MessageID, err)
	}
	if exists {
		return nil, nil
	}

	emailID, err := util.GenerateUUID()
	if err != nil {
		return nil, errGenerateUUID(msg.MessageID, err)
	}

	email := &database.EmailRaw{
		ID:           emailID,
		UserID:       integration.UserID,
		MessageID:    msg.MessageID,
		FromAddress:  msg.From,
		Subject:      msg.Subject,
		BodyText:     msg.BodyText,
		DateReceived: msg.Date,
		Processed:    false,
	}

	if err := s.db.SaveEmail(ctx, email); err != nil {
		return nil, errSaveEmail(emailID, err)
	}

	rabbitMsg := &rabbitmq.EmailMessage{
		EmailID:       email.ID,
		UserID:        email.UserID,
		MessageID:     email.MessageID,
		FromAddress:   email.FromAddress,
		Subject:       email.Subject,
		BodyText:      email.BodyText,
		DateReceived:  email.DateReceived.Format(time.RFC3339),
		SyncTimestamp: time.Now().Format(time.RFC3339),
	}

	s.log.Info(ctx, "Email processed", "email_id", emailID, "from", msg.From)
	return rabbitMsg, nil
}
