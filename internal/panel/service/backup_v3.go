package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// BackupExportOptions controls JSON v3 export scope.
type BackupExportOptions struct {
	IncludeTrafficMetrics bool
	IncludeCredentials    bool
	IncludeSupplemental   bool
}

// backupStore extends restore with supplemental export and full DB restore.
type backupStore interface {
	BackupRestorer
	ExportSupplemental(ctx context.Context, includeTraffic bool) (*domain.BackupSupplemental, error)
	ExportAdminCredentials(ctx context.Context) ([]domain.BackupAdminCredential, error)
	RestoreFull(ctx context.Context, databaseURL string, dump []byte) error
	RestoreReseller(ctx context.Context, rb *domain.ResellerBackup) error
	ListOrdersForAdmin(ctx context.Context, adminID uuid.UUID) ([]domain.Order, error)
	ExportFullArchive(ctx context.Context, databaseURL string, manifest domain.BackupManifest) ([]byte, error)
	UnpackFullArchive(data []byte) (domain.BackupManifest, []byte, error)
}

// backupResellerSource loads reseller-scoped data for export.
type backupResellerSource interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error)
	ListWalletLedger(ctx context.Context, adminID uuid.UUID, limit int) ([]domain.WalletLedgerEntry, error)
	GetPortalBranding(ctx context.Context, adminID uuid.UUID) (*domain.PortalBranding, error)
}

type backupPaymentSource interface {
	Get(ctx context.Context, adminID uuid.UUID) (*domain.ResellerPaymentConfig, error)
}

// SetDatabaseURL enables full pg_dump/pg_restore backups.
func (s *BackupService) SetDatabaseURL(url string) { s.databaseURL = url }

// SetResellerSources wires reseller export dependencies.
func (s *BackupService) SetResellerSources(admins backupResellerSource, pay backupPaymentSource) {
	s.resellerAdmins = admins
	s.resellerPay = pay
}

// Manifest previews what a JSON export would contain without downloading it.
func (s *BackupService) Manifest(ctx context.Context) (*domain.BackupManifest, error) {
	b, err := s.export(ctx, BackupExportOptions{
		IncludeCredentials: false, IncludeSupplemental: true,
	})
	if err != nil {
		return nil, err
	}
	m := domain.BuildBackupManifest(b, domain.BackupFormatJSON)
	m.ExcludedTables = append([]string{}, domain.ExcludedFromJSONBackup...)
	return &m, nil
}

// ExportV3 assembles a complete v3 snapshot with optional supplemental data.
func (s *BackupService) ExportV3(ctx context.Context, opts BackupExportOptions) (*domain.Backup, error) {
	return s.export(ctx, opts)
}

func (s *BackupService) export(ctx context.Context, opts BackupExportOptions) (*domain.Backup, error) {
	b, err := s.Export(ctx)
	if err != nil {
		return nil, err
	}
	b.Version = domain.BackupVersion

	store, ok := s.restorer.(backupStore)
	if opts.IncludeSupplemental && ok {
		sup, err := store.ExportSupplemental(ctx, opts.IncludeTrafficMetrics)
		if err != nil {
			return nil, fmt.Errorf("export supplemental: %w", err)
		}
		b.Supplemental = sup
	}
	if opts.IncludeCredentials && ok {
		creds, err := store.ExportAdminCredentials(ctx)
		if err != nil {
			return nil, fmt.Errorf("export admin credentials: %w", err)
		}
		b.AdminCredentials = creds
	}
	manifest := domain.BuildBackupManifest(b, domain.BackupFormatJSON)
	b.Manifest = &manifest
	return b, nil
}

