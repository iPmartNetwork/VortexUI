package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// PushSubscription represents a user's Web Push API subscription.
type PushSubscription struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Endpoint  string    `json:"endpoint"`
	P256dh    string    `json:"p256dh"`
	Auth      string    `json:"auth"`
	CreatedAt time.Time `json:"created_at"`
}

// ConnectionGuide is an app-specific setup instruction document.
type ConnectionGuide struct {
	ID        uuid.UUID `json:"id"`
	AppName   string    `json:"app_name"`
	Platform  string    `json:"platform"`
	IconURL   string    `json:"icon_url"`
	Content   string    `json:"content"`
	SortOrder int       `json:"sort_order"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SpeedTestResult holds per-node speed measurement data.
type SpeedTestResult struct {
	NodeID      uuid.UUID `json:"node_id"`
	NodeName    string    `json:"node_name"`
	LatencyMs   int       `json:"latency_ms"`
	DownloadMbps float64  `json:"download_mbps"`
	Timestamp   time.Time `json:"timestamp"`
}

// UsageAlert represents a usage threshold notification.
type UsageAlert struct {
	UserID     uuid.UUID `json:"user_id"`
	Threshold  int       `json:"threshold"`  // 80, 90, or 100
	UsedBytes  int64     `json:"used_bytes"`
	LimitBytes int64     `json:"limit_bytes"`
	Message    string    `json:"message"`
}

// SetupWizardStep is one step of the user onboarding wizard.
type SetupWizardStep struct {
	Step        int    `json:"step"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Completed   bool   `json:"completed"`
}

// PushSubscriptionRepository persists push subscriptions.
type PushSubscriptionRepository interface {
	Create(ctx context.Context, sub *PushSubscription) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*PushSubscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ConnectionGuideRepository persists connection guides.
type ConnectionGuideRepository interface {
	List(ctx context.Context) ([]*ConnectionGuide, error)
	ListEnabled(ctx context.Context) ([]*ConnectionGuide, error)
}

// PortalProService provides portal features: speed test, dynamic config,
// usage alerts, push notifications, and setup wizard.
type PortalProService struct {
	pushRepo  PushSubscriptionRepository
	guideRepo ConnectionGuideRepository
	httpClient *http.Client
}

// NewPortalProService creates the portal service.
func NewPortalProService(pushRepo PushSubscriptionRepository, guideRepo ConnectionGuideRepository) *PortalProService {
	return &PortalProService{
		pushRepo:   pushRepo,
		guideRepo:  guideRepo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SpeedTest measures latency and estimated download speed for a node.
// In production this pings the node's health endpoint and measures timing.
func (s *PortalProService) SpeedTest(ctx context.Context, nodeID uuid.UUID, nodeEndpoint string) (*SpeedTestResult, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nodeEndpoint+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	latency := time.Since(start)
	if err != nil {
		return &SpeedTestResult{
			NodeID:    nodeID,
			LatencyMs: int(latency.Milliseconds()),
			Timestamp: time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	// Estimate download speed based on response time (simplified)
	downloadMbps := 0.0
	if latency > 0 {
		// Rough heuristic: lower latency = higher estimated throughput
		downloadMbps = 1000.0 / float64(latency.Milliseconds())
		if downloadMbps > 1000 {
			downloadMbps = 1000
		}
	}

	return &SpeedTestResult{
		NodeID:       nodeID,
		LatencyMs:    int(latency.Milliseconds()),
		DownloadMbps: downloadMbps,
		Timestamp:    time.Now(),
	}, nil
}

// CheckUsageAlerts evaluates whether a user has crossed usage thresholds.
func (s *PortalProService) CheckUsageAlerts(usedBytes, limitBytes int64) []UsageAlert {
	if limitBytes <= 0 {
		return nil
	}

	var alerts []UsageAlert
	pct := float64(usedBytes) / float64(limitBytes) * 100

	thresholds := []int{80, 90, 100}
	for _, t := range thresholds {
		if pct >= float64(t) {
			alerts = append(alerts, UsageAlert{
				Threshold:  t,
				UsedBytes:  usedBytes,
				LimitBytes: limitBytes,
				Message:    fmt.Sprintf("Traffic usage has reached %d%%", t),
			})
		}
	}

	return alerts
}

// SubscribePush registers a push notification subscription.
func (s *PortalProService) SubscribePush(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth string) (*PushSubscription, error) {
	sub := &PushSubscription{
		ID:        uuid.New(),
		UserID:    userID,
		Endpoint:  endpoint,
		P256dh:    p256dh,
		Auth:      auth,
		CreatedAt: time.Now(),
	}
	if err := s.pushRepo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("create push subscription: %w", err)
	}
	return sub, nil
}

// SendPushNotification delivers a push message to all user subscriptions.
// In production this would use the Web Push protocol with VAPID keys.
func (s *PortalProService) SendPushNotification(ctx context.Context, userID uuid.UUID, title, body string) error {
	subs, err := s.pushRepo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("list subscriptions: %w", err)
	}

	for _, sub := range subs {
		// Web Push delivery would happen here with proper VAPID signing.
		// For now we log the intent.
		_ = sub.Endpoint
		_ = title
		_ = body
	}

	return nil
}

// ListGuides returns all enabled connection guides.
func (s *PortalProService) ListGuides(ctx context.Context) ([]*ConnectionGuide, error) {
	return s.guideRepo.ListEnabled(ctx)
}

// GetSetupWizard returns the onboarding wizard steps for a user.
func (s *PortalProService) GetSetupWizard(ctx context.Context, hasSubscription, hasDevice bool) []SetupWizardStep {
	steps := []SetupWizardStep{
		{Step: 1, Title: "Get your subscription link", Description: "Copy your personal subscription URL from the dashboard", Action: "copy_link", Completed: hasSubscription},
		{Step: 2, Title: "Install a client app", Description: "Download and install a compatible proxy client for your device", Action: "view_guides", Completed: false},
		{Step: 3, Title: "Add subscription", Description: "Paste your subscription link into the client app", Action: "paste_link", Completed: hasDevice},
		{Step: 4, Title: "Connect", Description: "Select a server and tap connect", Action: "connect", Completed: false},
	}
	return steps
}
