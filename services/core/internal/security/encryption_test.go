package security

import "testing"

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	enc := NewEncryptor("12345678901234567890123456789012")

	plaintext := "test-password-123"
	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("Encrypt returned empty ciphertext")
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("Decrypt = %q, want %q", decrypted, plaintext)
	}
}

func TestEncrypt_InvalidKey(t *testing.T) {
	enc := NewEncryptor("short-key")
	if _, err := enc.Encrypt("data"); err == nil {
		t.Fatal("expected error with invalid key, got nil")
	}
}
