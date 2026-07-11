package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/platform/postgres"
)

// IPValidationMiddleware validates client IP against whitelist/blacklist rules
type IPValidationMiddleware struct {
	ipRuleRepo   *postgres.IPAccessRuleRepository
	ipValidator  *service.IPValidatorService
	log          *slog.Logger
}

// NewIPValidationMiddleware creates new IP validation middleware
func NewIPValidationMiddleware(ipRuleRepo *postgres.IPAccessRuleRepository, ipValidator *service.IPValidatorService, log *slog.Logger) *IPValidationMiddleware {
	if log == nil {
		log = slog.Default()
	}
	return &IPValidationMiddleware{
		ipRuleRepo:  ipRuleRepo,
		ipValidator: ipValidator,
		log:         log,
	}
}

// ValidateClientIP middleware checks if client IP is allowed
func (m *IPValidationMiddleware) ValidateClientIP() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract client IP
			clientIP := m.extractClientIP(c)
			if clientIP == "" {
				m.log.Warn("could not determine client IP")
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "could not determine client IP"})
			}

			// Store IP in context for later use
			c.Set("client_ip", clientIP)

			// Get admin ID if available
			adminIDVal := c.Get("admin_id")
			if adminIDVal == nil {
				// Not an authenticated request, check global IP rules only
				return m.validateGlobalIP(c, clientIP, next)
			}

			adminIDStr, ok := adminIDVal.(string)
			if !ok {
				return next(c)
			}

			adminID, err := uuid.Parse(adminIDStr)
			if err != nil {
				return next(c)
			}

			// Validate IP against admin-specific and global rules
			return m.validateAdminIP(c, adminID, clientIP, next)
		}
	}
}

// validateGlobalIP checks global IP rules
func (m *IPValidationMiddleware) validateGlobalIP(c echo.Context, clientIP string, next echo.HandlerFunc) error {
	// Get global blacklist rules
	globalBlacklist, err := m.ipRuleRepo.GetGlobalRules(c.Request().Context())
	if err != nil {
		m.log.Error("failed to get global blacklist", "error", err)
		// Continue anyway, don't block on error
		return next(c)
	}

	// Check if IP is globally blocked
	for _, rule := range globalBlacklist {
		if rule.Active {
			allowed, err := m.ipValidator.IPInRange(clientIP, rule.IPAddress)
			if err == nil && allowed {
				m.log.Warn("IP globally blocked", "ip", clientIP)
				return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
			}
		}
	}

	return next(c)
}

// validateAdminIP checks admin-specific and global rules
func (m *IPValidationMiddleware) validateAdminIP(c echo.Context, adminID uuid.UUID, clientIP string, next echo.HandlerFunc) error {
	// Get admin-specific rules
	adminRules, err := m.ipRuleRepo.ListRules(c.Request().Context(), &adminID, nil)
	if err != nil {
		m.log.Error("failed to get admin IP rules", "admin_id", adminID, "error", err)
		// Continue anyway, don't block on error
		return next(c)
	}

	// Get global rules
	globalRules, err := m.ipRuleRepo.GetGlobalRules(c.Request().Context())
	if err != nil {
		m.log.Error("failed to get global IP rules", "error", err)
		// Continue anyway, don't block on error
		return next(c)
	}

	// Check global blacklist first
	for _, rule := range globalRules {
		if rule.Active && rule.RuleType == "blacklist" {
			allowed, err := m.ipValidator.IPInRange(clientIP, rule.IPAddress)
			if err == nil && allowed {
				m.log.Warn("IP globally blacklisted", "ip", clientIP, "admin_id", adminID)
				return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
			}
		}
	}

	// Check admin-specific whitelist (if exists, only these IPs allowed)
	whitelistType := "whitelist"
	hasWhitelist := false
	for _, rule := range adminRules {
		if rule.Active && rule.RuleType == whitelistType {
			hasWhitelist = true
			allowed, err := m.ipValidator.IPInRange(clientIP, rule.IPAddress)
			if err == nil && allowed {
				// IP in whitelist, allow
				c.Set("ip_allowed", true)
				return next(c)
			}
		}
	}

	// If whitelist exists and we reach here, IP not in whitelist
	if hasWhitelist {
		m.log.Warn("IP not in admin whitelist", "ip", clientIP, "admin_id", adminID)
		return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
	}

	// Check admin-specific blacklist
	for _, rule := range adminRules {
		if rule.Active && rule.RuleType == "blacklist" {
			allowed, err := m.ipValidator.IPInRange(clientIP, rule.IPAddress)
			if err == nil && allowed {
				m.log.Warn("IP in admin blacklist", "ip", clientIP, "admin_id", adminID)
				return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
			}
		}
	}

	// IP allowed
	c.Set("ip_allowed", true)
	return next(c)
}

// extractClientIP extracts client IP from various sources
func (m *IPValidationMiddleware) extractClientIP(c echo.Context) string {
	// Try X-Forwarded-For (may contain multiple IPs, use first)
	if xff := c.Request().Header.Get("X-Forwarded-For"); xff != "" {
		// Get first IP from comma-separated list
		ips := parseIPList(xff)
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Try X-Real-IP
	if xri := c.Request().Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	if remoteAddr := c.Request().RemoteAddr; remoteAddr != "" {
		// Extract IP without port
		if idx := len(remoteAddr) - 1; idx >= 0 {
			for i := idx; i >= 0; i-- {
				if remoteAddr[i] == ':' {
					return remoteAddr[:i]
				}
			}
		}
		return remoteAddr
	}

	return ""
}

// parseIPList splits comma-separated IP list and trims spaces
func parseIPList(ips string) []string {
	result := []string{}
	current := ""
	for _, char := range ips {
		if char == ',' {
			if trimmed := trimSpace(current); trimmed != "" {
				result = append(result, trimmed)
			}
			current = ""
		} else {
			current += string(char)
		}
	}
	if trimmed := trimSpace(current); trimmed != "" {
		result = append(result, trimmed)
	}
	return result
}

// trimSpace removes leading/trailing spaces
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

// RequireWhitelistOnly helper to require whitelist rules for a route
func (m *IPValidationMiddleware) RequireWhitelistOnly(c echo.Context, adminID uuid.UUID) bool {
	whitelistType := "whitelist"
	rules, err := m.ipRuleRepo.ListRules(c.Request().Context(), &adminID, &whitelistType)
	if err != nil {
		return false
	}
	return len(rules) > 0
}
