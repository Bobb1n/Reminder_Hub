package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

type Encryptor interface {
	Encrypt(text string) (string, error)
	Decrypt(cipherText string) (string, error)
}

type encryptor struct {
	key []byte
}

func NewEncryptor(key string) Encryptor {
	return &encryptor{key: []byte(key)}
}

func (e *encryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", errFailedToCreateCipher(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errFailedToCreateGCM(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errFailedToGenerateNonce(err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *encryptor) Decrypt(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", errFailedToDecodeBase64(err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", errFailedToCreateCipher(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errFailedToCreateGCM(err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errCiphertextTooShort()
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errFailedToDecrypt(err)
	}

	return string(plaintext), nil
}
