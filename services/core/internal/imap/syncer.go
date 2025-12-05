package imap

import (
	"context"
	"time"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/database"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/rabbitmq"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/security"
	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

	var processed int
	for _, msg := range msgs {
		if err := s.processMessage(integration, msg, logger); err != nil {
			logger.Warn().Err(err).Str("msg_id", msg.MessageID).Msg("Process failed")
			continue
		}
		processed++
	}

	ctx := context.Background()
	if err := s.db.UpdateLastSync(ctx, integration.ID); err != nil {
		return errUpdateLastSync(integration.ID, err)
	}

	logger.Info().Int("processed", processed).Msg("Sync done")
	return nil
}

func (s *Syncer) processMessage(integration *database.EmailIntegration, msg *EmailMessage, logger zerolog.Logger) error {
	ctx := context.Background()

	exists, err := s.db.EmailExists(ctx, integration.UserID, msg.MessageID)
	if err != nil {
		return errCheckEmailExistence(msg.MessageID, err)
	}
	if exists {
		return nil
	}

	emailID, err := util.GenerateUUID()
	if err != nil {
		return errGenerateUUID(msg.MessageID, err)
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
		return errSaveEmail(emailID, err)
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

	if err := s.rabbit.PublishEmail(rabbitMsg); err != nil {
		return errPublishEmail(emailID, err)
	}

	logger.Info().Str("email_id", emailID).Str("from", msg.From).Msg("Email processed")
	return nil
}
