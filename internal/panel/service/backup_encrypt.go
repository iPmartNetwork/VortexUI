package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// BackupDestination is where encrypted backups are stored.
type BackupDestination string

const (
	BackupDestLocal    BackupDestination = "local"
	BackupDestS3      BackupDestination = "s3"
	BackupDestTelegram BackupDestination = "telegram"
)

// BackupEncryptConfig holds backup encryption and destination settings.
type BackupEncryptConfig struct {
	Secret       string            `json:"secret"`
	Destination  BackupDestination `json:"destination"`
	LocalPath    string            `json:"local_path,omitempty"`
	S3Bucket     string            `json:"s3_bucket,omitempty"`
	S3Region     string            `json:"s3_region,omitempty"`
	S3Endpoint   string            `json:"s3_endpoint,omitempty"`
	TelegramChat int64             `json:"telegram_chat,omitempty"`
}

// BackupPayload is the data structure that gets encrypted in a backup.
type BackupPayload struct {
	Version   string         `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	Data      map[string]any `json:"data"`
}

// BackupEncryptService handles creating and restoring encrypted backups.
type BackupEncryptService struct {
	config BackupEncryptConfig
}

// NewBackupEncryptService creates the service.
func NewBackupEncryptService(config BackupEncryptConfig) *BackupEncryptService {
	return &BackupEncryptService{config: config}
}

// CreateBackup serializes, compresses, and encrypts the backup payload.
func (s *BackupEncryptService) CreateBackup(ctx context.Context, data map[string]any) ([]byte, error) {
	payload := BackupPayload{
		Version:   "1.0",
		CreatedAt: time.Now(),
		Data:      data,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// Compress with gzip
	var compressed bytes.Buffer
	gw := gzip.NewWriter(&compressed)
	if _, err := gw.Write(jsonData); err != nil {
		return nil, fmt.Errorf("compress: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip: %w", err)
	}

	// Encrypt with AES-256-GCM
	encrypted, err := s.encrypt(compressed.Bytes())
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	return encrypted, nil
}

// RestoreBackup decrypts, decompresses, and deserializes a backup.
func (s *BackupEncryptService) RestoreBackup(ctx context.Context, encrypted []byte) (*BackupPayload, error) {
	// Decrypt
	compressed, err := s.decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	// Decompress
	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("open gzip: %w", err)
	}
	defer gr.Close() //nolint:errcheck

	jsonData, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("decompress: %w", err)
	}

	// Deserialize
	var payload BackupPayload
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	return &payload, nil
}

func (s *BackupEncryptService) encrypt(plaintext []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(s.config.Secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (s *BackupEncryptService) decrypt(ciphertext []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(s.config.Secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aead.Open(nil, nonce, ct, nil)
}

// EncryptBackup encrypts raw bytes with a passphrase using AES-256-GCM.
// Compatible with the backup export/import pipeline.
func EncryptBackup(data []byte, passphrase string) ([]byte, error) {
	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return aead.Seal(nonce, nonce, data, nil), nil
}

// DecryptBackup decrypts data encrypted by EncryptBackup.
func DecryptBackup(data []byte, passphrase string) ([]byte, error) {
	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aead.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return aead.Open(nil, nonce, ciphertext, nil)
}
