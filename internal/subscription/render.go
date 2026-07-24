package subscription

// RenderConfig holds the template settings from sub_settings that control how
// subscription output is customized. These templates contain format variable
// tokens (e.g., {USERNAME}, {NODE_NAME}) that are resolved at render time.
type RenderConfig struct {
	ProfileTitleTemplate string
	RemarkTemplate       string
	AddressTemplate      string
}

// ApplyTemplates applies the remark and address templates to each proxy entry,
// resolving format variables with per-proxy context (different Protocol,
// Transport, NodeName per entry). It returns a new slice with modified proxies;
// the original slice is not mutated.
func ApplyTemplates(config RenderConfig, resolver *VarResolver, ctx VarContext, proxies []Proxy) []Proxy {
	if len(proxies) == 0 {
		return proxies
	}

	result := make([]Proxy, len(proxies))
	copy(result, proxies)

	for i := range result {
		// Build per-proxy context by overriding protocol/transport/node fields
		// from the proxy itself, so each entry resolves with its own values.
		proxyCtx := ctx
		proxyCtx.Protocol = string(result[i].Protocol)
		proxyCtx.Transport = result[i].Network
		// NodeName is already set in ctx from the caller; proxies from different
		// nodes would have different contexts passed in by the subscription service.

		// Apply remark template to the proxy's display name.
		if config.RemarkTemplate != "" {
			result[i].Name = resolver.Resolve(config.RemarkTemplate, proxyCtx)
		}

		// Apply address template to the proxy's host field if template is not empty.
		if config.AddressTemplate != "" {
			result[i].Host = resolver.Resolve(config.AddressTemplate, proxyCtx)
		}
	}

	return result
}

// ResolveProfileTitle resolves the profile title template using the given
// variable context. This produces the subscription-level title displayed in
// client applications (e.g., the selector group name).
func ResolveProfileTitle(config RenderConfig, resolver *VarResolver, ctx VarContext) string {
	if config.ProfileTitleTemplate == "" {
		return ""
	}
	return resolver.Resolve(config.ProfileTitleTemplate, ctx)
}
