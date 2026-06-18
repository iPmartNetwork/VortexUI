// Package geoip resolves IP addresses to ISO 3166-1 alpha-2 country codes using
// a MaxMind GeoLite2-Country database. A disabled resolver (no DB configured)
// returns "" for every lookup, so geo features degrade gracefully.
package geoip

import (
	"net"
	"sync"

	"github.com/oschwald/maxminddb-golang"
)

// Resolver wraps a MaxMind database reader. The zero value (and a nil pointer)
// is a valid disabled resolver.
type Resolver struct {
	mu sync.RWMutex
	db *maxminddb.Reader
}

// Open loads the mmdb at path. An empty path returns a disabled resolver with no
// error. A non-empty path that cannot be opened returns an error so the caller
// can log it and fall back to a disabled resolver.
func Open(path string) (*Resolver, error) {
	if path == "" {
		return &Resolver{}, nil
	}
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}
	return &Resolver{db: db}, nil
}

// Disabled returns a resolver that always yields "".
func Disabled() *Resolver { return &Resolver{} }

// Enabled reports whether a database is loaded.
func (r *Resolver) Enabled() bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.db != nil
}

// Country returns the ISO country code for ip, or "" if unknown/disabled.
func (r *Resolver) Country(ip string) string {
	if r == nil {
		return ""
	}
	r.mu.RLock()
	db := r.db
	r.mu.RUnlock()
	if db == nil {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	var rec struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}
	if err := db.Lookup(parsed, &rec); err != nil {
		return ""
	}
	return rec.Country.ISOCode
}

// Close releases the database.
func (r *Resolver) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.db != nil {
		err := r.db.Close()
		r.db = nil
		return err
	}
	return nil
}
