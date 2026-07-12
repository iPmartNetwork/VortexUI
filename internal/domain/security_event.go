package domain

// SecurityEvent is a runtime security observation pushed from a node agent.
type SecurityEvent struct {
	SourceIP    string
	Port        int
	Method      string // tls_probe | http_probe | fingerprint
	Fingerprint string
	UserAgent   string
}
