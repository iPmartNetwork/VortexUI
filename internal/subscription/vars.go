package subscription

import (
	"fmt"
	"strings"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// Template variable tokens supported in a SubHost's remark and address. They are
// substituted at render time per user/node. The set mirrors the subset of
// Marzban/Pasarguard's FormatVariables that VortexUI can supply today; tokens
// the panel cannot resolve cleanly are still defined (as empty strings) so a
// host definition referencing them never crashes and never leaks a stray token.
const (
	varUsername      = "{USERNAME}"
	varServerIP      = "{SERVER_IP}"
	varServerIPv6    = "{SERVER_IPV6}"
	varDataUsage     = "{DATA_USAGE}"
	varDataLimit     = "{DATA_LIMIT}"
	varDataLeft      = "{DATA_LEFT}"
	varDaysLeft      = "{DAYS_LEFT}"
	varExpireDate    = "{EXPIRE_DATE}"
	varAdminUsername = "{ADMIN_USERNAME}"
)

// unlimited is the symbol used for an unbounded data limit / never-expiring
// account, matching the convention used by Marzban-style panels.
const unlimited = "∞"

// FormatVars builds the substitution map for a user against a specific node's
// public address. serverIP / serverIP6 come from the resolved node host used to
// build the proxy; pass "" for an IPv6 the panel cannot determine. Every
// supported token is always present in the returned map (empty when the panel
// has no clean value), so Expand never has to guess.
func FormatVars(u *domain.User, serverIP, serverIP6 string) map[string]string {
	return formatVarsAt(u, serverIP, serverIP6, time.Now())
}

// formatVarsAt is the testable core of FormatVars with an injectable clock.
func formatVarsAt(u *domain.User, serverIP, serverIP6 string, now time.Time) map[string]string {
	vars := map[string]string{
		varUsername:      "",
		varServerIP:      serverIP,
		varServerIPv6:    serverIP6,
		varDataUsage:     "0 B",
		varDataLimit:     unlimited,
		varDataLeft:      unlimited,
		varDaysLeft:      unlimited,
		varExpireDate:    "",
		varAdminUsername: "", // panel has only an admin id on the user, not a name
	}
	if u == nil {
		return vars
	}

	vars[varUsername] = u.Username
	vars[varDataUsage] = humanBytes(u.UsedTraffic)

	if u.DataLimit > 0 {
		vars[varDataLimit] = humanBytes(u.DataLimit)
		left := u.DataLimit - u.UsedTraffic
		if left < 0 {
			left = 0
		}
		vars[varDataLeft] = humanBytes(left)
	}

	if u.ExpireAt != nil {
		vars[varExpireDate] = u.ExpireAt.Format("2006-01-02")
		days := int(u.ExpireAt.Sub(now).Hours() / 24)
		if days < 0 {
			days = 0
		}
		vars[varDaysLeft] = fmt.Sprintf("%d", days)
	}

	return vars
}

// Expand replaces every known {TOKEN} in s with its value from vars in a single
// left-to-right pass, leaving unknown tokens (and unmatched braces) literal —
// mirroring Marzban's FormatVariables.__missing__ behavior. It never errors.
func Expand(s string, vars map[string]string) string {
	if !strings.ContainsRune(s, '{') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '{' {
			if j := strings.IndexByte(s[i+1:], '}'); j >= 0 {
				token := s[i : i+1+j+1] // includes the surrounding braces
				if v, ok := vars[token]; ok {
					b.WriteString(v)
				} else {
					b.WriteString(token) // unknown: leave literal
				}
				i += 1 + j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// humanBytes renders a byte count in binary units (KiB-style steps shown as
// KB/MB/... for brevity), used by the DATA_* template variables.
func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
