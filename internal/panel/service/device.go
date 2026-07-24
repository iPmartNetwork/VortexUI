package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// Device-level sentinel errors.
var (
	ErrDeviceLocked        = errors.New("device not registered and lock is enabled")
	ErrDeviceLimitExceeded = errors.New("device limit exceeded")
)

// DeviceService manages HWID-based device registrations and access control.
type DeviceService struct {
	repo  port.DeviceRepository
	users port.UserRepository
}

// NewDeviceService wires the device service.
func NewDeviceService(repo port.DeviceRepository, users port.UserRepository) *DeviceService {
	return &DeviceService{repo: repo, users: users}
}

// RegisterDevice upserts a device entry for the given user. If the HWID already
// exists, it updates the OS and last_seen timestamp.
func (s *DeviceService) RegisterDevice(ctx context.Context, userID uuid.UUID, hwid, os string) error {
	d := &domain.Device{
		ID:     uuid.New(),
		UserID: userID,
		HWID:   hwid,
		OS:     os,
	}
	return s.repo.Upsert(ctx, d)
}

// CheckDeviceAllowed enforces device_lock and device_limit policies:
//  1. If device_lock is true and the HWID is not already registered → ErrDeviceLocked
//  2. If device_limit > 0, current count >= limit, and HWID not registered → ErrDeviceLimitExceeded
//  3. Otherwise the device is allowed.
func (s *DeviceService) CheckDeviceAllowed(ctx context.Context, userID uuid.UUID, hwid string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if this HWID is already registered for the user.
	exists, err := s.repo.Exists(ctx, userID, hwid)
	if err != nil {
		return err
	}

	// Rule 1: device lock enforcement.
	if user.DeviceLock && !exists {
		return ErrDeviceLocked
	}

	// Rule 2: device limit enforcement.
	if user.DeviceLimit > 0 && !exists {
		count, err := s.repo.CountByUser(ctx, userID)
		if err != nil {
			return err
		}
		if count >= user.DeviceLimit {
			return ErrDeviceLimitExceeded
		}
	}

	return nil
}

// RevokeDevice removes a specific device registration for a user.
func (s *DeviceService) RevokeDevice(ctx context.Context, userID uuid.UUID, hwid string) error {
	return s.repo.Delete(ctx, userID, hwid)
}

// BulkResetHWIDs removes all device registrations for the given users.
func (s *DeviceService) BulkResetHWIDs(ctx context.Context, userIDs []uuid.UUID) error {
	return s.repo.DeleteAllForUsers(ctx, userIDs)
}

// ListDevices returns all registered devices for the given user.
func (s *DeviceService) ListDevices(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	return s.repo.ListByUser(ctx, userID)
}
