package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// FieldEncryptor provides AES-256-GCM encryption/decryption for sensitive fields.
// The master key is derived from the panel's secret configuration using SHA-256.
type FieldEncryptor struct {
	aead cipher.AEAD
}

// NewFieldEncryptor derives a 256-bit key from the panel secret and creates
// the AES-256-GCM encryptor.
func NewFieldEncryptor(panelSecret string) (*FieldEncryptor, error) {
	if panelSecret == "" {
		return nil, errors.New("panel secret must not be empty")
	}

	// Derive a 32-byte key from the secret using SHA-256.
	hash := sha256.Sum256([]byte(panelSecret))

	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &FieldEncryptor{aead: aead}, nil
}

// Encrypt encrypts plaintext and returns a base64-encoded ciphertext string.
// Each call uses a fresh random nonce (prepended to the ciphertext).
func (e *FieldEncryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := e.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decodes a base64-encoded ciphertext string and returns the plaintext.
func (e *FieldEncryptor) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	nonceSize := e.aead.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptBytes encrypts raw bytes and returns base64.
func (e *FieldEncryptor) EncryptBytes(plaintext []byte) (string, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := e.aead.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptBytes decodes base64 and decrypts to raw bytes.
func (e *FieldEncryptor) DecryptBytes(encoded string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	nonceSize := e.aead.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}
