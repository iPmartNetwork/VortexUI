package geoip

import (
	"net"
	"strings"
)

// HostOnly strips a port from host:port; bare hostnames/IPs pass through.
func HostOnly(hostPort string) string {
	hostPort = strings.TrimSpace(hostPort)
	if hostPort == "" {
		return ""
	}
	if h, _, err := net.SplitHostPort(hostPort); err == nil {
		return h
	}
	return hostPort
}

// IsLocal reports loopback or RFC1918 addresses used by in-process/local nodes.
func IsLocal(host string) bool {
	ip := net.ParseIP(HostOnly(host))
	if ip == nil {
		h := strings.ToLower(HostOnly(host))
		return h == "localhost" || h == "local"
	}
	return ip.IsLoopback() || ip.IsPrivate()
}
