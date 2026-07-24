package subscription

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// Template variable tokens supported in subscription templates (profile title,
// remark, address). They are substituted at render time per user/node/proxy
// context. All 20+ variables are always present in the resolved map (empty when
// the panel has no clean value), so Expand never has to guess.
const (
	// Server & network
	varServerIP   = "{SERVER_IP}"
	varServerIPv6 = "{SERVER_IPV6}"

	// User identity
	varUsername      = "{USERNAME}"
	varAdminUsername = "{ADMIN_USERNAME}"

	// Protocol & transport
	varProtocol  = "{PROTOCOL}"
	varTransport = "{TRANSPORT}"

	// Data accounting
	varDataUsage       = "{DATA_USAGE}"
	varDataLimit       = "{DATA_LIMIT}"
	varDataLeft        = "{DATA_LEFT}"
	varUsagePercentage = "{USAGE_PERCENTAGE}"

	// Time & expiry
	varDaysLeft      = "{DAYS_LEFT}"
	varExpireDate    = "{EXPIRE_DATE}"
	varJalaliExpire  = "{JALALI_EXPIRE_DATE}"
	varTimeLeft      = "{TIME_LEFT}"

	// Status
	varStatusEmoji = "{STATUS_EMOJI}"

	// Node metadata
	varNodeName = "{NODE_NAME}"
	varNodeFlag = "{NODE_FLAG}"

	// ISP & quality
	varISPName      = "{ISP_NAME}"
	varQualityScore = "{QUALITY_SCORE}"
	varOnlineCount  = "{ONLINE_COUNT}"
)

// unlimited is the symbol used for an unbounded data limit / never-expiring
// account, matching the convention used by Marzban-style panels.
const unlimited = "∞"

// Status emoji mapping: reflects the user's lifecycle state as a visual icon.
var statusEmojiMap = map[domain.UserStatus]string{
	domain.UserStatusActive:   "✅",
	domain.UserStatusLimited:  "🟡",
	domain.UserStatusExpired:  "❌",
	domain.UserStatusDisabled: "🔴",
	domain.UserStatusOnHold:   "⏸️",
}

// StatusEmoji returns the emoji representing the given user status.
func StatusEmoji(status domain.UserStatus) string {
	if e, ok := statusEmojiMap[status]; ok {
		return e
	}
	return ""
}

// VarContext holds all contextual information needed to resolve template
// variables for a subscription render. Fields are populated by the subscription
// service from the request and panel state.
type VarContext struct {
	User          *domain.User
	AdminUsername string // resolved admin display name (empty if unknown)
	NodeName      string // display name of the node
	NodeIP        string // node's public IPv4 address
	NodeIPv6      string // node's public IPv6 address (empty if unavailable)
	NodeFlag      string // country flag emoji for the node (e.g. "🇩🇪")
	Protocol      string // protocol label (e.g. "vless", "trojan")
	Transport     string // transport type (e.g. "ws", "grpc", "tcp")
	ISP           string // ISP display name
	OnlineCount   int    // current online user count on this node
	QualityScore  int    // connection quality score 0-100
}

// VarResolver resolves template variables given a VarContext. It is stateless
// and safe for concurrent use. The Now field can be set for testing; when nil,
// time.Now() is used.
type VarResolver struct {
	Now func() time.Time // injectable clock for testing; nil = time.Now
}

// now returns the current time, using the injectable clock if set.
func (r *VarResolver) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

// Resolve replaces all {VAR} tokens in template with their resolved values from
// the given context. Unknown tokens are left literal.
func (r *VarResolver) Resolve(template string, ctx VarContext) string {
	vars := r.buildVars(ctx)
	return Expand(template, vars)
}

// BuildVars constructs the full variable map for a given context. Exported for
// callers who need the raw map (e.g. the legacy FormatVars path).
func (r *VarResolver) BuildVars(ctx VarContext) map[string]string {
	return r.buildVars(ctx)
}

