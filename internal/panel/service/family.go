package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// FamilyService manages family/group subscription logic.
type FamilyService struct {
	repo  port.FamilyRepository
	users port.UserRepository
	now   func() time.Time
}

// NewFamilyService wires dependencies.
func NewFamilyService(repo port.FamilyRepository, users port.UserRepository) *FamilyService {
	return &FamilyService{repo: repo, users: users, now: time.Now}
}

// CreateGroupInput describes a new family group.
type CreateGroupInput struct {
	Name        string
	OwnerID     uuid.UUID
	DataLimit   int64
	MaxMembers  int
	MemberQuota int64
}

// CreateGroup creates a new family/group subscription.
func (s *FamilyService) CreateGroup(ctx context.Context, in CreateGroupInput) (*domain.FamilyGroup, error) {
	if in.Name == "" {
		return nil, errors.New("group name is required")
	}
	if in.MaxMembers <= 0 {
		in.MaxMembers = 5
	}
	// Verify owner exists.
	owner, err := s.users.GetByID(ctx, in.OwnerID)
	if err != nil {
		return nil, errors.New("owner user not found")
	}
	g := &domain.FamilyGroup{
		ID:          uuid.New(),
		Name:        in.Name,
		OwnerID:     in.OwnerID,
		OwnerName:   owner.Username,
		DataLimit:   in.DataLimit,
		MaxMembers:  in.MaxMembers,
		MemberQuota: in.MemberQuota,
		CreatedAt:   s.now(),
	}
	if err := s.repo.CreateGroup(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

// GetGroup returns a group with its members.
func (s *FamilyService) GetGroup(ctx context.Context, id uuid.UUID) (*domain.FamilyGroup, error) {
	g, err := s.repo.GetGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.ListMembers(ctx, id)
	if err != nil {
		return nil, err
	}
	g.Members = members
	return g, nil
}

// ListGroups returns all family groups.
func (s *FamilyService) ListGroups(ctx context.Context, limit, offset int) ([]*domain.FamilyGroup, int, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListGroups(ctx, limit, offset)
}

// AddMember adds a user to a group.
func (s *FamilyService) AddMember(ctx context.Context, groupID, userID uuid.UUID, label string) (*domain.FamilyMember, error) {
	g, err := s.repo.GetGroup(ctx, groupID)
	if err != nil {
		return nil, errors.New("group not found")
	}
	members, err := s.repo.ListMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if len(members) >= g.MaxMembers {
		return nil, errors.New("group is full (max members reached)")
	}
	// Check user exists.
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	m := &domain.FamilyMember{
		ID:       uuid.New(),
		GroupID:  groupID,
		UserID:   userID,
		Username: user.Username,
		Label:    label,
		JoinedAt: s.now(),
	}
	if err := s.repo.AddMember(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// RemoveMember removes a user from a group.
func (s *FamilyService) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	return s.repo.RemoveMember(ctx, groupID, userID)
}

// DeleteGroup removes a family group.
func (s *FamilyService) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteGroup(ctx, id)
}
