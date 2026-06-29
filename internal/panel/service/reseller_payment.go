package service

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// ResellerPaymentRepository is the persistence port for per-reseller payment config.
type ResellerPaymentRepository interface {
	Get(ctx context.Context, adminID uuid.UUID) (*domain.ResellerPaymentConfig, error)
	Upsert(ctx context.Context, cfg *domain.ResellerPaymentConfig) error
}

// ResellerPaymentService handles per-reseller payment configuration CRUD and validation.
type ResellerPaymentService struct {
	repo ResellerPaymentRepository
}

// NewResellerPaymentService wires the service.
func NewResellerPaymentService(repo ResellerPaymentRepository) *ResellerPaymentService {
	return &ResellerPaymentService{repo: repo}
}

// GetPaymentConfig loads the reseller's payment config, returning empty defaults when none exists.
func (s *ResellerPaymentService) GetPaymentConfig(ctx context.Context, adminID uuid.UUID) (*domain.ResellerPaymentConfig, error) {
	cfg, err := s.repo.Get(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return &domain.ResellerPaymentConfig{
			AdminID:         adminID,
			CryptoAddresses: make(map[string]string),
			EnabledMethods:  []string{},
		}, nil
	}
	if cfg.CryptoAddresses == nil {
		cfg.CryptoAddresses = make(map[string]string)
	}
	if cfg.EnabledMethods == nil {
		cfg.EnabledMethods = []string{}
	}
	return cfg, nil
}

// cardNumberRe is a loose check: 16-19 digits, optionally separated by dashes/spaces.
var cardNumberRe = regexp.MustCompile(`^[\d\- ]{0,25}$`)

var (
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrInvalidCardNumber    = errors.New("card number format invalid")
)

// SavePaymentConfig validates and persists the reseller's payment config.
func (s *ResellerPaymentService) SavePaymentConfig(ctx context.Context, adminID uuid.UUID, cfg *domain.ResellerPaymentConfig) error {
	// Validate enabled_methods
	for _, m := range cfg.EnabledMethods {
		if !domain.KnownPaymentMethods[m] {
			return ErrInvalidPaymentMethod
		}
	}
	// Card number sanity (allow empty)
	if cfg.CardNumber != "" && !cardNumberRe.MatchString(cfg.CardNumber) {
		return ErrInvalidCardNumber
	}
	cfg.AdminID = adminID
	if cfg.CryptoAddresses == nil {
		cfg.CryptoAddresses = make(map[string]string)
	}
	if cfg.EnabledMethods == nil {
		cfg.EnabledMethods = []string{}
	}
	return s.repo.Upsert(ctx, cfg)
}
