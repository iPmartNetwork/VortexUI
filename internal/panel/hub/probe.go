package hub

import (
	"context"
	"net"
	"time"
)

func probeNetwork(ctx context.Context, address string) bool {
	if address == "" {
		return false
	}
	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