// ExportFull creates a gzip tar with pg_dump + manifest for server migration.
func (s *BackupService) ExportFull(ctx context.Context) ([]byte, *domain.BackupManifest, error) {
	if s.databaseURL == "" {
		return nil, nil, errors.New("full backup requires database URL")
	}
	store, ok := s.restorer.(backupStore)
	if !ok {
		return nil, nil, errors.New("full backup not available")
	}
	manifest, err := s.Manifest(ctx)
	if err != nil {
		return nil, nil, err
	}
	archive, err := store.ExportFullArchive(ctx, s.databaseURL, *manifest)
	if err != nil {
		return nil, nil, err
	}
	return archive, manifest, nil
}

// RestoreWithMode applies a JSON config or full database archive restore.
func (s *BackupService) RestoreWithMode(ctx context.Context, mode domain.BackupRestoreMode, b *domain.Backup, fullArchive []byte) (*domain.BackupRestoreReport, error) {
	switch mode {
	case domain.BackupRestoreFull:
		store, ok := s.restorer.(backupStore)
		if !ok {
			return nil, errors.New("full restore not available")
		}
		if s.databaseURL == "" {
			return nil, errors.New("full restore requires database URL")
		}
		if len(fullArchive) == 0 {
			return nil, errors.New("empty full backup archive")
		}
		manifest, dump, err := store.UnpackFullArchive(fullArchive)
		if err != nil {
			return nil, err
		}
		if err := store.RestoreFull(ctx, s.databaseURL, dump); err != nil {
			return nil, err
		}
		return &domain.BackupRestoreReport{
			Mode:     domain.BackupRestoreFull,
			Restored: manifest.Counts,
			Warnings: manifest.Warnings,
		}, nil
	default:
		if b == nil {
			return nil, errors.New("empty backup")
		}
		if !domain.IsSupportedBackupVersion(b.Version) {
			return nil, fmt.Errorf("unsupported backup version %d", b.Version)
		}
		if err := s.restorer.Restore(ctx, b); err != nil {
			return nil, err
		}
		counts := domain.BackupCounts{
			Roles: len(b.Roles), Admins: len(b.Admins), Plans: len(b.Plans),
			Nodes: len(b.Nodes), Inbounds: len(b.Inbounds), Outbounds: len(b.Outbounds),
			Routing: len(b.Routing), Balancers: len(b.Balancers),
			Users: len(b.Users), Bindings: len(b.Bindings),
		}
		if b.Supplemental != nil && b.Supplemental.Tables != nil {
			counts.Orders = len(b.Supplemental.Tables["orders"])
			counts.WalletLedger = len(b.Supplemental.Tables["admin_wallet_ledger"])
			counts.WalletDeposits = len(b.Supplemental.Tables["wallet_deposits"])
			counts.WalletPackages = len(b.Supplemental.Tables["wallet_packages"])
		}
		var warnings []string
		if b.Manifest != nil {
			warnings = b.Manifest.Warnings
		}
		return &domain.BackupRestoreReport{
			Mode: domain.BackupRestoreConfig, Restored: counts, Warnings: warnings,
		}, nil
	}
}

// ExportReseller builds a scoped backup for one reseller account.
func (s *BackupService) ExportReseller(ctx context.Context, adminID uuid.UUID, listUsers func(context.Context, uuid.UUID) ([]*domain.User, []domain.UserProxy, error)) (*domain.ResellerBackup, error) {
	if listUsers == nil {
		return nil, errors.New("user lister required")
	}
	users, bindings, err := listUsers(ctx, adminID)
	if err != nil {
		return nil, err
	}
	rb := &domain.ResellerBackup{
		Version:    domain.ResellerBackupVersion,
		ExportedAt: s.now(),
		AdminID:    adminID,
		Users:      users,
		Bindings:   bindings,
	}
	if s.resellerAdmins != nil {
		if admin, err := s.resellerAdmins.GetByID(ctx, adminID); err == nil && admin != nil {
			rb.AdminUsername = admin.Username
		}
		ledger, err := s.resellerAdmins.ListWalletLedger(ctx, adminID, 100_000)
		if err != nil {
			return nil, fmt.Errorf("wallet ledger: %w", err)
		}
		rb.WalletLedger = ledger
		if branding, err := s.resellerAdmins.GetPortalBranding(ctx, adminID); err == nil {
			rb.PortalBranding = branding
		}
	}
	if s.resellerPay != nil {
		if cfg, err := s.resellerPay.Get(ctx, adminID); err == nil {
			rb.PaymentConfig = cfg
		}
	}
	if store, ok := s.restorer.(backupStore); ok {
		orders, err := store.ListOrdersForAdmin(ctx, adminID)
		if err != nil {
			return nil, fmt.Errorf("orders: %w", err)
		}
		rb.Orders = orders
	}
	usage := domain.BuildUsageSummary(users, nil)
	rb.Manifest = domain.ResellerBackupManifest{
		UserCount:             len(users),
		TotalUsedTraffic:      usage.TotalUsedTraffic,
		TotalDataLimit:        usage.TotalDataLimit,
		TotalRemainingTraffic: usage.TotalRemainingTraffic,
		OrdersCount:           len(rb.Orders),
		LedgerEntries:         len(rb.WalletLedger),
	}
	return rb, nil
}

