package postgres

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

const (
	fullBackupManifestName = "manifest.json"
	fullBackupDumpName     = "database.dump"
)

// DumpDatabase runs pg_dump in custom format (-Fc) against databaseURL.
func DumpDatabase(ctx context.Context, databaseURL string) ([]byte, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	pgDump, err := exec.LookPath("pg_dump")
	if err != nil {
		return nil, fmt.Errorf("pg_dump not found in PATH: install postgresql-client for full database backup")
	}
	tmp, err := os.CreateTemp("", "vortex-dump-*.dump")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	_ = tmp.Close()

	args := []string{"-Fc", "--no-owner", "--no-acl", "-f", tmpPath, databaseURL}
	cmd := exec.CommandContext(ctx, pgDump, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pg_dump failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return os.ReadFile(tmpPath)
}

// RestoreDatabase runs pg_restore --clean --if-exists against databaseURL.
func RestoreDatabase(ctx context.Context, databaseURL string, dump []byte) error {
	if strings.TrimSpace(databaseURL) == "" {
		return fmt.Errorf("database URL is required")
	}
	if len(dump) == 0 {
		return fmt.Errorf("empty database dump")
	}
	pgRestore, err := exec.LookPath("pg_restore")
	if err != nil {
		return fmt.Errorf("pg_restore not found in PATH: install postgresql-client for full database restore")
	}
	tmp, err := os.CreateTemp("", "vortex-restore-*.dump")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(dump); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	args := []string{
		"--clean", "--if-exists", "--no-owner", "--no-acl",
		"-d", databaseURL, tmpPath,
	}
	cmd := exec.CommandContext(ctx, pgRestore, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		// pg_restore may exit non-zero with warnings; treat hard failures only when output suggests it.
		msg := strings.TrimSpace(string(out))
		if msg != "" && !strings.Contains(msg, "warning") {
			return fmt.Errorf("pg_restore failed: %w: %s", err, msg)
		}
		if err != nil && !strings.Contains(strings.ToLower(msg), "errors ignored on restore") {
			return fmt.Errorf("pg_restore failed: %w: %s", err, msg)
		}
	}
	return nil
}

// PackFullBackup creates a gzip tar archive with manifest + pg_dump payload.
func PackFullBackup(manifest domain.BackupManifest, dump []byte) ([]byte, error) {
	manifest.Format = domain.BackupFormatFull
	manifest.ExportedAt = time.Now().UTC()
	rawManifest, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := writeTarFile(tw, fullBackupManifestName, rawManifest); err != nil {
		return nil, err
	}
	if err := writeTarFile(tw, fullBackupDumpName, dump); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnpackFullBackup extracts manifest and dump from a gzip tar archive.
func UnpackFullBackup(data []byte) (manifest domain.BackupManifest, dump []byte, err error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return domain.BackupManifest{}, nil, fmt.Errorf("invalid gzip archive: %w", err)
	}
	defer func() {
		if closeErr := gz.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return domain.BackupManifest{}, nil, err
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			return domain.BackupManifest{}, nil, err
		}
		name := filepath.Base(hdr.Name)
		switch name {
		case fullBackupManifestName:
			if err := json.Unmarshal(body, &manifest); err != nil {
				return domain.BackupManifest{}, nil, fmt.Errorf("manifest: %w", err)
			}
		case fullBackupDumpName:
			dump = body
		}
	}
	if len(dump) == 0 {
		return manifest, nil, fmt.Errorf("archive missing %s", fullBackupDumpName)
	}
	return manifest, dump, nil
}

func writeTarFile(tw *tar.Writer, name string, data []byte) error {
	hdr := &tar.Header{
		Name:    name,
		Mode:    0o644,
		Size:    int64(len(data)),
		ModTime: time.Now().UTC(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}
