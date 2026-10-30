package backuprestore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Store owns the on-disk version archives for configured applications.
type Store struct {
	Root string
}

// BackedUpApplications lists application IDs with at least one completed version.
func (store Store) BackedUpApplications() ([]string, error) {
	entries, err := os.ReadDir(store.Root)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	applications := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() || !validApplication(entry.Name()) {
			continue
		}
		versions, err := store.Versions(entry.Name())
		if err != nil {
			return nil, err
		}
		if len(versions) > 0 {
			applications = append(applications, entry.Name())
		}
	}
	sort.Strings(applications)
	return applications, nil
}

// Backup writes a complete archive to a temporary file before publishing its version.
func (store Store) Backup(spec Spec) (string, error) {
	if !validApplication(spec.Application) {
		return "", fmt.Errorf("invalid application ID")
	}
	directory := filepath.Join(store.Root, spec.Application)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return "", err
	}
	temporary, err := os.CreateTemp(directory, ".snapshot-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(temporary.Name())
	defer temporary.Close()
	if err := temporary.Chmod(0o600); err != nil {
		return "", err
	}
	if err := WriteArchive(temporary, spec); err != nil {
		return "", err
	}
	if err := temporary.Sync(); err != nil {
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	version := time.Now().UTC().Format("20060102T150405.000000000Z")
	for suffix := 0; suffix < 100; suffix++ {
		candidate := version
		if suffix != 0 {
			candidate = fmt.Sprintf("%s-%d", version, suffix)
		}
		destination := filepath.Join(directory, candidate+".tar.gz")
		if err := os.Link(temporary.Name(), destination); err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", err
		}
		return candidate, nil
	}
	return "", fmt.Errorf("could not allocate backup version")
}

// Versions lists completed archive versions for one configured application.
func (store Store) Versions(application string) ([]string, error) {
	if !validApplication(application) {
		return nil, fmt.Errorf("invalid application ID")
	}
	entries, err := os.ReadDir(filepath.Join(store.Root, application))
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".tar.gz") {
			versions = append(versions, strings.TrimSuffix(entry.Name(), ".tar.gz"))
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return versions, nil
}

// OpenVersion opens only a named archive beneath the application's store root.
func (store Store) OpenVersion(application, version string) (io.ReadCloser, error) {
	if !validApplication(application) || version == "" || strings.Contains(version, "..") || strings.ContainsAny(version, "/\\") || strings.HasPrefix(version, "-") {
		return nil, fmt.Errorf("invalid version or application")
	}
	for _, char := range version {
		if !((char >= '0' && char <= '9') || char == 'T' || char == 'Z' || char == '.' || char == '-') {
			return nil, fmt.Errorf("invalid version")
		}
	}
	return os.Open(filepath.Join(store.Root, application, version+".tar.gz"))
}
