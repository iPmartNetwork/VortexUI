package acme

import (
	"context"
	"sync"
)

// MemoryStore is an in-process CertStore for ACME issuance caching.
type MemoryStore struct {
	mu    sync.RWMutex
	certs map[string]*Certificate
}

// NewMemoryStore builds an empty in-memory cert store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{certs: make(map[string]*Certificate)}
}

func (m *MemoryStore) Get(_ context.Context, domain string) (*Certificate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if c, ok := m.certs[domain]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, nil
}

func (m *MemoryStore) Put(_ context.Context, cert *Certificate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *cert
	m.certs[cert.Domain] = &cp
	return nil
}
