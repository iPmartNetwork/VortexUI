package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"
)

// AutoBackup periodically exports a backup and sends it to configured
// destinations (Telegram chat and/or S3-compatible storage).
type AutoBackup struct {
	backup   *BackupService
	interval time.Duration
	log      *slog.Logger

	// Telegram destination (optional).
	TelegramToken  string
	TelegramChatID string

	// S3 destination (optional).
	S3Endpoint  string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
}

// NewAutoBackup wires the auto-backup loop.
func NewAutoBackup(backup *BackupService, interval time.Duration, log *slog.Logger) *AutoBackup {
	if interval == 0 {
		interval = 24 * time.Hour
	}
	if log == nil {
		log = slog.Default()
	}
	return &AutoBackup{backup: backup, interval: interval, log: log}
}

// Run ticks until ctx is cancelled.
func (ab *AutoBackup) Run(ctx context.Context) {
	ticker := time.NewTicker(ab.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ab.tick(ctx)
		}
	}
}

func (ab *AutoBackup) tick(ctx context.Context) {
	data, err := ab.backup.Export(ctx)
	if err != nil {
		ab.log.Warn("auto-backup export failed", "err", err)
		return
	}
	raw, _ := json.Marshal(data)
	filename := fmt.Sprintf("vortexui-backup-%s.json", time.Now().Format("2006-01-02"))

	if ab.TelegramToken != "" && ab.TelegramChatID != "" {
		if err := ab.sendToTelegram(ctx, raw, filename); err != nil {
			ab.log.Warn("auto-backup telegram send failed", "err", err)
		} else {
			ab.log.Info("auto-backup sent to telegram")
		}
	}

	if ab.S3Endpoint != "" && ab.S3Bucket != "" {
		if err := ab.uploadToS3(ctx, raw, filename); err != nil {
			ab.log.Warn("auto-backup S3 upload failed", "err", err)
		} else {
			ab.log.Info("auto-backup uploaded to S3", "bucket", ab.S3Bucket)
		}
	}
}

func (ab *AutoBackup) sendToTelegram(ctx context.Context, data []byte, filename string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", ab.TelegramToken)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("chat_id", ab.TelegramChatID)
	_ = writer.WriteField("caption", "📦 Auto backup — "+time.Now().Format("2006-01-02 15:04"))
	part, err := writer.CreateFormFile("document", filename)
	if err != nil {
		return err
	}
	_, _ = part.Write(data)
	_ = writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram: status %d", resp.StatusCode)
	}
	return nil
}

func (ab *AutoBackup) uploadToS3(ctx context.Context, data []byte, filename string) error {
	// Simple PUT to S3-compatible endpoint (MinIO, Cloudflare R2, etc.)
	url := fmt.Sprintf("%s/%s/%s", ab.S3Endpoint, ab.S3Bucket, filename)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// Basic auth for simple S3 endpoints (production: use AWS SDK v4 signing)
	if ab.S3AccessKey != "" {
		req.SetBasicAuth(ab.S3AccessKey, ab.S3SecretKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("s3: status %d", resp.StatusCode)
	}
	return nil
}
