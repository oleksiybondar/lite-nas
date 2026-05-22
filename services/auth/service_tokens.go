package main

import (
	"crypto/subtle"
	"sync"
	"time"
)

// serviceTokenRecord tracks one active service-to-service token session.
type serviceTokenRecord struct {
	Service      string
	RefreshToken string
	ExpiresAt    time.Time
}

// serviceTokenStore keeps active service-to-service token sessions in memory.
type serviceTokenStore struct {
	mu      sync.RWMutex
	records map[string]serviceTokenRecord
	now     func() time.Time
}

// newServiceTokenStore creates an in-memory service-token store.
func newServiceTokenStore(now func() time.Time) *serviceTokenStore {
	if now == nil {
		now = time.Now
	}

	return &serviceTokenStore{
		records: make(map[string]serviceTokenRecord),
		now:     now,
	}
}

// Upsert stores or replaces the active token record for one service.
func (s *serviceTokenStore) Upsert(service string, refreshToken string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[service] = serviceTokenRecord{
		Service:      service,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
}

// Rotate atomically rotates a refresh token when service and token match.
func (s *serviceTokenStore) Rotate(service string, refreshToken string, newRefreshToken string, newExpiresAt time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[service]
	if !ok {
		return false
	}
	if s.now().After(record.ExpiresAt) {
		delete(s.records, service)
		return false
	}
	if subtle.ConstantTimeCompare([]byte(record.RefreshToken), []byte(refreshToken)) != 1 {
		return false
	}

	s.records[service] = serviceTokenRecord{
		Service:      service,
		RefreshToken: newRefreshToken,
		ExpiresAt:    newExpiresAt,
	}
	return true
}
