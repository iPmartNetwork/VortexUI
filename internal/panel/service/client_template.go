package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ClientTemplateService manages client templates and subscription approval workflows.
type ClientTemplateService struct {
	templates port.ClientTemplateRepository
	approvals port.ApprovalQueueRepository
}

// NewClientTemplateService constructs a ClientTemplateService with the required dependencies.
func NewClientTemplateService(
	templates port.ClientTemplateRepository,
	approvals port.ApprovalQueueRepository,
) *ClientTemplateService {
	return &ClientTemplateService{
		templates: templates,
		approvals: approvals,
	}
}

// MatchClient evaluates the given User-Agent against all enabled client templates
// ordered by priority (descending). Returns the highest-priority match, or nil
// if no template matches.
func (s *ClientTemplateService) MatchClient(ctx context.Context, userAgent string) (*domain.ClientTemplate, error) {
	templates, err := s.templates.ListEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled templates: %w", err)
	}

	for _, t := range templates {
		re, err := regexp.Compile("(?i)" + t.ClientPattern)
		if err != nil {
			// Skip templates with invalid regex patterns.
			continue
		}
		if re.MatchString(userAgent) {
			return t, nil
		}
	}
	return nil, nil
}

// SubmitApproval queues a subscription approval request for admin review.
func (s *ClientTemplateService) SubmitApproval(ctx context.Context, userID uuid.UUID, data map[string]any) error {
	a := &domain.SubscriptionApproval{
		ID:          uuid.New(),
		UserID:      userID,
		RequestData: data,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	return s.approvals.Create(ctx, a)
}

// Approve resolves an approval request as approved by the given admin.
func (s *ClientTemplateService) Approve(ctx context.Context, approvalID, adminID uuid.UUID) error {
	return s.approvals.Approve(ctx, approvalID, adminID)
}

// Reject resolves an approval request as rejected.
func (s *ClientTemplateService) Reject(ctx context.Context, approvalID uuid.UUID) error {
	return s.approvals.Reject(ctx, approvalID)
}

// PreviewTemplate renders a template string with sample user data for live
// preview. Supports basic variable substitution for demonstration purposes.
func (s *ClientTemplateService) PreviewTemplate(template string, sampleUserID uuid.UUID) (string, error) {
	// Provide sample data for preview rendering.
	replacements := map[string]string{
		"{USER_ID}":   sampleUserID.String(),
		"{USERNAME}":  "sample_user",
		"{EXPIRE_AT}": time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
		"{DATA_LIMIT}": "100GB",
		"{PROTOCOL}":  "vless",
		"{TRANSPORT}": "grpc",
		"{NODE_NAME}": "US-Node-01",
	}

	result := template
	for k, v := range replacements {
		result = strings.ReplaceAll(result, k, v)
	}
	return result, nil
}

// --- CRUD pass-through methods ---

// CreateTemplate creates a new client template.
func (s *ClientTemplateService) CreateTemplate(ctx context.Context, t *domain.ClientTemplate) error {
	return s.templates.Create(ctx, t)
}

// GetTemplate retrieves a client template by ID.
func (s *ClientTemplateService) GetTemplate(ctx context.Context, id uuid.UUID) (*domain.ClientTemplate, error) {
	return s.templates.GetByID(ctx, id)
}

// UpdateTemplate updates an existing client template.
func (s *ClientTemplateService) UpdateTemplate(ctx context.Context, t *domain.ClientTemplate) error {
	return s.templates.Update(ctx, t)
}

// DeleteTemplate removes a client template by ID.
func (s *ClientTemplateService) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return s.templates.Delete(ctx, id)
}

// ListTemplates returns all client templates ordered by priority.
func (s *ClientTemplateService) ListTemplates(ctx context.Context) ([]*domain.ClientTemplate, error) {
	return s.templates.List(ctx)
}

// ListPendingApprovals returns all pending approval requests.
func (s *ClientTemplateService) ListPendingApprovals(ctx context.Context) ([]*domain.SubscriptionApproval, error) {
	return s.approvals.ListPending(ctx)
}

// GetApproval retrieves an approval request by ID.
func (s *ClientTemplateService) GetApproval(ctx context.Context, id uuid.UUID) (*domain.SubscriptionApproval, error) {
	return s.approvals.GetByID(ctx, id)
}
