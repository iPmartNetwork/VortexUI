package hub

import (
	"errors"
	"strings"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

func classifyDialError(err error) domain.NodeDiagnostics {
	return buildDiagnostics(err, false)
}

func buildDiagnostics(err error, networkOK bool) domain.NodeDiagnostics {
	now := time.Now()
	if err == nil {
		return domain.NodeDiagnostics{
			Code:             domain.NodeDiagOK,
			Message:          "connected",
			NetworkReachable: networkOK,
			CAMatch:          true,
			CheckedAt:        &now,
		}
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "x509"),
		strings.Contains(lower, "certificate"),
		strings.Contains(lower, "handshake"),
		strings.Contains(lower, "unknown authority"),
		strings.Contains(lower, "tls"):
		return domain.NodeDiagnostics{
			Code:             domain.NodeDiagMTLS,
			Message:          "mTLS handshake failed — node certs likely from a different panel install",
			NetworkReachable: networkOK,
			CAMatch:          false,
			CheckedAt:        &now,
		}
	case strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "no route to host"),
		strings.Contains(lower, "network is unreachable"),
		strings.Contains(lower, "no such host"),
		strings.Contains(lower, "timeout"),
		strings.Contains(lower, "deadline exceeded"):
		return domain.NodeDiagnostics{
			Code:             domain.NodeDiagUnreachable,
			Message:          msg,
			NetworkReachable: networkOK,
			CAMatch:          false,
			CheckedAt:        &now,
		}
	default:
		return domain.NodeDiagnostics{
			Code:             domain.NodeDiagUnknown,
			Message:          msg,
			NetworkReachable: networkOK,
			CAMatch:          false,
			CheckedAt:        &now,
		}
	}
}

func deriveDiag(status domain.NodeStatus, health domain.NodeHealth, lastErr string, networkOK bool) domain.NodeDiagnostics {
	now := time.Now()
	if status == domain.NodeConnected && health.CoreRunning {
		return domain.NodeDiagnostics{
			Code:             domain.NodeDiagOK,
			Message:          "connected",
			NetworkReachable: true,
			CAMatch:          true,
			CheckedAt:        &now,
		}
	}
	if status == domain.NodeConnected && !health.CoreRunning {
		return domain.NodeDiagnostics{
			Code:             domain.NodeDiagCoreDown,
			Message:          "agent reachable but proxy core is not running",
			NetworkReachable: true,
			CAMatch:          true,
			CheckedAt:        &now,
		}
	}
	if lastErr != "" {
		return buildDiagnostics(errors.New(lastErr), networkOK)
	}
	return domain.NodeDiagnostics{
		Code:             domain.NodeDiagUnreachable,
		Message:          "not connected",
		NetworkReachable: networkOK,
		CAMatch:          false,
		CheckedAt:        &now,
	}
}
