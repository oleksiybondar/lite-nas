package sessions

import (
	"sort"
	"sync"
	"time"
)

// StoreOptions controls refresh-token context enforcement.
type StoreOptions struct {
	EnforceClientIP bool
	RevocationTTL   time.Duration
}

// SessionSummary exposes one active session without its refresh credential.
type SessionSummary struct {
	SessionID string
	Subject   string
	Login     string
	Source    string
	ExpiresAt time.Time
}

// List returns active refresh-session metadata without revealing token hashes.
func (s *Store) List() []SessionSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	records := make([]SessionSummary, 0, len(s.records))
	for _, record := range s.records {
		if now.Before(record.ExpiresAt) {
			records = append(records, SessionSummary{SessionID: record.SessionID, Subject: record.Subject, Login: record.Login, Source: record.Source, ExpiresAt: record.ExpiresAt})
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].SessionID < records[j].SessionID })
	return records
}

// Store keeps active refresh-token records in memory.
type Store struct {
	mu      sync.Mutex
	now     func() time.Time
	options StoreOptions
	records map[string]RefreshRecord
	revoked map[string]time.Time
}

// NewStore constructs an empty refresh-token store.
//
// Parameters:
//   - now: clock used for expiry checks; time.Now is used when nil.
//   - options: context enforcement policy for refresh requests.
func NewStore(now func() time.Time, options StoreOptions) *Store {
	if now == nil {
		now = time.Now
	}

	return &Store{
		now:     now,
		options: options,
		records: make(map[string]RefreshRecord),
		revoked: make(map[string]time.Time),
	}
}

// Create creates a new refresh record and returns its raw token value.
func (s *Store) Create(input CreateInput) (RefreshToken, RefreshRecord, error) {
	if err := validateCreateInput(input); err != nil {
		return RefreshToken{}, RefreshRecord{}, err
	}

	tokenValue, err := newRefreshToken()
	if err != nil {
		return RefreshToken{}, RefreshRecord{}, err
	}

	record := RefreshRecord{
		TokenHash: hashToken(tokenValue),
		ExpiresAt: s.now().UTC().Add(input.TTL),
		SessionID: input.SessionID,
		Subject:   input.Subject,
		Login:     input.Login,
		Source:    input.Source,
		ClientIP:  input.Context.ClientIP,
		UserAgent: input.Context.UserAgent,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[record.TokenHash] = record

	return RefreshToken{Value: tokenValue, ExpiresAt: record.ExpiresAt}, record, nil
}

// Rotate consumes a valid refresh token and replaces it with a new opaque
// token while preserving the original absolute expiry.
func (s *Store) Rotate(token string, context RefreshContext) (RefreshToken, RefreshRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, err := s.requireCurrentRecordLocked(token, context)
	if err != nil {
		return RefreshToken{}, RefreshRecord{}, err
	}

	delete(s.records, hashToken(token))

	tokenValue, err := newRefreshToken()
	if err != nil {
		return RefreshToken{}, RefreshRecord{}, err
	}

	record.TokenHash = hashToken(tokenValue)
	record.ClientIP = context.ClientIP
	s.records[record.TokenHash] = record

	return RefreshToken{Value: tokenValue, ExpiresAt: record.ExpiresAt}, record, nil
}

// Revoke deletes a refresh token from active state.
func (s *Store) Revoke(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokenHash := hashToken(token)
	record, ok := s.records[tokenHash]
	if !ok {
		return ErrUnknownRefreshToken
	}

	delete(s.records, tokenHash)
	s.revokeLocked(record.SessionID)
	return nil
}

// revokeLocked records a bounded access-token denial while the store lock is held.
func (s *Store) revokeLocked(sessionID string) {
	ttl := s.options.RevocationTTL
	if ttl < time.Hour {
		ttl = time.Hour
	}
	s.revoked[sessionID] = s.now().UTC().Add(ttl)
}

// RevokeSession removes its refresh token and rejects its access tokens until they expire.
func (s *Store) RevokeSession(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	found := false
	for hash, record := range s.records {
		if record.SessionID == sessionID {
			delete(s.records, hash)
			found = true
		}
	}
	if found {
		s.revokeLocked(sessionID)
	}
	return found
}

// IsRevoked reports whether an access token's session remains in the deny list.
func (s *Store) IsRevoked(sessionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	expires, ok := s.revoked[sessionID]
	return ok && s.now().UTC().Before(expires)
}

// Clear removes every active refresh token from the store.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, record := range s.records {
		s.revokeLocked(record.SessionID)
	}
	s.records = make(map[string]RefreshRecord)
}

// CleanupExpired removes expired refresh records and returns the number of
// records removed.
func (s *Store) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now().UTC()
	removed := 0
	for tokenHash, record := range s.records {
		if now.Before(record.ExpiresAt) {
			continue
		}

		delete(s.records, tokenHash)
		removed++
	}
	for sessionID, expiry := range s.revoked {
		if !now.Before(expiry) {
			delete(s.revoked, sessionID)
		}
	}

	return removed
}

// Len returns the number of active refresh records.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.records)
}

func (s *Store) requireCurrentRecordLocked(token string, context RefreshContext) (RefreshRecord, error) {
	tokenHash := hashToken(token)
	record, ok := s.records[tokenHash]
	if !ok {
		return RefreshRecord{}, ErrUnknownRefreshToken
	}

	if !s.now().UTC().Before(record.ExpiresAt) {
		delete(s.records, tokenHash)
		return RefreshRecord{}, ErrExpiredRefreshToken
	}

	if !s.contextMatches(record, context) {
		delete(s.records, tokenHash)
		return RefreshRecord{}, ErrRefreshTokenContextMismatch
	}

	return record, nil
}

func (s *Store) contextMatches(record RefreshRecord, context RefreshContext) bool {
	if record.UserAgent != context.UserAgent {
		return false
	}

	return !s.options.EnforceClientIP || record.ClientIP == context.ClientIP
}

func validateCreateInput(input CreateInput) error {
	if input.SessionID == "" ||
		input.Subject == "" ||
		input.Login == "" ||
		input.Context.UserAgent == "" ||
		input.TTL <= 0 {
		return ErrInvalidRefreshSession
	}

	return nil
}
