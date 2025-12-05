package security

import (
	"errors"
	"fmt"
)

func errFailedToCreateCipher(err error) error {
	return fmt.Errorf("failed to create cipher: %w", err)
}

func errFailedToCreateGCM(err error) error {
	return fmt.Errorf("failed to create GCM: %w", err)
}

func errFailedToGenerateNonce(err error) error {
	return fmt.Errorf("failed to generate nonce: %w", err)
}

func errFailedToDecodeBase64(err error) error {
	return fmt.Errorf("failed to decode base64: %w", err)
}

func errCiphertextTooShort() error {
	return errors.New("ciphertext too short")
}

func errFailedToDecrypt(err error) error {
	return fmt.Errorf("failed to decrypt: %w", err)
}
