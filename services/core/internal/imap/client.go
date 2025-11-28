package imap

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/Bobb1n/Reminder_Hub/tree/develop/services/core/internal/models"
	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
)

type IMAPClient struct {
	config *models.EmailIntegration
	client *client.Client
}

func NewIMAPClient(config *models.EmailIntegration) (*IMAPClient, error) {
	var c *client.Client
	var err error

	addr := fmt.Sprintf("%s:%d", config.IMAPHost, config.IMAPPort)

	if config.UseSSL {
		c, err = client.DialTLS(addr, &tls.Config{
			ServerName: config.IMAPHost,
		})
	} else {
		c, err = client.Dial(addr)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	// Send ID command for better compatibility
	idClient := id.NewClient(c)
	idClient.ID(id.ID{"name": "ReminderHub", "version": "1.0.0"})

	log.Printf("✅ Connected to IMAP server: %s", config.IMAPHost)
	return &IMAPClient{config: config, client: c}, nil
}

func (ic *IMAPClient) Login() error {
	// Password authentication
	err := ic.client.Login(ic.config.EmailAddress, ic.config.Password)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	log.Printf("✅ Authenticated with email: %s", ic.config.EmailAddress)
	return nil
}

func (ic *IMAPClient) FetchEmails(since time.Time) ([]*models.EmailRaw, error) {
	// Select INBOX
	_, err := ic.client.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}

	// Search for recent emails (since last sync)
	criteria := imap.NewSearchCriteria()
	criteria.Since = since
	criteria.WithoutFlags = []string{"\\Deleted"}

	seqNums, err := ic.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search emails: %w", err)
	}

	if len(seqNums) == 0 {
		log.Printf("📭 No new emails since %s", since.Format("2006-01-02 15:04:05"))
		return []*models.EmailRaw{}, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	// Fetch only needed fields for performance
	items := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchBodyStructure,
		imap.FetchInternalDate,
		imap.FetchFlags,
		imap.FetchUid,
	}

	go func() {
		done <- ic.client.Fetch(seqSet, items, messages)
	}()

	var emails []*models.EmailRaw
	for msg := range messages {
		email := &models.EmailRaw{
			UserID:       ic.config.UserID,
			MessageID:    msg.Envelope.MessageId,
			FromAddress:  extractSender(msg.Envelope),
			Subject:      msg.Envelope.Subject,
			DateReceived: msg.InternalDate,
		}

		// Extract body text if available
		if text := ic.extractBodyText(msg); text != "" {
			email.BodyText = text
		}

		emails = append(emails, email)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	log.Printf("📧 Fetched %d new emails for %s", len(emails), ic.config.EmailAddress)
	return emails, nil
}

func (ic *IMAPClient) extractBodyText(msg *imap.Message) string {
	// Simplified body extraction - in production you'd want to handle different MIME types
	// For now, we'll try to get the first text/plain part
	for _, part := range msg.Body {
		// This is a simplified approach - you might want to implement proper MIME parsing
		if part != nil {
			// In a real implementation, you'd read the body section and parse it
			// For now, return a placeholder
			return "Email body content - implement proper MIME parsing"
		}
	}
	return "No body content"
}

func extractSender(envelope *imap.Envelope) string {
	if envelope == nil || len(envelope.From) == 0 {
		return "unknown@unknown.com"
	}

	from := envelope.From[0]
	if from.MailboxName != "" && from.HostName != "" {
		return fmt.Sprintf("%s@%s", from.MailboxName, from.HostName)
	}

	return from.Address()
}

func (ic *IMAPClient) Logout() error {
	return ic.client.Logout()
}
