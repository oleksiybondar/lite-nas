package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"lite-nas/shared/loggingmanager/query"
)

const eventSequenceDirectory = "/var/lib/lite-nas/operational-safety-automation/event-ids"

// eventIDAllocator persists only ID sequences, never tasks or effect state.
type eventIDAllocator struct {
	directory string
	mu        sync.Mutex
}

// newEventIDAllocator binds generated alert IDs to a service-owned directory.
func newEventIDAllocator(directory string) (*eventIDAllocator, error) {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	return &eventIDAllocator{directory: directory}, nil
}

// Next reserves a unique generated ID before the alert can be published.
func (allocator *eventIDAllocator) Next(prefix string) (string, error) {
	if !outputPrefixPattern.MatchString(prefix) {
		return "", errors.New("invalid output ID prefix")
	}
	allocator.mu.Lock()
	defer allocator.mu.Unlock()
	path := filepath.Join(allocator.directory, prefix)
	sequence := uint64(0)
	data, err := os.ReadFile(path)
	if err == nil {
		sequence, err = strconv.ParseUint(strings.TrimSpace(string(data)), 10, 32)
		if err != nil {
			return "", fmt.Errorf("invalid saved event sequence for %s: %w", prefix, err)
		}
		sequence++
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if sequence > query.EventIDMaxSequence {
		return "", errors.New("event ID sequence exhausted")
	}
	id, err := query.BuildEventID(prefix, uint32(sequence))
	if err != nil {
		return "", err
	}
	temporary, err := os.CreateTemp(allocator.directory, ".event-id-*")
	if err != nil {
		return "", err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return "", err
	}
	if _, err := temporary.WriteString(strconv.FormatUint(sequence, 10) + "\n"); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return "", err
	}
	directory, err := os.Open(allocator.directory)
	if err != nil {
		return "", err
	}
	if err := directory.Sync(); err != nil {
		_ = directory.Close()
		return "", err
	}
	if err := directory.Close(); err != nil {
		return "", err
	}
	return id, nil
}