// RestoreReseller applies a scoped reseller backup.
func (s *BackupService) RestoreReseller(ctx context.Context, rb *domain.ResellerBackup) (*domain.BackupRestoreReport, error) {
	if rb == nil {
		return nil, errors.New("empty reseller backup")
	}
	if rb.Version != domain.ResellerBackupVersion && rb.Version != 1 {
		return nil, fmt.Errorf("unsupported reseller backup version %d", rb.Version)
	}
	store, ok := s.restorer.(backupStore)
	if !ok {
		return nil, errors.New("reseller restore not available")
	}
	if err := store.RestoreReseller(ctx, rb); err != nil {
		return nil, err
	}
	return &domain.BackupRestoreReport{
		Mode: domain.BackupRestoreConfig,
		Restored: domain.BackupCounts{
			Users: len(rb.Users), Bindings: len(rb.Bindings),
			Orders: len(rb.Orders), WalletLedger: len(rb.WalletLedger),
		},
	}, nil
}

// EncryptExport optionally encrypts exported JSON bytes.
func EncryptExport(data []byte, passphrase string) ([]byte, error) {
	return EncryptBackup(data, passphrase)
}

// DecryptImport optionally decrypts imported backup bytes.
func DecryptImport(data []byte, passphrase string) ([]byte, error) {
	return DecryptBackup(data, passphrase)
}

// ParseBackupJSON unmarshals a backup document after optional decryption.
func ParseBackupJSON(raw []byte, passphrase string) (*domain.Backup, error) {
	decrypted, err := DecryptImport(raw, passphrase)
	if err != nil {
		return nil, err
	}
	var b domain.Backup
	if err := json.Unmarshal(decrypted, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// ParseResellerBackupJSON unmarshals a reseller backup document.
func ParseResellerBackupJSON(raw []byte) (*domain.ResellerBackup, error) {
	var rb domain.ResellerBackup
	if err := json.Unmarshal(raw, &rb); err != nil {
		return nil, err
	}
	return &rb, nil
}

// LegacyResellerBackup converts v1 reseller export (users only) to v2 shape.
func LegacyResellerBackup(raw json.RawMessage) (*domain.ResellerBackup, error) {
	var legacy struct {
		Version    int            `json:"version"`
		ExportedAt time.Time      `json:"exported_at"`
		AdminID    uuid.UUID      `json:"admin_id"`
		Users      []*domain.User `json:"users"`
	}
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return nil, err
	}
	if legacy.Version != 1 || legacy.AdminID == uuid.Nil {
		return nil, errors.New("not a legacy reseller backup")
	}
	return &domain.ResellerBackup{
		Version: domain.ResellerBackupVersion, ExportedAt: legacy.ExportedAt,
		AdminID: legacy.AdminID, Users: legacy.Users,
		Manifest: domain.ResellerBackupManifest{UserCount: len(legacy.Users)},
	}, nil
}
