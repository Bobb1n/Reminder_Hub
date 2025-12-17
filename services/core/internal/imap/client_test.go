package imap

import (
	"testing"
	"time"

	"github.com/emersion/go-imap"
)

func TestBuildSearchCriteria_DefaultSince(t *testing.T) {
	crit := buildSearchCriteria(nil)
	if len(crit.WithoutFlags) == 0 || crit.WithoutFlags[0] != "\\Seen" {
		t.Fatalf("unexpected WithoutFlags: %#v", crit.WithoutFlags)
	}
}

func TestFormatAddress(t *testing.T) {
	addrs := []*imap.Address{{MailboxName: "user", HostName: "example.com"}}
	if got := formatAddress(addrs); got != "user@example.com" {
		t.Fatalf("formatAddress = %q, want %q", got, "user@example.com")
	}
}

func TestParseMessage_FillsFields(t *testing.T) {
	msg := &imap.Message{
		Envelope: &imap.Envelope{Subject: "subj", MessageId: "id", Date: time.Now()},
	}
	if em, err := parseMessage(msg); err != nil || em.Subject != "subj" {
		t.Fatalf("parseMessage em=%+v err=%v", em, err)
	}
}
