package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lite-nas/shared/fileaccess"
)

const (
	captionGrantStatePath = "/var/lib/lite-nas-auth/caption-grants.json"
	captionGrantLifetime  = 7 * 24 * time.Hour
)

// captionGrantRecord holds only the hash of an opaque File Service credential.
type captionGrantRecord struct {
	JobID      string    `json:"job_id"`
	Username   string    `json:"username"`
	ExpiresAt  time.Time `json:"expires_at"`
	PathHashes []string  `json:"path_hashes"`
}

// captionGrantStore keeps job grants durable across Auth Service restarts.
type captionGrantStore struct {
	mu      sync.Mutex
	path    string
	records map[string]captionGrantRecord
	now     func() time.Time
}

// openCaptionGrantStore loads the protected state file or starts an empty store.
func openCaptionGrantStore(path string, now func() time.Time) (*captionGrantStore, error) {
	if now == nil {
		now = time.Now
	}
	store := &captionGrantStore{path: path, records: make(map[string]captionGrantRecord), now: now}
	data, err := os.ReadFile(path) // #nosec G304 -- fixed service state path or test fixture
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read caption grants: %w", err)
	}
	if err := json.Unmarshal(data, &store.records); err != nil || store.records == nil {
		return nil, errors.New("invalid caption grant state")
	}
	return store, nil
}

// issue writes one exact-path grant before returning its bearer value.
func (s *captionGrantStore) issue(jobID, username string, pathHashes []string) (string, time.Time, error) {
	if s == nil || jobID == "" || username == "" || len(pathHashes) == 0 || len(pathHashes) > 10000 {
		return "", time.Time{}, errors.New("invalid caption grant")
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, err
	}
	grant := base64.RawURLEncoding.EncodeToString(bytes)
	digest := sha256.Sum256([]byte(grant))
	key := hex.EncodeToString(digest[:])
	expiresAt := s.now().UTC().Add(captionGrantLifetime)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, record := range s.records {
		if record.JobID == jobID {
			return "", time.Time{}, errors.New("caption job already has a grant")
		}
	}
	next := make(map[string]captionGrantRecord, len(s.records)+1)
	for existingKey, record := range s.records {
		if s.now().Before(record.ExpiresAt) {
			next[existingKey] = record
		}
	}
	next[key] = captionGrantRecord{JobID: jobID, Username: username, ExpiresAt: expiresAt, PathHashes: append([]string(nil), pathHashes...)}
	if err := s.write(next); err != nil {
		return "", time.Time{}, err
	}
	s.records = next
	return grant, expiresAt, nil
}

// validate allows only a listed path while the grant is active.
func (s *captionGrantStore) validate(grant, path string) string {
	if s == nil || len(grant) != 43 || fileaccess.ValidateHostPath(path) != nil {
		return ""
	}
	digest := sha256.Sum256([]byte(grant))
	key := hex.EncodeToString(digest[:])
	pathDigest := sha256.Sum256([]byte(path))
	pathKey := hex.EncodeToString(pathDigest[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.records[key]
	if !found || !s.now().Before(record.ExpiresAt) {
		return ""
	}
	for _, allowed := range record.PathHashes {
		if allowed == pathKey {
			return record.Username
		}
	}
	return ""
}

// revoke removes a grant durably before reporting success.
func (s *captionGrantStore) revoke(grant string) (bool, error) {
	if s == nil || len(grant) != 43 {
		return false, nil
	}
	digest := sha256.Sum256([]byte(grant))
	key := hex.EncodeToString(digest[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, found := s.records[key]; !found {
		return false, nil
	}
	next := make(map[string]captionGrantRecord, len(s.records)-1)
	for existingKey, record := range s.records {
		if existingKey != key {
			next[existingKey] = record
		}
	}
	if err := s.write(next); err != nil {
		return false, err
	}
	s.records = next
	return true, nil
}

// write atomically replaces protected grant state before acknowledging a change.
func (s *captionGrantStore) write(records map[string]captionGrantRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(s.path), ".caption-grants-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), s.path)
}
