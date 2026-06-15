//go:build !linux

package hostmetrics

// Read returns zero metrics on non-Linux platforms (dev machines); production
// nodes run Linux.
func Read() Metrics { return Metrics{} }
