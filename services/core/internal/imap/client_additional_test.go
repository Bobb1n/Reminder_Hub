package imap

import (
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"github.com/stretchr/testify/assert"
)

func TestFormatAddress_MultipleAddresses(t *testing.T) {
	addrs := []*imap.Address{
		{MailboxName: "user1", HostName: "example.com"},
		{MailboxName: "user2", HostName: "test.com"},
	}
	result := formatAddress(addrs)
	assert.Equal(t, "user1@example.com, user2@test.com", result)
}

func TestFormatAddress_WithPersonalName(t *testing.T) {
	addrs := []*imap.Address{
		{PersonalName: "John Doe"},
	}
	result := formatAddress(addrs)
	assert.Equal(t, "John Doe", result)
}

func TestFormatAddress_Mixed(t *testing.T) {
	addrs := []*imap.Address{
		{MailboxName: "user", HostName: "example.com"},
		{PersonalName: "John Doe"},
	}
	result := formatAddress(addrs)
	assert.Equal(t, "user@example.com, John Doe", result)
}

func TestFormatAddress_Empty(t *testing.T) {
	addrs := []*imap.Address{}
	result := formatAddress(addrs)
	assert.Equal(t, "", result)
}

func TestFormatAddress_IncompleteAddress(t *testing.T) {
	addrs := []*imap.Address{
		{MailboxName: "user"},
	}
	result := formatAddress(addrs)
	assert.Equal(t, "", result)
}

func TestBuildSearchCriteria_WithSince(t *testing.T) {
	since := time.Now().Add(-48 * time.Hour)
	criteria := buildSearchCriteria(&since)

	assert.Equal(t, []string{"\\Seen"}, criteria.WithoutFlags)
	assert.Equal(t, since, criteria.Since)
}

func TestBuildSearchCriteria_WithoutSince(t *testing.T) {
	criteria := buildSearchCriteria(nil)

	assert.Equal(t, []string{"\\Seen"}, criteria.WithoutFlags)
	expectedSince := time.Now().Add(-24 * time.Hour)
	diff := expectedSince.Sub(criteria.Since)
	if diff < 0 {
		diff = -diff
	}

	assert.True(t, diff < time.Second, "Since should be approximately 24 hours ago")
}

func TestParseMessage_WithEnvelope(t *testing.T) {
	msg := &imap.Message{
		Envelope: &imap.Envelope{
			Subject:   "Test Subject",
			MessageId: "test-id",
			Date:      time.Now(),
			From: []*imap.Address{
				{MailboxName: "test", HostName: "example.com"},
			},
		},
		InternalDate: time.Now().Add(-1 * time.Hour),
	}

	emailMsg, err := parseMessage(msg)

	assert.NoError(t, err)
	assert.Equal(t, "Test Subject", emailMsg.Subject)
	assert.Equal(t, "test-id", emailMsg.MessageID)
	assert.Equal(t, "test@example.com", emailMsg.From)
}

func TestParseMessage_WithInternalDate(t *testing.T) {
	internalDate := time.Now()
	msg := &imap.Message{
		Envelope: &imap.Envelope{
			Date: time.Time{}, // Zero time
		},
		InternalDate: internalDate,
	}

	emailMsg, err := parseMessage(msg)

	assert.NoError(t, err)
	assert.Equal(t, internalDate, emailMsg.Date)
}
