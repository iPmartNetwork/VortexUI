package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ReferralService manages the invite/referral system.
type ReferralService struct {
	repo  port.ReferralRepository
	users port.UserRepository
	now   func() time.Time
}

// NewReferralService wires the referral service.
func NewReferralService(repo port.ReferralRepository, users port.UserRepository) *ReferralService {
	return &ReferralService{repo: repo, users: users, now: time.Now}
}

// GetConfig returns the referral program configuration.
func (s *ReferralService) GetConfig(ctx context.Context) (*domain.ReferralConfig, error) {
	c, err := s.repo.GetConfig(ctx)
	if err != nil {
		def := domain.DefaultReferralConfig()
		return &def, nil
	}
	return c, nil
}

// UpdateConfig saves the referral config.
func (s *ReferralService) UpdateConfig(ctx context.Context, c *domain.ReferralConfig) error {
	return s.repo.SaveConfig(ctx, c)
}

// GetOrCreateCode returns the user's referral code, creating one if it doesn't exist.
func (s *ReferralService) GetOrCreateCode(ctx context.Context, userID uuid.UUID) (*domain.ReferralCode, error) {
	existing, err := s.repo.GetCodeByUser(ctx, userID)
	if err == nil {
		return existing, nil
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	code := generateCode()
	rc := &domain.ReferralCode{
		ID:        uuid.New(),
		UserID:    userID,
		Username:  user.Username,
		Code:      code,
		Uses:      0,
		MaxUses:   0,
		CreatedAt: s.now(),
	}
	if err := s.repo.CreateCode(ctx, rc); err != nil {
		return nil, err
	}
	return rc, nil
}

// ApplyReferral processes a referral: the referred user signed up using a code.
func (s *ReferralService) ApplyReferral(ctx context.Context, code string, referredUserID uuid.UUID) error {
	config, _ := s.GetConfig(ctx)
	if !config.Enabled {
		return errors.New("referral program is disabled")
	}
	rc, err := s.repo.GetCodeByCode(ctx, code)
	if err != nil {
		return errors.New("invalid referral code")
	}
	if rc.MaxUses > 0 && rc.Uses >= rc.MaxUses {
		return errors.New("referral code has reached max uses")
	}
	if rc.UserID == referredUserID {
		return errors.New("cannot refer yourself")
	}

	referred, err := s.users.GetByID(ctx, referredUserID)
	if err != nil {
		return errors.New("referred user not found")
	}

	// Record the event.
	event := &domain.ReferralEvent{
		ID:           uuid.New(),
		ReferrerID:   rc.UserID,
		ReferrerName: rc.Username,
		ReferredID:   referredUserID,
		ReferredName: referred.Username,
		CodeUsed:     code,
		RewardType:   config.RewardType,
		RewardAmount: config.RewardAmount,
		CreatedAt:    s.now(),
	}

	// Apply reward to referrer.
	if !config.RequirePaid {
		if err := s.applyReward(ctx, rc.UserID, config); err == nil {
			event.RewardApplied = true
		}
	}

	_ = s.repo.SaveEvent(ctx, event)
	_ = s.repo.IncrementUses(ctx, rc.ID)

	return nil
}

func (s *ReferralService) applyReward(ctx context.Context, userID uuid.UUID, config *domain.ReferralConfig) error {
	switch config.RewardType {
	case domain.RewardData:
		// Add data to user's limit.
		user, err := s.users.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		user.DataLimit += config.RewardAmount
		return s.users.Update(ctx, user)
	case domain.RewardDays:
		user, err := s.users.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		extension := time.Duration(config.RewardAmount) * 24 * time.Hour
		if user.ExpireAt != nil {
			newExpiry := user.ExpireAt.Add(extension)
			user.ExpireAt = &newExpiry
		} else {
			newExpiry := s.now().Add(extension)
			user.ExpireAt = &newExpiry
		}
		return s.users.Update(ctx, user)
	default:
		return nil
	}
}

// ListEvents returns referral events for a user (or all if nil).
func (s *ReferralService) ListEvents(ctx context.Context, userID *uuid.UUID) ([]*domain.ReferralEvent, error) {
	return s.repo.ListEvents(ctx, userID, 100)
}

// ListCodes returns all referral codes (admin view).
func (s *ReferralService) ListCodes(ctx context.Context, limit, offset int) ([]*domain.ReferralCode, int, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListCodes(ctx, limit, offset)
}

func generateCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "VX-" + hex.EncodeToString(b)
}
