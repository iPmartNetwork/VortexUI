package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/platform/postgres"
)

// IPControlHandlers handles IP access control requests
type IPControlHandlers struct {
	ipRuleRepo   *postgres.IPAccessRuleRepository
	ipValidator  *service.IPValidatorService
	log          *slog.Logger
}

// NewIPControlHandlers creates new IP control handlers
func NewIPControlHandlers(ipRuleRepo *postgres.IPAccessRuleRepository, ipValidator *service.IPValidatorService, log *slog.Logger) *IPControlHandlers {
	if log == nil {
		log = slog.Default()
	}
	return &IPControlHandlers{
		ipRuleRepo:  ipRuleRepo,
		ipValidator: ipValidator,
		log:         log,
	}
}

// CreateRuleRequest creates a new IP rule
type CreateRuleRequest struct {
	RuleType    string `json:"rule_type"` // "whitelist" or "blacklist"
	IPAddress   string `json:"ip_address"` // CIDR notation
	Description string `json:"description,omitempty"`
}

// IPRuleResponse represents an IP rule in the response
type IPRuleResponse struct {
	ID          uuid.UUID `json:"id"`
	AdminID     *uuid.UUID `json:"admin_id,omitempty"`
	RuleType    string `json:"rule_type"`
	IPAddress   string `json:"ip_address"`
	Description string `json:"description,omitempty"`
	Active      bool `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateRule creates a new IP access rule
func (h *IPControlHandlers) CreateRule(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	var req CreateRuleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Validate rule type
	if req.RuleType != "whitelist" && req.RuleType != "blacklist" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "rule_type must be 'whitelist' or 'blacklist'"})
	}

	// Validate IP range format
	valid, err := h.ipValidator.ParseCIDR(req.IPAddress)
	if err != nil || !valid {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid IP address or CIDR range"})
	}

	// Create rule
	rule := &domain.IPAccessRule{
		ID:          uuid.New(),
		AdminID:     &adminID,
		RuleType:    req.RuleType,
		IPAddress:   req.IPAddress,
		Description: req.Description,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.ipRuleRepo.SaveRule(c.Request().Context(), rule); err != nil {
		h.log.Error("failed to save IP rule", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create rule"})
	}

	h.log.Info("IP rule created", "rule_id", rule.ID, "admin_id", adminID, "type", req.RuleType)
	return c.JSON(http.StatusCreated, h.ruleToResponse(rule))
}

// ListRulesResponse lists IP rules
type ListRulesResponse struct {
	Rules []IPRuleResponse `json:"rules"`
	Total int `json:"total"`
}

// ListRules retrieves IP rules for admin
func (h *IPControlHandlers) ListRules(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	ruleType := c.QueryParam("type") // Optional filter: "whitelist" or "blacklist"
	var ruleTypePtr *string
	if ruleType != "" {
		ruleTypePtr = &ruleType
	}

	rules, err := h.ipRuleRepo.ListRules(c.Request().Context(), &adminID, ruleTypePtr)
	if err != nil {
		h.log.Error("failed to list IP rules", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list rules"})
	}

	response := ListRulesResponse{
		Rules: make([]IPRuleResponse, len(rules)),
		Total: len(rules),
	}

	for i, rule := range rules {
		response.Rules[i] = h.ruleToResponse(&rule)
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateRuleRequest updates an IP rule
type UpdateRuleRequest struct {
	RuleType    *string `json:"rule_type,omitempty"`
	IPAddress   *string `json:"ip_address,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool `json:"active,omitempty"`
}

// UpdateRule updates an IP access rule
func (h *IPControlHandlers) UpdateRule(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	ruleID := c.Param("id")
	if ruleID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "rule ID required"})
	}

	ruleUUID, err := uuid.Parse(ruleID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid rule ID"})
	}

	var req UpdateRuleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Get existing rule
	rule, err := h.ipRuleRepo.GetRule(c.Request().Context(), ruleUUID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "rule not found"})
	}

	// Verify ownership
	if rule.AdminID != nil && *rule.AdminID != adminID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "cannot modify rule of another admin"})
	}

	// Update fields
	if req.RuleType != nil {
		if *req.RuleType != "whitelist" && *req.RuleType != "blacklist" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "rule_type must be 'whitelist' or 'blacklist'"})
		}
		rule.RuleType = *req.RuleType
	}

	if req.IPAddress != nil {
		valid, err := h.ipValidator.ParseCIDR(*req.IPAddress)
		if err != nil || !valid {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid IP address or CIDR range"})
		}
		rule.IPAddress = *req.IPAddress
	}

	if req.Description != nil {
		rule.Description = *req.Description
	}

	if req.Active != nil {
		rule.Active = *req.Active
	}

	rule.UpdatedAt = time.Now()

	if err := h.ipRuleRepo.UpdateRule(c.Request().Context(), rule); err != nil {
		h.log.Error("failed to update IP rule", "rule_id", ruleUUID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update rule"})
	}

	h.log.Info("IP rule updated", "rule_id", ruleUUID, "admin_id", adminID)
	return c.JSON(http.StatusOK, h.ruleToResponse(rule))
}

// DeleteRule deletes an IP access rule
func (h *IPControlHandlers) DeleteRule(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	ruleID := c.Param("id")
	if ruleID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "rule ID required"})
	}

	ruleUUID, err := uuid.Parse(ruleID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid rule ID"})
	}

	// Get rule to verify ownership
	rule, err := h.ipRuleRepo.GetRule(c.Request().Context(), ruleUUID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "rule not found"})
	}

	// Verify ownership
	if rule.AdminID != nil && *rule.AdminID != adminID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "cannot delete rule of another admin"})
	}

	if err := h.ipRuleRepo.DeleteRule(c.Request().Context(), ruleUUID); err != nil {
		h.log.Error("failed to delete IP rule", "rule_id", ruleUUID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete rule"})
	}

	h.log.Info("IP rule deleted", "rule_id", ruleUUID, "admin_id", adminID)
	return c.JSON(http.StatusOK, map[string]string{"message": "rule deleted successfully"})
}

// GetGlobalRulesResponse returns global IP rules
type GetGlobalRulesResponse struct {
	Rules []IPRuleResponse `json:"rules"`
	Total int `json:"total"`
}

// GetGlobalRules retrieves global (system-wide) IP rules
func (h *IPControlHandlers) GetGlobalRules(c echo.Context) error {
	// This should be restricted to system admins

	rules, err := h.ipRuleRepo.GetGlobalRules(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get global IP rules", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get rules"})
	}

	response := GetGlobalRulesResponse{
		Rules: make([]IPRuleResponse, len(rules)),
		Total: len(rules),
	}

	for i, rule := range rules {
		response.Rules[i] = h.ruleToResponse(&rule)
	}

	return c.JSON(http.StatusOK, response)
}

// Helper function to convert rule to response
func (h *IPControlHandlers) ruleToResponse(rule *domain.IPAccessRule) IPRuleResponse {
	return IPRuleResponse{
		ID:          rule.ID,
		AdminID:     rule.AdminID,
		RuleType:    rule.RuleType,
		IPAddress:   rule.IPAddress,
		Description: rule.Description,
		Active:      rule.Active,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}
