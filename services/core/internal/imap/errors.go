package imap

import (
	"fmt"
)

func errDecryptPassword(id string, err error) error {
	return fmt.Errorf("decrypt password for integration %s: %w", id, err)
}

func errCreateIMAPClient(host string, port int, err error) error {
	return fmt.Errorf("create IMAP client for %s:%d: %w", host, port, err)
}

func errLoginToIMAP(host, email string, err error) error {
	return fmt.Errorf("login to IMAP %s as %s: %w", host, email, err)
}

func errGetMessages(email string, err error) error {
	return fmt.Errorf("get messages for %s: %w", email, err)
}

func errUpdateLastSync(id string, err error) error {
	return fmt.Errorf("update last sync for integration %s: %w", id, err)
}

func errCheckEmailExistence(msgID string, err error) error {
	return fmt.Errorf("check email existence for message %s: %w", msgID, err)
}

func errGenerateUUID(msgID string, err error) error {
	return fmt.Errorf("generate UUID for message %s: %w", msgID, err)
}

func errSaveEmail(emailID string, err error) error {
	return fmt.Errorf("save email %s to database: %w", emailID, err)
}

func errPublishEmail(emailID string, err error) error {
	return fmt.Errorf("publish email %s to RabbitMQ: %w", emailID, err)
}
