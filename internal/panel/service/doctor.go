package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"time"
)

// DoctorCheck represents a single health check result.
type DoctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pass", "warn", "fail"
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// DoctorReport is the aggregated result of all health checks.
type DoctorReport struct {
	Checks    []DoctorCheck `json:"checks"`
	Summary   string        `json:"summary"` // "healthy", "degraded", "unhealthy"
	Timestamp time.Time     `json:"timestamp"`
}

// DoctorService runs system health diagnostics.
type DoctorService struct {
	dbDSN    string
	redisDSN string
	tlsSNIs  []string
	ports    []int
}

// NewDoctorService creates the doctor service.
func NewDoctorService(dbDSN, redisDSN string, tlsSNIs []string, ports []int) *DoctorService {
	return &DoctorService{
		dbDSN:    dbDSN,
		redisDSN: redisDSN,
		tlsSNIs:  tlsSNIs,
		ports:    ports,
	}
}

// RunAll executes all health checks and returns a report.
func (d *DoctorService) RunAll(ctx context.Context) *DoctorReport {
	report := &DoctorReport{
		Timestamp: time.Now(),
	}

	report.Checks = append(report.Checks, d.checkDatabase(ctx))
	report.Checks = append(report.Checks, d.checkRedis(ctx))
	report.Checks = append(report.Checks, d.checkDNS(ctx))
	report.Checks = append(report.Checks, d.checkPorts(ctx)...)
	report.Checks = append(report.Checks, d.checkTLS(ctx)...)
	report.Checks = append(report.Checks, d.checkDisk(ctx))

	// Summarize
	fails := 0
	warns := 0
	for _, c := range report.Checks {
		if c.Status == "fail" {
			fails++
		} else if c.Status == "warn" {
			warns++
		}
	}
	if fails > 0 {
		report.Summary = "unhealthy"
	} else if warns > 0 {
		report.Summary = "degraded"
	} else {
		report.Summary = "healthy"
	}

	return report
}

func (d *DoctorService) checkDatabase(ctx context.Context) DoctorCheck {
	check := DoctorCheck{Name: "Database connectivity"}
	start := time.Now()

	conn, err := net.DialTimeout("tcp", extractHost(d.dbDSN), 5*time.Second)
	if err != nil {
		check.Status = "fail"
		check.Message = fmt.Sprintf("cannot reach database: %v", err)
		return check
	}
	conn.Close()

	check.Status = "pass"
	check.Latency = time.Since(start).String()
	check.Message = "database reachable"
	return check
}

func (d *DoctorService) checkRedis(ctx context.Context) DoctorCheck {
	check := DoctorCheck{Name: "Redis connectivity"}
	start := time.Now()

	if d.redisDSN == "" {
		check.Status = "warn"
		check.Message = "Redis not configured"
		return check
	}

	conn, err := net.DialTimeout("tcp", extractHost(d.redisDSN), 5*time.Second)
	if err != nil {
		check.Status = "fail"
		check.Message = fmt.Sprintf("cannot reach Redis: %v", err)
		return check
	}
	conn.Close()

	check.Status = "pass"
	check.Latency = time.Since(start).String()
	check.Message = "Redis reachable"
	return check
}

func (d *DoctorService) checkDNS(ctx context.Context) DoctorCheck {
	check := DoctorCheck{Name: "DNS resolution"}
	start := time.Now()

	_, err := net.LookupHost("dns.google")
	if err != nil {
		check.Status = "fail"
		check.Message = fmt.Sprintf("DNS resolution failed: %v", err)
		return check
	}

	check.Status = "pass"
	check.Latency = time.Since(start).String()
	check.Message = "DNS working"
	return check
}

func (d *DoctorService) checkPorts(ctx context.Context) []DoctorCheck {
	var checks []DoctorCheck
	for _, port := range d.ports {
		check := DoctorCheck{Name: fmt.Sprintf("Port %d availability", port)}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			check.Status = "warn"
			check.Message = fmt.Sprintf("port %d already in use (expected if panel is running)", port)
		} else {
			ln.Close()
			check.Status = "pass"
			check.Message = "port available"
		}
		checks = append(checks, check)
	}
	return checks
}

func (d *DoctorService) checkTLS(ctx context.Context) []DoctorCheck {
	var checks []DoctorCheck
	for _, sni := range d.tlsSNIs {
		check := DoctorCheck{Name: fmt.Sprintf("TLS certificate: %s", sni)}

		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", sni+":443", &tls.Config{
			ServerName: sni,
		})
		if err != nil {
			check.Status = "warn"
			check.Message = fmt.Sprintf("TLS handshake failed: %v", err)
			checks = append(checks, check)
			continue
		}

		state := conn.ConnectionState()
		conn.Close()

		if len(state.PeerCertificates) > 0 {
			cert := state.PeerCertificates[0]
			daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
			if daysLeft < 7 {
				check.Status = "warn"
				check.Message = fmt.Sprintf("certificate expires in %d days", daysLeft)
			} else {
				check.Status = "pass"
				check.Message = fmt.Sprintf("valid, expires in %d days", daysLeft)
			}
		} else {
			check.Status = "warn"
			check.Message = "no peer certificates"
		}
		checks = append(checks, check)
	}
	return checks
}

func (d *DoctorService) checkDisk(ctx context.Context) DoctorCheck {
	check := DoctorCheck{Name: "Disk space"}

	// Simple check: ensure current dir is writable and has some space
	f, err := os.CreateTemp("", "vortexui-doctor-*")
	if err != nil {
		check.Status = "fail"
		check.Message = fmt.Sprintf("cannot write to temp: %v", err)
		return check
	}
	f.Close()
	os.Remove(f.Name())

	check.Status = "pass"
	check.Message = "disk writable"
	return check
}

// extractHost extracts host:port from a DSN-like string for basic connectivity check.
func extractHost(dsn string) string {
	// Handle postgres:// or redis:// URLs
	if len(dsn) > 10 {
		// Find @host:port/ pattern
		for i := len(dsn) - 1; i >= 0; i-- {
			if dsn[i] == '@' {
				rest := dsn[i+1:]
				for j := 0; j < len(rest); j++ {
					if rest[j] == '/' || rest[j] == '?' {
						return rest[:j]
					}
				}
				return rest
			}
		}
	}
	return dsn
}
