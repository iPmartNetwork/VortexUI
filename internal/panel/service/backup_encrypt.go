package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// EncryptBackup encrypts backup data with AES-256-GCM using the provided passphrase.
// The passphrase is hashed with SHA-256 to derive the key.
// Output format: [12-byte nonce][ciphertext+tag]
func EncryptBackup(data []byte, passphrase string) ([]byte, error) {
	if passphrase == "" {
		return data, nil // no encryption
	}
	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

// DecryptBackup decrypts an AES-256-GCM encrypted backup.
func DecryptBackup(encrypted []byte, passphrase string) ([]byte, error) {
	if passphrase == "" {
		return encrypted, nil
	}
	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
