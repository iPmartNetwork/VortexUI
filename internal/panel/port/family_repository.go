package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// FamilyRepository persists family/group subscription data.
type FamilyRepository interface {
	CreateGroup(ctx context.Context, g *domain.FamilyGroup) error
	GetGroup(ctx context.Context, id uuid.UUID) (*domain.FamilyGroup, error)
	GetGroupByOwner(ctx context.Context, ownerID uuid.UUID) (*domain.FamilyGroup, error)
	UpdateGroup(ctx context.Context, g *domain.FamilyGroup) error
	DeleteGroup(ctx context.Context, id uuid.UUID) error
	ListGroups(ctx context.Context, limit, offset int) ([]*domain.FamilyGroup, int, error)

	AddMember(ctx context.Context, m *domain.FamilyMember) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]domain.FamilyMember, error)
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.FamilyMember, error)
	UpdateMemberTraffic(ctx context.Context, groupID, userID uuid.UUID, delta int64) error
}
