// Package hostmetrics reports coarse host CPU/memory/disk utilisation so a node
// (including the in-process local node) can surface real resource usage. The
// Linux implementation reads /proc and statfs; other platforms return zeros.
package hostmetrics

// Metrics is a utilisation snapshot, each field a percentage 0–100.
type Metrics struct {
	CPU  float64
	Mem  float64
	Disk float64
}
