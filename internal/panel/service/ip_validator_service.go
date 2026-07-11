package service

import (
	"fmt"
	"log/slog"
	"net"
	"strings"
)

// IPValidatorService implements IP address validation and CIDR checking
type IPValidatorService struct {
	log *slog.Logger
}

// NewIPValidatorService creates a new IP validator service
func NewIPValidatorService(log *slog.Logger) *IPValidatorService {
	if log == nil {
		log = slog.Default()
	}
	return &IPValidatorService{
		log: log,
	}
}

// IsIPAllowed checks if IP is allowed for specific admin (considering both whitelist/blacklist)
func (i *IPValidatorService) IsIPAllowed(ipAddress string, adminRules []interface{}, globalRules []interface{}) (bool, error) {
	// First check global blacklist
	for _, rule := range globalRules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			ruleType := ruleMap["rule_type"].(string)
			if ruleType == "blacklist" && ruleMap["active"].(bool) {
				allowed, err := i.IPInRange(ipAddress, ruleMap["ip_address"].(string))
				if err == nil && allowed {
					i.log.Warn("IP blocked by global blacklist", "ip", ipAddress)
					return false, nil
				}
			}
		}
	}

	// Check admin-specific whitelist (if exists, only these IPs allowed)
	hasWhitelist := false
	for _, rule := range adminRules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			ruleType := ruleMap["rule_type"].(string)
			if ruleType == "whitelist" && ruleMap["active"].(bool) {
				hasWhitelist = true
				allowed, err := i.IPInRange(ipAddress, ruleMap["ip_address"].(string))
				if err == nil && allowed {
					return true, nil
				}
			}
		}
	}

	// If whitelist exists and we reach here, IP not in whitelist
	if hasWhitelist {
		i.log.Warn("IP not in admin whitelist", "ip", ipAddress)
		return false, nil
	}

	// Check admin-specific blacklist
	for _, rule := range adminRules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			ruleType := ruleMap["rule_type"].(string)
			if ruleType == "blacklist" && ruleMap["active"].(bool) {
				allowed, err := i.IPInRange(ipAddress, ruleMap["ip_address"].(string))
				if err == nil && allowed {
					i.log.Warn("IP blocked by admin blacklist", "ip", ipAddress)
					return false, nil
				}
			}
		}
	}

	return true, nil
}

// IsIPBlocked checks if IP is globally blocked
func (i *IPValidatorService) IsIPBlocked(ipAddress string, globalBlacklistRules []interface{}) (bool, error) {
	for _, rule := range globalBlacklistRules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			if !ruleMap["active"].(bool) {
				continue
			}
			allowed, err := i.IPInRange(ipAddress, ruleMap["ip_address"].(string))
			if err == nil && allowed {
				return true, nil
			}
		}
	}
	return false, nil
}

// IsIPWhitelisted checks if IP is in whitelist
func (i *IPValidatorService) IsIPWhitelisted(ipAddress string, whitelistRules []interface{}) (bool, error) {
	for _, rule := range whitelistRules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			if !ruleMap["active"].(bool) {
				continue
			}
			allowed, err := i.IPInRange(ipAddress, ruleMap["ip_address"].(string))
			if err == nil && allowed {
				return true, nil
			}
		}
	}
	return false, nil
}

// ParseCIDR validates CIDR notation
func (i *IPValidatorService) ParseCIDR(cidr string) (bool, error) {
	// Handle single IP addresses (convert to /32)
	if !strings.Contains(cidr, "/") {
		cidr = cidr + "/32"
	}

	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		i.log.Debug("invalid CIDR format", "cidr", cidr, "error", err)
		return false, err
	}

	return true, nil
}

// IPInRange checks if IP falls within CIDR range
func (i *IPValidatorService) IPInRange(ipAddress, cidrRange string) (bool, error) {
	// Handle single IP addresses
	if !strings.Contains(cidrRange, "/") {
		// Exact IP match
		return ipAddress == cidrRange, nil
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		i.log.Debug("invalid IP address", "ip", ipAddress)
		return false, nil
	}

	_, ipNet, err := net.ParseCIDR(cidrRange)
	if err != nil {
		i.log.Debug("invalid CIDR range", "cidr", cidrRange, "error", err)
		return false, err
	}

	return ipNet.Contains(ip), nil
}

// GetClientIP extracts client IP from various sources
func (i *IPValidatorService) GetClientIP(xForwardedFor, xRealIP, remoteAddr string) string {
	// Try X-Forwarded-For (may contain multiple IPs)
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP
	if xRealIP != "" {
		return xRealIP
	}

	// Use remote address directly
	if remoteAddr != "" {
		// Remove port if present
		if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
			return remoteAddr[:idx]
		}
		return remoteAddr
	}

	return ""
}

// ValidateIPRange validates an IP range format
func (i *IPValidatorService) ValidateIPRange(ipRange string) error {
	// Try as CIDR first
	if _, err := i.ParseCIDR(ipRange); err == nil {
		return nil
	}

	// Try as single IP
	if ip := net.ParseIP(ipRange); ip != nil {
		return nil
	}

	// Try as IP range (start-end)
	parts := strings.Split(ipRange, "-")
	if len(parts) == 2 {
		if ip := net.ParseIP(strings.TrimSpace(parts[0])); ip != nil {
			if ip := net.ParseIP(strings.TrimSpace(parts[1])); ip != nil {
				return nil
			}
		}
	}

	i.log.Debug("invalid IP range format", "range", ipRange)
	return fmt.Errorf("invalid IP range format: %s", ipRange)
}
