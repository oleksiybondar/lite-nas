package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	contract "lite-nas/shared/contracts/securityautomation"
)

const incidentStateDirectory = "/var/lib/lite-nas/security-automation"

var incidentUsername = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)

// incidentRecord durably identifies every UFW rule that an incident may own.
// Pending records are written before any command, so restart can reconcile them.
type incidentRecord struct {
	ID             string                   `json:"id"`
	RuleID         string                   `json:"rule_id"`
	ActionName     string                   `json:"action_name,omitempty"`
	Username       string                   `json:"username,omitempty"`
	TTY            string                   `json:"tty,omitempty"`
	DevicePort     string                   `json:"device_port,omitempty"`
	DeviceProduct  string                   `json:"device_product,omitempty"`
	DeviceHash     string                   `json:"device_hash,omitempty"`
	DeviceID       string                   `json:"device_id,omitempty"`
	SourceIP       string                   `json:"source_ip"`
	Interfaces     []string                 `json:"interfaces"`
	StartedAt      time.Time                `json:"started_at"`
	ExpiresAt      time.Time                `json:"expires_at"`
	State          contract.MitigationState `json:"state"`
	Outcome        string                   `json:"outcome"`
	AlertPublished bool                     `json:"alert_published"`
}

// incidentStore persists one bounded, root-owned-effect record per incident.
type incidentStore struct {
	directory string
	mutex     sync.Mutex
	records   map[string]incidentRecord
}

// openIncidentStore refuses malformed or unrecognized on-disk records before
// an effect worker can change host state.
func openIncidentStore(directory string) (*incidentStore, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	if len(entries) > 4096 {
		return nil, errors.New("incident store is full")
	}
	store := &incidentStore{directory: directory, records: make(map[string]incidentRecord, len(entries))}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".incident-") && entry.Type().IsRegular() {
			continue // An interrupted atomic write never owned a host effect.
		}
		if !entry.Type().IsRegular() || filepath.Ext(entry.Name()) != ".json" {
			return nil, fmt.Errorf("unexpected incident store entry %q", entry.Name())
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil || len(data) > 16*1024 {
			return nil, fmt.Errorf("cannot read incident %q", entry.Name())
		}
		var record incidentRecord
		if err := json.Unmarshal(data, &record); err != nil || entry.Name() != record.ID+".json" || validateIncidentRecord(record) != nil {
			return nil, fmt.Errorf("invalid incident record %q", entry.Name())
		}
		store.records[record.ID] = record
	}
	return store, nil
}

// validIncidentState excludes a normal state from the active effect journal.
func validIncidentState(state contract.MitigationState) bool {
	switch state {
	case contract.StatePending, contract.StateMitigating, contract.StateRestoring, contract.StateFailed:
		return true
	default:
		return false
	}
}

// validateIncidentRecord enforces the bounded on-disk effect identity contract.
func validateIncidentRecord(record incidentRecord) error {
	if !ufwIncidentID.MatchString(record.ID) || !automationID.MatchString(record.RuleID) ||
		record.StartedAt.IsZero() ||
		!record.ExpiresAt.After(record.StartedAt) || !validIncidentState(record.State) || len(record.Outcome) > 1024 {
		return errors.New("invalid incident fields")
	}
	if record.ActionName == "revoke_web_sessions" {
		if !incidentUsername.MatchString(record.Username) || len(record.Interfaces) != 0 || record.TTY != "" {
			return errors.New("invalid web-session incident target")
		}
		_, err := validateUFWSource(record.SourceIP)
		return err
	}
	if record.ActionName == "login_terminate_session" {
		if !incidentUsername.MatchString(record.Username) || !verifiedTTY.MatchString(record.TTY) || len(record.Interfaces) != 0 || record.SourceIP != "" {
			return errors.New("invalid login-session incident target")
		}
		return nil
	}
	if record.ActionName == "usb_block_device" {
		if !usbAlertPort.MatchString(record.DevicePort) || !usbProductID.MatchString(record.DeviceProduct) || record.DeviceHash != "" && !usbHash.MatchString(record.DeviceHash) || record.DeviceID != "" && !usbNumericID.MatchString(record.DeviceID) || record.SourceIP != "" || record.Username != "" || record.TTY != "" || len(record.Interfaces) != 0 {
			return errors.New("invalid USBGuard incident target")
		}
		return nil
	}
	if record.ActionName != "" && record.ActionName != "block_source_ip" && record.ActionName != "block_scan_source_ip" || len(record.Interfaces) == 0 || len(record.Interfaces) > 16 || record.Username != "" || record.TTY != "" || record.DevicePort != "" || record.DeviceProduct != "" || record.DeviceHash != "" || record.DeviceID != "" {
		return errors.New("invalid UFW incident target")
	}
	if _, err := validateUFWSource(record.SourceIP); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(record.Interfaces))
	for _, iface := range record.Interfaces {
		if !ufwInterfaceName.MatchString(iface) || iface[0] == '-' {
			return errors.New("invalid incident interface")
		}
		if _, exists := seen[iface]; exists {
			return errors.New("duplicate incident interface")
		}
		seen[iface] = struct{}{}
	}
	return nil
}

// newIncidentID returns an unguessable, alphanumeric ID suitable for an owned
// UFW comment and the existing incident RPC path contract.
func newIncidentID() (string, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(random[:]), nil
}

// save writes a complete record atomically before its state is exposed to RPCs.
func (store *incidentStore) save(record incidentRecord) error {
	if err := validateIncidentRecord(record); err != nil {
		return err
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return store.saveLocked(record)
}

// transition persists a state change only if another owner has not advanced
// the incident since the caller read it.
func (store *incidentStore) transition(record incidentRecord, expected contract.MitigationState) (bool, error) {
	if err := validateIncidentRecord(record); err != nil {
		return false, err
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	current, exists := store.records[record.ID]
	if !exists || current.State != expected {
		return false, nil
	}
	return true, store.saveLocked(record)
}

// saveLocked writes a validated record while the store mutex is held.
func (store *incidentStore) saveLocked(record incidentRecord) error {
	if len(store.records) >= 4096 && store.records[record.ID].ID == "" {
		return errors.New("incident store is full")
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(store.directory, ".incident-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), filepath.Join(store.directory, record.ID+".json")); err != nil {
		return err
	}
	store.records[record.ID] = record
	directory, err := os.Open(store.directory)
	if err != nil {
		return err
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return err
	}
	return nil
}

// list returns a stable snapshot of incident records for RPC reads and expiry.
func (store *incidentStore) list() []incidentRecord {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	result := make([]incidentRecord, 0, len(store.records))
	for _, record := range store.records {
		result = append(result, record)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt.Before(result[j].StartedAt) })
	return result
}

// get returns one immutable record snapshot.
func (store *incidentStore) get(id string) (incidentRecord, bool) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	record, ok := store.records[id]
	return record, ok
}

// delete removes only a verified restored incident and syncs the directory.
func (store *incidentStore) delete(id string) error {
	if !ufwIncidentID.MatchString(id) {
		return errors.New("invalid incident ID")
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if _, ok := store.records[id]; !ok {
		return errors.New("unknown incident")
	}
	if err := os.Remove(filepath.Join(store.directory, id+".json")); err != nil {
		return err
	}
	delete(store.records, id)
	directory, err := os.Open(store.directory)
	if err != nil {
		return err
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return err
	}
	return nil
}