// buildVars is the internal variable resolution engine.
func (r *VarResolver) buildVars(ctx VarContext) map[string]string {
	now := r.now()

	vars := map[string]string{
		varServerIP:        ctx.NodeIP,
		varServerIPv6:      ctx.NodeIPv6,
		varUsername:        "",
		varAdminUsername:   ctx.AdminUsername,
		varProtocol:       ctx.Protocol,
		varTransport:      ctx.Transport,
		varDataUsage:      "0 B",
		varDataLimit:      unlimited,
		varDataLeft:       unlimited,
		varUsagePercentage: "0%",
		varDaysLeft:       unlimited,
		varExpireDate:     "",
		varJalaliExpire:   "",
		varTimeLeft:       unlimited,
		varStatusEmoji:    "",
		varNodeName:       ctx.NodeName,
		varNodeFlag:       ctx.NodeFlag,
		varISPName:        ctx.ISP,
		varQualityScore:   fmt.Sprintf("%d", ctx.QualityScore),
		varOnlineCount:    fmt.Sprintf("%d", ctx.OnlineCount),
	}

	u := ctx.User
	if u == nil {
		return vars
	}

	vars[varUsername] = u.Username
	vars[varDataUsage] = humanBytes(u.UsedTraffic)
	vars[varStatusEmoji] = StatusEmoji(u.DerivedStatus(now))

	if u.DataLimit > 0 {
		vars[varDataLimit] = humanBytes(u.DataLimit)
		left := u.DataLimit - u.UsedTraffic
		if left < 0 {
			left = 0
		}
		vars[varDataLeft] = humanBytes(left)
		pct := float64(u.UsedTraffic) / float64(u.DataLimit) * 100
		if pct > 100 {
			pct = 100
		}
		vars[varUsagePercentage] = fmt.Sprintf("%.0f%%", pct)
	}

	if u.ExpireAt != nil {
		vars[varExpireDate] = u.ExpireAt.Format("2006-01-02")
		vars[varJalaliExpire] = gregorianToJalali(*u.ExpireAt)

		dur := u.ExpireAt.Sub(now)
		if dur < 0 {
			dur = 0
		}
		days := int(dur.Hours() / 24)
		vars[varDaysLeft] = fmt.Sprintf("%d", days)
		vars[varTimeLeft] = formatDuration(dur)
	}

	return vars
}

// FormatVars builds the substitution map for a user against a specific node's
// public address. This is the legacy API — new code should use VarResolver.
// It is preserved for backward compatibility with existing callers.
func FormatVars(u *domain.User, serverIP, serverIP6 string) map[string]string {
	return formatVarsAt(u, serverIP, serverIP6, time.Now())
}

// formatVarsAt is the testable core of FormatVars with an injectable clock.
func formatVarsAt(u *domain.User, serverIP, serverIP6 string, now time.Time) map[string]string {
	r := &VarResolver{Now: func() time.Time { return now }}
	ctx := VarContext{
		User:   u,
		NodeIP: serverIP,
		NodeIPv6: serverIP6,
	}
	return r.buildVars(ctx)
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

// formatDuration renders a time.Duration as a human-friendly string like
// "5d 3h" or "2h 15m" or "45m". Returns "0m" for non-positive durations.
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	switch {
	case days > 0 && hours > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case days > 0:
		return fmt.Sprintf("%dd", days)
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
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

// --- Jalali (Shamsi) calendar conversion ---

// gregorianToJalali converts a Gregorian time.Time to a Jalali date string
// formatted as "YYYY/MM/DD". Uses the algorithm by Kazimierz M. Borkowski.
func gregorianToJalali(t time.Time) string {
	jy, jm, jd := toJalali(t.Year(), int(t.Month()), t.Day())
	return fmt.Sprintf("%04d/%02d/%02d", jy, jm, jd)
}

// toJalali converts a Gregorian date (year, month, day) to Jalali (year, month, day).
// Algorithm based on the work of Kazimierz M. Borkowski (1988).
func toJalali(gy, gm, gd int) (int, int, int) {
	var gdm = [12]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}

	var gy2 int
	if gm > 2 {
		gy2 = gy + 1
	} else {
		gy2 = gy
	}

	days := 355666 + (365 * gy) + intDiv(gy2+3, 4) - intDiv(gy2+99, 100) + intDiv(gy2+399, 400) + gd + gdm[gm-1]
	jy := -1595 + (33 * intDiv(days, 12053))
	days = days % 12053

	jy += 4 * intDiv(days, 1461)
	days = days % 1461

	if days > 366 {
		jy += intDiv(days-1, 365)
		days = (days - 1) % 365
	}

	var jm, jd int
	if days < 186 {
		jm = 1 + intDiv(days, 31)
		jd = 1 + (days % 31)
	} else {
		jm = 7 + intDiv(days-186, 30)
		jd = 1 + ((days - 186) % 30)
	}

	return jy, jm, jd
}

// intDiv performs integer division truncating toward negative infinity (floor division).
func intDiv(a, b int) int {
	return int(math.Floor(float64(a) / float64(b)))
}
