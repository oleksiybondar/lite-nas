package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxMetadataBytes = 4 << 20

// uploadRecord is the durable authority for one resumable staged upload.
type uploadRecord struct {
	ID                 string     `json:"id"`
	Owner              string     `json:"owner"`
	UID                string     `json:"uid"`
	Target             string     `json:"target"`
	Filename           string     `json:"filename"`
	Size               int64      `json:"size"`
	SHA256             string     `json:"sha256"`
	ContentType        string     `json:"content_type,omitempty"`
	ChunkSize          int64      `json:"chunk_size"`
	RequestedChunkSize int64      `json:"requested_chunk_size,omitempty"`
	Offset             int64      `json:"offset"`
	NextIndex          int64      `json:"next_chunk_index"`
	StageDevice        uint64     `json:"stage_device"`
	StageInode         uint64     `json:"stage_inode"`
	StageUID           uint32     `json:"stage_uid"`
	DestinationExists  bool       `json:"destination_exists"`
	DestinationDevice  uint64     `json:"destination_device,omitempty"`
	DestinationInode   uint64     `json:"destination_inode,omitempty"`
	LastChunkSHA256    string     `json:"last_chunk_sha256,omitempty"`
	LastChunkSize      int64      `json:"last_chunk_size,omitempty"`
	IdempotencyKey     string     `json:"idempotency_key"`
	State              string     `json:"state"`
	Expires            time.Time  `json:"expires"`
	mu                 sync.Mutex `json:"-"`
}

// uploadStore maps opaque IDs and canonical destinations to durable records.
type uploadStore struct {
	directory string
	mu        sync.Mutex
	records   map[string]*uploadRecord
	targets   map[string]string
	keys      map[string]string
}

// newUploadStore opens the private durable directory and rebuilds its index.
func newUploadStore(directory string) (*uploadStore, error) {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return nil, err
	}
	store := &uploadStore{
		directory: directory, records: make(map[string]*uploadRecord),
		targets: make(map[string]string), keys: make(map[string]string),
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var total int64
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Size() > 4096 {
			return nil, errors.New("invalid upload metadata entry")
		}
		total += info.Size()
		if total > maxMetadataBytes {
			return nil, errors.New("upload metadata reservation exceeded")
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		var record uploadRecord
		if err := json.Unmarshal(data, &record); err != nil || record.ID+".json" != entry.Name() || record.Owner == "" || record.Target == "" {
			return nil, errors.New("invalid upload metadata")
		}
		if store.targets[record.Target] != "" {
			return nil, errors.New("duplicate active upload target")
		}
		store.add(&record)
	}
	return store, nil
}

// newUploadID creates an unpredictable URL-safe identity for a transfer.
func newUploadID() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// add indexes a record while the store mutex is held or during startup.
func (s *uploadStore) add(record *uploadRecord) {
	s.records[record.ID] = record
	if record.State != "completed" {
		s.targets[record.Target] = record.ID
	}
	s.keys[record.Owner+"\x00"+record.IdempotencyKey] = record.ID
}

// persist writes a complete record before replacing its old durable version.
func (s *uploadStore) persist(record *uploadRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if len(data) > 4096 {
		return errors.New("upload metadata record exceeds limit")
	}
	temporary, err := os.CreateTemp(s.directory, ".upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporary.Name(), filepath.Join(s.directory, record.ID+".json")); err != nil {
		return err
	}
	directory, err := os.Open(s.directory)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

// remove deletes one record after its stage has been safely removed.
func (s *uploadStore) remove(record *uploadRecord) error {
	if err := os.Remove(filepath.Join(s.directory, record.ID+".json")); err != nil {
		return err
	}
	directory, err := os.Open(s.directory)
	if err != nil {
		return err
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, record.ID)
	if s.targets[record.Target] == record.ID {
		delete(s.targets, record.Target)
	}
	delete(s.keys, record.Owner+"\x00"+record.IdempotencyKey)
	return nil
}

// find returns an indexed record without granting any authorization to use it.
func (s *uploadStore) find(id string) (*uploadRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[id]
	if record == nil {
		return nil, fmt.Errorf("unknown upload")
	}
	return record, nil
}
