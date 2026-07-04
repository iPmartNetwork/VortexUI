package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// NodeRegistrar is the slice of the hub the node service drives: bringing nodes
// under (or out of) live management. *hub.Hub satisfies it.
type NodeRegistrar interface {
	Register(ctx context.Context, node *domain.Node) error
	Deregister(nodeID uuid.UUID)
}

// NodeLogQuerier fetches recent core log lines from a node. *hub.Hub satisfies it.
type NodeLogQuerier interface {
	Logs(ctx context.Context, nodeID uuid.UUID, limit int) ([]string, error)
}

// NodeCoreController restarts/stops a node's core. *hub.Hub satisfies it.
type NodeCoreController interface {
	RestartCore(ctx context.Context, nodeID uuid.UUID) error
	StopCore(ctx context.Context, nodeID uuid.UUID) error
	UpdateGeo(ctx context.Context, nodeID uuid.UUID, geoipURL, geositeURL string) (geoip, geosite int64, err error)
}

// Iran-focused geo routing databases (country + domain rules), used by default
// when an admin refreshes a node's geo data. They add geoip:ir / geosite:ir and
// related categories so routing can target Iranian IPs and domains.
const (
	DefaultGeoIPURL   = "https://github.com/chocolate4u/Iran-v2ray-rules/releases/latest/download/geoip.dat"
	DefaultGeoSiteURL = "https://github.com/chocolate4u/Iran-v2ray-rules/releases/latest/download/geosite.dat"
)

// NodeService manages the node fleet's persistent records and keeps the hub's
// live management in step with them.
type NodeService struct {
	repo      port.NodeRepository
	registrar NodeRegistrar
	logs      NodeLogQuerier
	core      NodeCoreController
	geo       GeoCountryResolver
	now       func() time.Time
}

// NewNodeService wires the service.
func NewNodeService(repo port.NodeRepository, registrar NodeRegistrar) *NodeService {
	return &NodeService{repo: repo, registrar: registrar, now: time.Now}
}

// SetLogQuerier wires the source of live node logs (the hub).
func (s *NodeService) SetLogQuerier(q NodeLogQuerier) {
	if q != nil {
		s.logs = q
	}
}

// SetCoreController wires the restart/stop commands (the hub).
func (s *NodeService) SetCoreController(c NodeCoreController) {
	if c != nil {
		s.core = c
	}
}

// SetGeoResolver wires MaxMind (or disabled) country lookup for node locations.
func (s *NodeService) SetGeoResolver(g GeoCountryResolver) {
	s.geo = g
}

// Logs fetches up to limit recent core log lines from a node. Errors when no log
// source is wired or the node is unreachable.
func (s *NodeService) Logs(ctx context.Context, id uuid.UUID, limit int) ([]string, error) {
	if s.logs == nil {
		return nil, errors.New("node logs source not configured")
	}
	return s.logs.Logs(ctx, id, limit)
}

// RestartCore restarts a node's proxy engine.
func (s *NodeService) RestartCore(ctx context.Context, id uuid.UUID) error {
	if s.core == nil {
		return errors.New("core controller not configured")
	}
	return s.core.RestartCore(ctx, id)
}

// StopCore stops a node's proxy engine.
func (s *NodeService) StopCore(ctx context.Context, id uuid.UUID) error {
	if s.core == nil {
		return errors.New("core controller not configured")
	}
	return s.core.StopCore(ctx, id)
}

// UpdateGeo refreshes a node's geoip/geosite routing databases. Empty URLs fall
// back to the Iran-focused defaults. Returns the bytes written for each file.
func (s *NodeService) UpdateGeo(ctx context.Context, id uuid.UUID, geoipURL, geositeURL string) (geoip, geosite int64, err error) {
	if s.core == nil {
		return 0, 0, errors.New("core controller not configured")
	}
	if geoipURL == "" {
		geoipURL = DefaultGeoIPURL
	}
	if geositeURL == "" {
		geositeURL = DefaultGeoSiteURL
	}
	return s.core.UpdateGeo(ctx, id, geoipURL, geositeURL)
}

// CreateNodeInput describes a new node.
type CreateNodeInput struct {
	Name         string
	Address      string
	Core         domain.CoreType
	UsageRatio   float64
	Endpoint     string
	Region       string
	CountryCode  string
	LocationAuto *bool
}

// Create persists a node and immediately brings it under hub management so its
// traffic/health loops start without a panel restart.
func (s *NodeService) Create(ctx context.Context, in CreateNodeInput) (*domain.Node, error) {
	if in.Name == "" || in.Address == "" {
		return nil, errors.New("name and address are required")
	}
	core := in.Core
	if core == "" {
		core = domain.CoreXray
	}
	ratio := in.UsageRatio
	if ratio == 0 {
		ratio = 1
	}
	n := &domain.Node{
		ID:           uuid.New(),
		Name:         in.Name,
		Address:      in.Address,
		Core:         core,
		Status:       domain.NodeDisconnected,
		UsageRatio:   ratio,
		Endpoint:     in.Endpoint,
		Region:       in.Region,
		CountryCode:  in.CountryCode,
		LocationAuto: true,
		CreatedAt:    s.now(),
	}
	if in.LocationAuto != nil {
		n.LocationAuto = *in.LocationAuto
	}
	ApplyNodeGeo(n, s.geo)
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	if err := s.registrar.Register(ctx, n); err != nil {
		// Persisted but not yet managed; a panel restart (or re-register) recovers.
		return n, err
	}
	return n, nil
}

// List returns all nodes.
func (s *NodeService) List(ctx context.Context) ([]*domain.Node, error) {
	return s.repo.List(ctx)
}

// Get returns one node.
func (s *NodeService) Get(ctx context.Context, id uuid.UUID) (*domain.Node, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateNodeInput is the mutable subset of a node.
type UpdateNodeInput struct {
	Name         string
	Address      string
	UsageRatio   float64
	Endpoint     string
	Region       string
	CountryCode  string
	LocationAuto *bool
	RegionSet    bool
	CountrySet   bool
	LocationSet  bool
}

// Update persists node changes and re-establishes hub management so an address
// or ratio change takes effect immediately (the old connection is dropped and a
// fresh one dialed against the new address).
func (s *NodeService) Update(ctx context.Context, id uuid.UUID, in UpdateNodeInput) (*domain.Node, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != "" {
		n.Name = in.Name
	}
	if in.Address != "" {
		n.Address = in.Address
	}
	if in.UsageRatio > 0 {
		n.UsageRatio = in.UsageRatio
	}
	// Endpoint can be set to empty string to clear it (revert to real IP).
	n.Endpoint = in.Endpoint
	if in.RegionSet {
		n.Region = in.Region
	}
	if in.CountrySet {
		n.CountryCode = in.CountryCode
	}
	if in.LocationSet && in.LocationAuto != nil {
		n.LocationAuto = *in.LocationAuto
	}
	if n.LocationAuto {
		ApplyNodeGeo(n, s.geo)
	}
	if err := s.repo.Update(ctx, n); err != nil {
		return nil, err
	}
	// Re-register to pick up a possibly-changed address.
	s.registrar.Deregister(n.ID)
	if err := s.registrar.Register(ctx, n); err != nil {
		return n, err
	}
	return n, nil
}

// Delete stops managing a node and removes it (its inbounds cascade in the DB).
func (s *NodeService) Delete(ctx context.Context, id uuid.UUID) error {
	s.registrar.Deregister(id)
	return s.repo.Delete(ctx, id)
}
