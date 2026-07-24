package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// WireGuardHandler serves WireGuard peer management, repair, QR code, and mesh endpoints.
type WireGuardHandler struct {
	svc *service.WireGuardService
}

// NewWireGuardHandler creates a new WireGuardHandler with the given service.
func NewWireGuardHandler(svc *service.WireGuardService) *WireGuardHandler {
	return &WireGuardHandler{svc: svc}
}

// Register mounts all WireGuard management routes on the given Echo group.
func (h *WireGuardHandler) Register(g *echo.Group) {
	wg := g.Group("/wireguard")

	// Peer management for a specific inbound
	wg.GET("/:inbound_id/peers", h.ListPeers)
	wg.POST("/:inbound_id/peers", h.CreatePeer)
	wg.PUT("/:inbound_id/peers/:user_id", h.UpdatePeerSettings)
	wg.GET("/:inbound_id/peers/:user_id/qr", h.GetQRCode)
	wg.POST("/:inbound_id/repair", h.RepairPeers)

	// Mesh management
	wg.POST("/mesh", h.CreateMesh)
	wg.GET("/mesh", h.ListMeshes)
	wg.GET("/mesh/:id", h.GetMesh)
}

// --- request/response types ---

type updatePeerSettingsRequest struct {
	MTU int    `json:"mtu"`
	DNS string `json:"dns"`
}

type createMeshRequest struct {
	Name      string            `json:"name"`
	CIDR      string            `json:"cidr"`
	NodeIDs   []string          `json:"node_ids"`
	Endpoints map[string]string `json:"endpoints"` // node_id -> host:port
}

// --- handlers ---

// ListPeers handles GET /api/v2/wireguard/:inbound_id/peers.
func (h *WireGuardHandler) ListPeers(c echo.Context) error {
	inboundID, err := uuid.Parse(c.Param("inbound_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}

	peers, err := h.svc.ListPeers(c.Request().Context(), inboundID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, peers)
}

// CreatePeer handles POST /api/v2/wireguard/:inbound_id/peers.
// Creates a new peer for a user (user_id in body).
func (h *WireGuardHandler) CreatePeer(c echo.Context) error {
	inboundID, err := uuid.Parse(c.Param("inbound_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	// EnsurePeers with a minimal User struct for the allocation
	user := &domain.User{ID: userID}
	in := domain.Inbound{ID: inboundID}
	peers, err := h.svc.EnsurePeers(c.Request().Context(), in, []*domain.User{user})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if len(peers) == 0 {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create peer")
	}
	return c.JSON(http.StatusCreated, peers[0])
}

// UpdatePeerSettings handles PUT /api/v2/wireguard/:inbound_id/peers/:user_id.
func (h *WireGuardHandler) UpdatePeerSettings(c echo.Context) error {
	inboundID, err := uuid.Parse(c.Param("inbound_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	var req updatePeerSettingsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	peer, err := h.svc.UpdatePeerSettings(c.Request().Context(), inboundID, userID, domain.WireGuardPeerSettings{
		MTU: req.MTU,
		DNS: req.DNS,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, peer)
}

// GetQRCode handles GET /api/v2/wireguard/:inbound_id/peers/:user_id/qr.
func (h *WireGuardHandler) GetQRCode(c echo.Context) error {
	inboundID, err := uuid.Parse(c.Param("inbound_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	// Endpoint host from query (required for QR generation)
	endpointHost := c.QueryParam("endpoint")
	if endpointHost == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "endpoint query param is required")
	}

	user := &domain.User{ID: userID}
	in := domain.Inbound{ID: inboundID}

	qrPNG, err := h.svc.GenerateQR(c.Request().Context(), in, user, endpointHost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Blob(http.StatusOK, "image/png", qrPNG)
}

// RepairPeers handles POST /api/v2/wireguard/:inbound_id/repair.
func (h *WireGuardHandler) RepairPeers(c echo.Context) error {
	inboundID, err := uuid.Parse(c.Param("inbound_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}

	in := domain.Inbound{ID: inboundID}
	report, err := h.svc.RepairPeers(c.Request().Context(), in)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, report)
}

// CreateMesh handles POST /api/v2/wireguard/mesh.
func (h *WireGuardHandler) CreateMesh(c echo.Context) error {
	var req createMeshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if len(req.NodeIDs) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "at least 2 nodes are required for a mesh")
	}

	nodeIDs := make([]uuid.UUID, 0, len(req.NodeIDs))
	for _, id := range req.NodeIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id: "+id)
		}
		nodeIDs = append(nodeIDs, uid)
	}

	endpoints := make(map[uuid.UUID]string)
	for k, v := range req.Endpoints {
		uid, err := uuid.Parse(k)
		if err != nil {
			continue
		}
		endpoints[uid] = v
	}

	mesh, err := h.svc.CreateMesh(c.Request().Context(), req.Name, req.CIDR, nodeIDs, endpoints)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, mesh)
}

// ListMeshes handles GET /api/v2/wireguard/mesh.
func (h *WireGuardHandler) ListMeshes(c echo.Context) error {
	meshes, err := h.svc.ListMeshes(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, meshes)
}

// GetMesh handles GET /api/v2/wireguard/mesh/:id.
func (h *WireGuardHandler) GetMesh(c echo.Context) error {
	meshID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid mesh ID")
	}

	mesh, err := h.svc.GetMesh(c.Request().Context(), meshID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mesh)
}
