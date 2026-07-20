package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core/reality"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/subscription"
)

// defaultStaleAfter is how long without a heartbeat before a node is pruned from
// subscriptions. Generous relative to the health interval so brief gaps don't
// flap a node out of clients' configs.
const defaultStaleAfter = 90 * time.Second

// SubscriptionService resolves a subscription token into the set of proxies a
// client should receive. It joins the user's credentials with each bound
// inbound's transport and the hosting node's public address, and prunes inbounds
// on nodes known to be unhealthy so clients aren't handed dead endpoints.
type SubscriptionService struct {
	users          port.UserRepository
	nodes          port.NodeRepository
	subHosts       port.SubHostRepository
	tlsTricks      port.TLSTricksRepository
	protocolGroups port.ProtocolGroupRepository
	ispProfiles    port.ISPProfileRepository
	packs          PackResolver
	staleAfter     time.Duration
	now            func() time.Time
}

// PackResolver resolves the routing pack a subscription should embed. It is the
// subset of *RoutingPackService the subscription service needs, kept small so
// the dependency is optional (nil-safe) and easy to fake in tests.
type PackResolver interface {
	GetUserPack(ctx context.Context, userID uuid.UUID) (string, error)
	GetGlobalDefault(ctx context.Context) (string, error)
	GetPack(ctx context.Context, id string) (*domain.RoutingPack, error)
}

// NewSubscriptionService wires the service. subHosts may be nil, in which case
// no Marzban-style host projection happens and every inbound emits its own
// single link exactly as before host support existed.
func NewSubscriptionService(users port.UserRepository, nodes port.NodeRepository, subHosts port.SubHostRepository) *SubscriptionService {
	return &SubscriptionService{users: users, nodes: nodes, subHosts: subHosts, staleAfter: defaultStaleAfter, now: time.Now}
}

// SetRoutingPacks injects the routing pack resolver used to embed a selected
// pack's rules into Clash/sing-box output. It is optional: with no resolver (or
// when no pack resolves) subscriptions render exactly as before (Req 3.3.3).
func (s *SubscriptionService) SetRoutingPacks(packs PackResolver) {
	s.packs = packs
}

// SetTLSTricks wires TLS/DPI profiles linked via inbound.evasion_profile_id so
// fragment and uTLS fingerprint land in client subscription configs.
func (s *SubscriptionService) SetTLSTricks(repo port.TLSTricksRepository) {
	s.tlsTricks = repo
}

// SetProtocolGroups wires the auto-protocol switching repositories so
// subscription rendering can discover groups and ISP profiles to emit per-group
// urltest/fallback outbounds. Optional: with nil repos the subscription renders
// a single flat group as before.
func (s *SubscriptionService) SetProtocolGroups(groups port.ProtocolGroupRepository, isp port.ISPProfileRepository) {
	s.protocolGroups = groups
	s.ispProfiles = isp
}

// SubResult bundles the resolved proxies with the owning user so the handler can
// render the body and emit usage headers. Rules carries the selected routing
// pack's rules (nil when none is selected) for Clash/sing-box embedding. Groups
// carries the auto-protocol-switching groups for per-group urltest/fallback
// rendering in Clash/sing-box.
type SubResult struct {
	User    *domain.User
	Proxies []subscription.Proxy
	Rules   []domain.RoutingRule
	Groups  []subscription.ProtocolGroupRender
}

// Build looks up the user by token and assembles a Proxy per enabled inbound.
func (s *SubscriptionService) Build(ctx context.Context, token string) (*SubResult, error) {
	user, err := s.users.GetBySubToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return s.buildFor(ctx, user)
}

// BuildForUser is the admin-facing counterpart of Build: it resolves a user's
// proxies by user id (not the opaque token) so the panel can show an operator a
// user's subscription links and QR codes on demand.
func (s *SubscriptionService) BuildForUser(ctx context.Context, userID uuid.UUID) (*SubResult, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.buildFor(ctx, user)
}

