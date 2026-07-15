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
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vortexui/vortexui/internal/domain"
)

// pgClientVersion runs "pg_dump --version" and extracts the major version.
func pgClientVersion(bin string) (int, error) {
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		return 0, fmt.Errorf("cannot determine %s version: %w", bin, err)
	}
	parts := strings.Fields(string(out))
	for _, p := range parts {
		if v, err := strconv.Atoi(strings.Split(p, ".")[0]); err == nil && v > 0 {
			return v, nil
		}
	}
	return 0, fmt.Errorf("cannot parse %s version from: %s", bin, strings.TrimSpace(string(out)))
}

const (
	fullBackupManifestName = "manifest.json"
	fullBackupDumpName     = "database.dump"
)

// serverMajorVersion queries the connected server's major version (e.g. 16).
// Used to pick a pg_dump/pg_restore binary that matches the server even when
// the distro's default package (found first on PATH) is an older or newer
// major version — pg_dump refuses to talk to a server newer than itself.
func serverMajorVersion(ctx context.Context, databaseURL string) (int, error) {
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return 0, err
	}
	defer func() { _ = conn.Close(ctx) }()
	var raw string
	if err := conn.QueryRow(ctx, "SHOW server_version_num").Scan(&raw); err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, err
	}
	return n / 10000, nil
}

// resolvePgBinary finds a pg_dump/pg_restore binary matching the server's
// major version. Debian/Ubuntu (PGDG) and RHEL-family (PGDG) both install
// version-specific client binaries under well-known paths that are not
// necessarily first on PATH, so check those before falling back to PATH.
func resolvePgBinary(ctx context.Context, databaseURL, tool string) (string, error) {
	if major, err := serverMajorVersion(ctx, databaseURL); err == nil {
		for _, c := range []string{
			fmt.Sprintf("/usr/lib/postgresql/%d/bin/%s", major, tool), // Debian/Ubuntu
			fmt.Sprintf("/usr/pgsql-%d/bin/%s", major, tool),          // RHEL/CentOS
		} {
			if st, statErr := os.Stat(c); statErr == nil && !st.IsDir() {
				return c, nil
			}
		}
	}
	return exec.LookPath(tool)
}

// DumpDatabase runs pg_dump in custom format (-Fc) against databaseURL.
func DumpDatabase(ctx context.Context, databaseURL string) ([]byte, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	pgDump, err := resolvePgBinary(ctx, databaseURL, "pg_dump")
	if err != nil {
		return nil, fmt.Errorf("pg_dump not found: install postgresql-client-16 (sudo apt-get install postgresql-client-16). On restricted networks (Iran/China), use a proxy or download manually from https://ftp.postgresql.org/pub/source/")
	}
	// Verify pg_dump version matches server version before running.
	// resolvePgBinary above already checks versioned paths, but it falls
	// back to whatever is on PATH — which may be the wrong major version.
	// Catch that here and try harder before giving a clear error.
	serverMajor, srvErr := serverMajorVersion(ctx, databaseURL)
	clientMajor, clientErr := pgClientVersion(pgDump)
	if srvErr == nil && clientErr == nil && serverMajor != clientMajor {
		// Try versioned paths one more time (in case resolvePgBinary's
		// version query had a transient issue or PATH was later updated).
		for _, candidate := range []string{
			fmt.Sprintf("/usr/lib/postgresql/%d/bin/pg_dump", serverMajor),
			fmt.Sprintf("/usr/pgsql-%d/bin/pg_dump", serverMajor),
		} {
			if st, statErr := os.Stat(candidate); statErr == nil && !st.IsDir() {
				pgDump = candidate
				clientMajor, clientErr = pgClientVersion(pgDump)
				break
			}
		}
	}
	// If the version still doesn't match, clear instructions for the user.
	if srvErr == nil && clientErr == nil && serverMajor != clientMajor {
		return nil, fmt.Errorf(
			"pg_dump version mismatch: server is PostgreSQL %d but pg_dump is version %d (found at %s). "+
				"Install the matching client: sudo apt-get install postgresql-client-%d. "+
				"For restricted networks, try a proxy or download from https://ftp.postgresql.org/pub/source/",
			serverMajor, clientMajor, pgDump, serverMajor,
		)
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
	pgRestore, err := resolvePgBinary(ctx, databaseURL, "pg_restore")
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
