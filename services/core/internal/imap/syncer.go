package imap

import (
	"context"
	"time"

	"core/internal/database"
	"core/internal/rabbitmq"
	"core/internal/security"
	"core/internal/util"
	"reminder-hub/pkg/logger"
	"reminder-hub/pkg/models"
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
		s.log.Error(ctx, "IMAP client error", "error", err)
		return errCreateIMAPClient(integration.ImapHost, integration.ImapPort, err)
	}
	defer func() {
		if err := imapClient.Logout(); err != nil {
			s.log.Warn(ctx, "Logout failed", "error", err)
		}
	}()

	if err := imapClient.Login(integration.EmailAddress, pass); err != nil {
		s.log.Error(ctx, "IMAP login error", "error", err)
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

	var currentBatch *models.RawEmails
	var processed int

	for _, msg := range msgs {
		rabbitMsg, err := s.processMessage(ctx, integration, msg)
		if err != nil {
			s.log.Warn(ctx, "Process failed", "error", err, "msg_id", msg.MessageID)
			continue
		}
		if rabbitMsg != nil {
			if currentBatch == nil {
				currentBatch = &models.RawEmails{}
			}
			currentBatch.RawEmail = append(currentBatch.RawEmail, *rabbitMsg)
			processed++

			if len(currentBatch.RawEmail) >= maxBatchSize {
				if err := s.rabbit.PublishEmailBatch(currentBatch); err != nil {
					return errPublishEmail(integration.ID, err)
				}
				s.log.Info(ctx, "Batch published", "batch_size", len(currentBatch.RawEmail))
				currentBatch = nil
			}
		}
	}

	if len(currentBatch.RawEmail) > 0 {
		if err := s.rabbit.PublishEmailBatch(currentBatch); err != nil {
			return errPublishEmail(integration.ID, err)
		}
		s.log.Info(ctx, "Final batch published", "batch_size", len(currentBatch.RawEmail))
	}

	if err := s.db.UpdateLastSync(ctx, integration.ID); err != nil {
		return errUpdateLastSync(integration.ID, err)
	}

	s.log.Info(ctx, "Sync done", "processed", processed)
	return nil
}

func (s *Syncer) processMessage(ctx context.Context, integration *database.EmailIntegration, msg *EmailMessage) (*models.RawEmail, error) {

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

	rabbitMsg := &models.RawEmail{
		EmailID:   email.ID,
		UserID:    email.UserID,
		MessageID: email.MessageID,
		From:      email.FromAddress,
		Subject:   email.Subject,
		Text:      email.BodyText,
		Date:      email.DateReceived.Format(time.RFC3339),
		TimeStamp: time.Now().Format(time.RFC3339),
	}

	s.log.Info(ctx, "Email processed", "email_id", emailID, "from", msg.From)
	return rabbitMsg, nil
}
