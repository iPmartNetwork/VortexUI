package doh

import "strings"

// BlockMatcher checks whether a DNS name should be blocked.
type BlockMatcher struct {
	custom []string
	ads    []string
	malware []string
}

// NewBlockMatcher builds suffix matchers for custom, ads, and malware lists.
func NewBlockMatcher(custom []string, blockAds, blockMalware bool) *BlockMatcher {
	m := &BlockMatcher{custom: normalizeList(custom)}
	if blockAds {
		m.ads = adsSuffixes
	}
	if blockMalware {
		m.malware = malwareSuffixes
	}
	return m
}

// Blocked reports whether qname matches any active list.
func (m *BlockMatcher) Blocked(qname string) bool {
	if m == nil {
		return false
	}
	name := normalizeQName(qname)
	if name == "" {
		return false
	}
	for _, list := range [][]string{m.custom, m.ads, m.malware} {
		if suffixMatch(name, list) {
			return true
		}
	}
	return false
}

func normalizeQName(qname string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(qname)), ".")
}

func normalizeList(in []string) []string {
	out := make([]string, 0, len(in))
	for _, item := range in {
		if s := normalizeQName(item); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func suffixMatch(name string, list []string) bool {
	for _, blocked := range list {
		if name == blocked || strings.HasSuffix(name, "."+blocked) {
			return true
		}
	}
	return false
}

// Curated suffix lists — lightweight stand-ins for geosite ads/malware categories.
var adsSuffixes = []string{
	"doubleclick.net",
	"googlesyndication.com",
	"googleadservices.com",
	"adservice.google.com",
	"ads.google.com",
	"pagead2.googlesyndication.com",
	"adnxs.com",
	"taboola.com",
	"outbrain.com",
	"popads.net",
	"moatads.com",
	"scorecardresearch.com",
}

var malwareSuffixes = []string{
	"malware.com",
	"phishing.com",
	"botnet.cc",
	"stealer.top",
	"cryptolocker.ru",
	"badware.ru",
	"evil.com",
	"sinkhole.cert.pl",
	"urlvir.com",
}
