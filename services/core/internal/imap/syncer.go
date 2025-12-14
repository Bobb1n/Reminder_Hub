package imap

import (
	"context"
	"time"

	"core/internal/database"
	"core/internal/rabbitmq"
	"core/internal/security"
	"core/internal/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const maxBatchSize = 7

type Syncer struct {
	db        *database.DB
	rabbit    *rabbitmq.Producer
	encryptor security.Encryptor
	timeout   time.Duration
}

func NewSyncer(db *database.DB, rabbit *rabbitmq.Producer, enc security.Encryptor, timeout time.Duration) *Syncer {
	return &Syncer{db: db, rabbit: rabbit, encryptor: enc, timeout: timeout}
}

func (s *Syncer) SyncIntegration(integration *database.EmailIntegration) error {
	logger := log.With().Str("id", integration.ID).Str("email", integration.EmailAddress).Logger()

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
			logger.Warn().Err(err).Msg("Logout failed")
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

	logger.Info().Int("count", len(msgs)).Msg("Messages found")

	if len(msgs) == 0 {
		ctx := context.Background()
		if err := s.db.UpdateLastSync(ctx, integration.ID); err != nil {
			return errUpdateLastSync(integration.ID, err)
		}
		logger.Info().Msg("No messages to process")
		return nil
	}

	var currentBatch []*rabbitmq.EmailMessage
	var processed int

	for _, msg := range msgs {
		rabbitMsg, err := s.processMessage(integration, msg, logger)
		if err != nil {
			logger.Warn().Err(err).Str("msg_id", msg.MessageID).Msg("Process failed")
			continue
		}
		if rabbitMsg != nil {
			currentBatch = append(currentBatch, rabbitMsg)
			processed++

			if len(currentBatch) >= maxBatchSize {
				if err := s.rabbit.PublishEmailBatch(currentBatch); err != nil {
					return errPublishEmail(integration.ID, err)
				}
				logger.Info().Int("batch_size", len(currentBatch)).Msg("Batch published")
				currentBatch = nil
			}
		}
	}

	if len(currentBatch) > 0 {
		if err := s.rabbit.PublishEmailBatch(currentBatch); err != nil {
			return errPublishEmail(integration.ID, err)
		}
		logger.Info().Int("batch_size", len(currentBatch)).Msg("Final batch published")
	}

	ctx := context.Background()
	if err := s.db.UpdateLastSync(ctx, integration.ID); err != nil {
		return errUpdateLastSync(integration.ID, err)
	}

	logger.Info().Int("processed", processed).Msg("Sync done")
	return nil
}

func (s *Syncer) processMessage(integration *database.EmailIntegration, msg *EmailMessage, logger zerolog.Logger) (*rabbitmq.EmailMessage, error) {
	ctx := context.Background()

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

	logger.Info().Str("email_id", emailID).Str("from", msg.From).Msg("Email processed")
	return rabbitMsg, nil
}