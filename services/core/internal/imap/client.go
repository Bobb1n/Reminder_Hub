package imap

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
)

func init() { imap.CharsetReader = charset.Reader }

type Client struct{ client *client.Client }

type EmailMessage struct {
	MessageID string
	From      string
	Subject   string
	BodyText  string
	Date      time.Time
	Message   *imap.Message
}

func NewIMAPClient(host string, port int, useSSL bool, timeout time.Duration) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	var c *client.Client
	var err error

	if useSSL {
		c, err = client.DialTLS(addr, &tls.Config{ServerName: host})
	} else {
		c, err = client.Dial(addr)
	}
	if err != nil {
		return nil, fmt.Errorf("connect error: %w", err)
	}

	c.Timeout = timeout
	return &Client{client: c}, nil
}

func (ic *Client) Login(email, password string) error {
	if err := ic.client.Login(email, password); err != nil {
		return fmt.Errorf("login error: %w", err)
	}
	return nil
}

func (ic *Client) Logout() error { return ic.client.Logout() }

func (ic *Client) GetUnseenMessages(since *time.Time) ([]*EmailMessage, error) {
	if _, err := ic.client.Select("INBOX", false); err != nil {
		return nil, fmt.Errorf("select inbox: %w", err)
	}

	criteria := buildSearchCriteria(since)
	seqNums, err := ic.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	if len(seqNums) == 0 {
		return nil, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(seqNums...)

	msgs := make(chan *imap.Message, 100)
	done := make(chan error, 1)

	go func() {
		done <- ic.client.Fetch(seqset, []imap.FetchItem{
			imap.FetchEnvelope, imap.FetchBody, imap.FetchFlags,
			imap.FetchInternalDate, imap.FetchRFC822,
		}, msgs)
	}()

	var result []*EmailMessage
	for msg := range msgs {
		if em, err := parseMessage(msg); err == nil {
			result = append(result, em)
		}
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	return result, nil
}

func parseMessage(msg *imap.Message) (*EmailMessage, error) {
	em := &EmailMessage{Message: msg}

	if env := msg.Envelope; env != nil {
		em.From = formatAddress(env.From)
		em.Subject = env.Subject
		em.MessageID = env.MessageId
		em.Date = env.Date
	}

	if em.Date.IsZero() && !msg.InternalDate.IsZero() {
		em.Date = msg.InternalDate
	}

	if msg.Body != nil {
		em.BodyText, _ = extractTextBody(msg)
	}

	return em, nil
}

func buildSearchCriteria(since *time.Time) *imap.SearchCriteria {
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}
	if since != nil {
		criteria.Since = *since
	} else {
		criteria.Since = time.Now().Add(-24 * time.Hour)
	}
	return criteria
}

func formatAddress(addrs []*imap.Address) string {
	var parts []string
	for _, a := range addrs {
		if a.MailboxName != "" && a.HostName != "" {
			parts = append(parts, fmt.Sprintf("%s@%s", a.MailboxName, a.HostName))
		} else if a.PersonalName != "" {
			parts = append(parts, a.PersonalName)
		}
	}
	return strings.Join(parts, ", ")
}

func extractTextBody(msg *imap.Message) (string, error) {
	for _, lit := range msg.Body {
		buf := make([]byte, lit.Len())
		n, _ := lit.Read(buf)
		if n == 0 {
			continue
		}

		content := string(buf[:n])
		if strings.Contains(strings.ToLower(content), "content-type: text/plain") {
			parts := strings.SplitN(content, "\r\n\r\n", 2)
			if len(parts) > 1 {
				return parts[1], nil
			}
		}
	}
	return "", fmt.Errorf("no text body")
}