// buildFor assembles a Proxy per enabled inbound for a resolved user. Disabled
// inbounds and unreachable node lookups are skipped so one bad inbound never
// breaks the whole subscription.
func (s *SubscriptionService) buildFor(ctx context.Context, user *domain.User) (*SubResult, error) {
	inbounds, err := s.users.InboundsFor(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	cache := map[uuid.UUID]*nodeInfo{} // resolve each node once
	hostsByInbound := s.hostsFor(ctx, inbounds)
	var proxies []subscription.Proxy
	for _, in := range inbounds {
		if !in.Enabled {
			continue
		}
		info, ok := cache[in.NodeID]
		if !ok {
			info = s.resolveNode(ctx, in.NodeID, user.Username, now)
			cache[in.NodeID] = info
		}
		// Include every enabled inbound the user is bound to (3x-ui style); only
		// skip when the node record or its host is missing. A momentarily
		// unhealthy node still gets its config — the client just can't connect
		// until it recovers, rather than the config silently vanishing.
		if info == nil || info.host == "" {
			continue
		}
		base := buildProxy(user, in, info.host, info.name)
		if s.tlsTricks != nil && in.EvasionProfileID != nil {
			if profile, err := s.tlsTricks.GetByID(ctx, *in.EvasionProfileID); err == nil {
				applyTLSProfile(&base, profile)
			}
		}
		// No enabled hosts for this inbound: emit the inbound's own link exactly
		// as before (no regression).
		hosts := hostsByInbound[in.ID]
		if len(hosts) == 0 {
			proxies = append(proxies, base)
			continue
		}
		// One Proxy per enabled host, in priority order, overlaying the base.
		vars := subscription.FormatVars(user, info.host, "")
		for _, h := range hosts {
			proxies = append(proxies, buildProxyWithHost(base, h, vars))
		}
	}

	// Auto-protocol switching: discover groups for the user's inbounds and
	// annotate proxies with their group membership. Build ProtocolGroupRender
	// slices for Clash/sing-box to emit per-group urltest/fallback outbounds.
	groups := s.resolveProtocolGroups(ctx, inbounds, proxies)

	return &SubResult{User: user, Proxies: proxies, Rules: s.resolveRules(ctx, user.ID), Groups: groups}, nil
}

// resolveRules picks the routing rules to embed for a user: the user's selected
// pack if set, otherwise the global default. It is fail-open — a missing
// resolver, any resolution error, or an unresolved pack yields nil rules so the
// subscription renders unchanged and a pack lookup never breaks a subscription
// (Req 3.3.3).
func (s *SubscriptionService) resolveRules(ctx context.Context, userID uuid.UUID) []domain.RoutingRule {
	if s.packs == nil {
		return nil
	}
	packID, err := s.packs.GetUserPack(ctx, userID)
	if err != nil {
		return nil
	}
	if packID == "" {
		packID, err = s.packs.GetGlobalDefault(ctx)
		if err != nil || packID == "" {
			return nil
		}
	}
	pack, err := s.packs.GetPack(ctx, packID)
	if err != nil || pack == nil {
		return nil
	}
	return pack.Rules
}

// resolveProtocolGroups discovers which ProtocolGroups the user's inbounds belong
// to, annotates proxies with GroupName, and builds the ProtocolGroupRender slice
// for renderers. It is fail-open: a nil repo or any error yields nil groups so
// subscriptions render unchanged.
func (s *SubscriptionService) resolveProtocolGroups(ctx context.Context, inbounds []domain.Inbound, proxies []subscription.Proxy) []subscription.ProtocolGroupRender {
	if s.protocolGroups == nil {
		return nil
	}
	// Collect enabled inbound IDs.
	inboundIDs := make([]uuid.UUID, 0, len(inbounds))
	for _, in := range inbounds {
		if in.Enabled {
			inboundIDs = append(inboundIDs, in.ID)
		}
	}
	if len(inboundIDs) == 0 {
		return nil
	}

	domainGroups, err := s.protocolGroups.GroupsForInbounds(ctx, inboundIDs)
	if err != nil || len(domainGroups) == 0 {
		return nil
	}

	// Build a lookup: inbound ID → proxy names (a single inbound can produce
	// multiple proxies when SubHosts are active).
	inboundToProxyNames := make(map[uuid.UUID][]string)
	// Also build inbound ID → index in inbounds slice for protocol lookup.
	inboundByID := make(map[uuid.UUID]domain.Inbound, len(inbounds))
	for _, in := range inbounds {
		inboundByID[in.ID] = in
	}
	// Map proxy name back to the inbound that generated it. We rely on the fact
	// that buildProxy names the proxy after the node name which is stable per
	// inbound. We'll use inbound tag matching as a fallback. The simplest
	// approach: iterate proxies and match by inbound order since proxies are
	// generated in inbound order.
	idx := 0
	for _, in := range inbounds {
		if !in.Enabled {
			continue
		}
		// Count how many proxies this inbound generated (1 or N if SubHosts exist).
		start := idx
		for idx < len(proxies) && idx-start < maxProxiesPerInbound(proxies, start, in) {
			idx++
		}
		for i := start; i < idx; i++ {
			inboundToProxyNames[in.ID] = append(inboundToProxyNames[in.ID], proxies[i].Name)
		}
	}

	// Build the render groups.
	var renderGroups []subscription.ProtocolGroupRender
	for _, g := range domainGroups {
		var proxyNames []string
		// Inbounds in priority order.
		for _, ibID := range g.InboundIDs {
			proxyNames = append(proxyNames, inboundToProxyNames[ibID]...)
		}
		if len(proxyNames) == 0 {
			continue
		}
		// Annotate proxies with the group name.
		nameSet := make(map[string]bool, len(proxyNames))
		for _, n := range proxyNames {
			nameSet[n] = true
		}
		for i := range proxies {
			if nameSet[proxies[i].Name] && proxies[i].GroupName == "" {
				proxies[i].GroupName = g.Name
			}
		}
		renderGroups = append(renderGroups, subscription.ProtocolGroupRender{
			Name:          g.Name,
			ProbeURL:      g.ProbeURL,
			ProbeInterval: g.ProbeInterval,
			ProbeTimeout:  g.ProbeTimeout,
			MaxRetries:    g.MaxRetries,
			ProxyNames:    proxyNames,
		})
	}
	return renderGroups
}

// maxProxiesPerInbound counts how many consecutive proxies starting at 'start'
// were generated from a single inbound, based on the proxy list structure.
func maxProxiesPerInbound(proxies []subscription.Proxy, start int, in domain.Inbound) int {
	// A proxy belongs to this inbound if it shares the same port and protocol.
	// We stop as soon as we see a different (port, protocol) pair.
	count := 0
	for i := start; i < len(proxies); i++ {
		if proxies[i].Port == in.Port && proxies[i].Protocol == in.Protocol {
			count++
		} else {
			break
		}
	}
	if count == 0 {
		count = 1 // at least 1 proxy per enabled inbound
	}
	return count
}

// hostsFor batch-loads the enabled SubHosts for the user's enabled inbounds in a
// single query (avoiding N+1), grouped by inbound id and sorted by priority. It
// fails open: a nil repo or a query error yields no hosts, so the subscription
// degrades to the pre-host behavior rather than breaking.
func (s *SubscriptionService) hostsFor(ctx context.Context, inbounds []domain.Inbound) map[uuid.UUID][]*domain.SubHost {
	if s.subHosts == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(inbounds))
	for _, in := range inbounds {
		if in.Enabled {
			ids = append(ids, in.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	hosts, err := s.subHosts.ListByInbounds(ctx, ids)
	if err != nil {
		return nil
	}
	byInbound := make(map[uuid.UUID][]*domain.SubHost, len(ids))
	for _, h := range hosts {
		if h == nil || !h.Enabled {
			continue
		}
		byInbound[h.InboundID] = append(byInbound[h.InboundID], h)
	}
	for id := range byInbound {
		group := byInbound[id]
		sort.SliceStable(group, func(i, j int) bool { return group[i].Priority < group[j].Priority })
	}
	return byInbound
}

type nodeInfo struct {
	host string
	name string
	live bool
}

func (s *SubscriptionService) resolveNode(ctx context.Context, nodeID uuid.UUID, username string, now time.Time) *nodeInfo {
	node, err := s.nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil
	}
	// If the node has a custom endpoint (tunnel/CDN/relay), use that instead of
	// the real IP so clients connect via the intermediate address.
	host := node.Endpoint
	if host == "" {
		host = hostOf(node.Address)
	}
	return &nodeInfo{
		host: host,
		name: username + " @ " + node.Name,
		live: node.Live(now, s.staleAfter),
	}
}

func buildProxy(u *domain.User, in domain.Inbound, host, name string) subscription.Proxy {
	p := subscription.Proxy{
		Name:       name,
		Protocol:   in.Protocol,
		Host:       host,
		Port:       in.Port,
		Network:    in.Network,
		Security:   string(in.Security),
		Path:       in.Path,
		Flow:       in.Flow,
		SNI:        first(in.SNI),
		HostHeader: first(in.Host),
	}
	// A TLS inbound carrying an auto-generated self-signed cert needs clients to
	// skip verification, or the handshake times out.
	if in.Security == domain.SecurityTLS {
		if _, ok := in.Raw["tls"]; ok {
			p.AllowInsecure = true
		}
	}
	switch in.Protocol {
	case domain.ProtoVLESS:
		p.UUID = u.Proxies.VLESSUUID.String()
	case domain.ProtoVMess:
		p.UUID = u.Proxies.VMessUUID.String()
	case domain.ProtoTrojan:
		p.Password = u.Proxies.TrojanPass
	case domain.ProtoShadowsocks:
		p.Password = u.Proxies.ShadowsocksP
		p.SSMethod = u.Proxies.SSMethod
	case domain.ProtoHysteria2:
		p.Password = u.Proxies.TrojanPass
		// Extract hysteria2-specific settings from inbound Raw.
		if h2, ok := in.Raw["hysteria2"].(map[string]any); ok {
			if obfs := hysteria2ObfsFromRaw(h2["obfs"]); obfs != "" {
				p.Hy2Obfs = obfs
			}
			if up, ok := rawIntVal(h2["up_mbps"]); ok {
				p.Hy2Up = up
			}
			if down, ok := rawIntVal(h2["down_mbps"]); ok {
				p.Hy2Down = down
			}
		}
		// Auto-generate obfs if not explicitly set (must match server auto-gen
		// in xray/config.go hysteria2ObfsPassword).
		if p.Hy2Obfs == "" {
			h := sha256.Sum256([]byte("hy2obfs:" + in.Tag + ":" + in.ID.String()))
			p.Hy2Obfs = hex.EncodeToString(h[:8])
		}
		// Hysteria2 with self-signed cert needs insecure
		p.AllowInsecure = true
		// Hysteria2 requires SNI for TLS handshake. Default to host when not
		// explicitly set so clients always have a server_name for the TLS block.
		if p.SNI == "" {
			p.SNI = host
		}
	case domain.ProtoTUIC:
		p.UUID = u.Proxies.VLESSUUID.String()
		p.Password = u.Proxies.TrojanPass
	}
	// REALITY clients need the public key + short id; pull them from the
	// inbound's neutral reality params and prefer its server name as the SNI.
	if in.Security == domain.SecurityReality {
		rp := reality.ParseParams(in.Raw["reality"])
		p.PublicKey = rp.PublicKey
		p.ShortID = first(rp.ShortIDs)
		if len(rp.ServerNames) > 0 {
			p.SNI = rp.ServerNames[0]
		}
		p.Fingerprint = "chrome"
	}
	return p
}

// applyTLSProfile overlays fragment / uTLS / mux from a TLS tricks profile onto
// a subscription proxy. Host overrides still win later in buildProxyWithHost.
func applyTLSProfile(p *subscription.Proxy, profile *domain.TLSTrickProfile) {
	if p == nil || profile == nil || !profile.Enabled {
		return
	}
	if profile.UTLSFingerprint != "" {
		p.Fingerprint = profile.UTLSFingerprint
	}
	if profile.MuxEnabled {
		p.Mux = true
	}
	if profile.FragmentEnabled {
		size := profile.FragmentSize
		if size == "" {
			size = "10-30"
		}
		interval := profile.FragmentInterval
		if interval == "" {
			interval = "10-20"
		}
		packets := profile.FragmentPackets
		if packets == "" {
			packets = "tlshello"
		}
		p.Fragment = size + "," + interval + "," + packets
	}
	if profile.PaddingEnabled && profile.PaddingSize != "" {
		p.Padding = profile.PaddingSize
	}
	if profile.ECHEnabled {
		p.ECH = true
	}
}

// buildProxyWithHost overlays a SubHost onto the inbound's base Proxy, producing
// the Proxy a client receives for that (inbound × host) pairing. Fields the host
// leaves empty (or at inbound_default) fall through to the inbound's own value,
// so a host only customizes what it explicitly sets. Remark and Address run
// through template-variable expansion against vars.
func buildProxyWithHost(base subscription.Proxy, h *domain.SubHost, vars map[string]string) subscription.Proxy {
	p := base // copy by value; base.ALPN is nil so there is no slice aliasing
	if remark := subscription.Expand(h.Remark, vars); remark != "" {
		p.Name = remark
	}
	if addr := subscription.Expand(h.Address, vars); addr != "" {
		p.Host = addr
	}
	if h.Port != nil {
		p.Port = *h.Port
	}
	if h.SNI != "" {
		p.SNI = h.SNI
	}
	if h.HostHeader != "" {
		p.HostHeader = h.HostHeader
	}
	if h.Path != "" {
		p.Path = h.Path
	}
	if h.ALPN != "" {
		p.ALPN = splitCSV(h.ALPN)
	}
	if h.Fingerprint != "" {
		p.Fingerprint = h.Fingerprint
	}
	// Security override: only when the host forces a specific layer; otherwise
	// inherit the inbound's security untouched.
	if h.Security != domain.HostSecurityInboundDefault && h.Security != "" {
		p.Security = string(h.Security)
	}
	if h.AllowInsecure {
		p.AllowInsecure = true
	}
	p.Mux = h.MuxEnable
	if h.Fragment != "" {
		p.Fragment = h.Fragment
	}
	return p
}

// splitCSV splits a comma-separated list (e.g. an ALPN string "h2,http/1.1")
// into trimmed, non-empty parts. It returns nil for an empty input.
func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// ApplyISPHint applies ISP-specific anti-DPI settings to all proxies that don't
// already carry a fragment override (i.e. no per-inbound evasion profile applied).
// isp is the raw query-param value (e.g. "mci", "irancell", "shatel").
func ApplyISPHint(proxies []subscription.Proxy, isp string) {
	if isp == "" {
		return
	}
	// Map short aliases to canonical ISPPreset values.
	preset := mapISPAlias(isp)
	profile := domain.ISPPresetDefaults(preset)
	if !profile.Enabled {
		return
	}
	for i := range proxies {
		// Only apply ISP preset when the proxy doesn't already have fragment/mux
		// from a per-inbound evasion profile.
		if proxies[i].Fragment == "" && profile.FragmentEnabled {
			applyTLSProfile(&proxies[i], &profile)
		}
	}
}

// mapISPAlias normalizes common short ISP names (used as query params) into the
// canonical domain.ISPPreset constant.
func mapISPAlias(s string) domain.ISPPreset {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "mci", "hamrah_aval", "hamrahaval":
		return domain.ISPHamrahAval
	case "irancell", "mtn":
		return domain.ISPIrancell
	case "mokhaberat", "tci", "mci_fixed":
		return domain.ISPMokhaberat
	case "shatel":
		return domain.ISPShatel
	case "asiatech":
		return domain.ISPAsiatech
	default:
		return domain.ISPPreset(s)
	}
}

// hostOf extracts the host from a "host:port" node address, tolerating a bare host.
func hostOf(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

func first(ss []string) string {
	if len(ss) > 0 {
		return ss[0]
	}
	return ""
}

// hysteria2ObfsFromRaw extracts the obfuscation password from the raw inbound
// config. The "obfs" field may be either a plain string password or a map with
// a nested "password" key (depending on how the admin configured it).
func hysteria2ObfsFromRaw(v any) string {
	switch o := v.(type) {
	case string:
		return o
	case map[string]any:
		s, _ := o["password"].(string)
		return s
	default:
		return ""
	}
}

// rawIntVal coerces a JSON-decoded numeric value (which may arrive as int,
// float64, or int64 depending on the decoder) into a plain int.
func rawIntVal(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case float64:
		return int(n), true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
}
