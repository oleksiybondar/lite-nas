package backuprestore

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
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
	return store.BackupWithMetadata(spec, "", "", "")
}

// BackupWithMetadata creates a snapshot attributed to a verified caller.
func (store Store) BackupWithMetadata(spec Spec, name, description, createdBy string) (string, error) {
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
	created := time.Now().UTC()
	version := created.Format("20060102T150405.000000000Z")
	for suffix := 0; suffix < 100; suffix++ {
		candidate := version
		if suffix != 0 {
			candidate = fmt.Sprintf("%s-%d", version, suffix)
		}
		destination := filepath.Join(directory, candidate+".tar.gz")
		if _, err := os.Stat(destination); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return "", err
		}
		temporary, err := os.CreateTemp(directory, ".snapshot-*")
		if err != nil {
			return "", err
		}
		metadata := Metadata{Application: spec.Application, Version: candidate, RestartUnits: spec.RestartUnits, Name: name, Description: description, CreatedAt: created.Format(time.RFC3339Nano), CreatedBy: createdBy}
		writeErr := writeArchive(temporary, spec, &metadata)
		if writeErr == nil {
			writeErr = temporary.Sync()
		}
		closeErr := temporary.Close()
		if writeErr == nil {
			writeErr = closeErr
		}
		if writeErr != nil {
			_ = os.Remove(temporary.Name())
			return "", writeErr
		}
		if err := os.Link(temporary.Name(), destination); err != nil {
			_ = os.Remove(temporary.Name())
			if os.IsExist(err) {
				continue
			}
			return "", err
		}
		_ = os.Remove(temporary.Name())
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

// Metadata reads the validated metadata header embedded in a completed archive.
func (store Store) Metadata(application, version string) (Metadata, error) {
	archive, err := store.OpenVersion(application, version)
	if err != nil {
		return Metadata{}, err
	}
	defer archive.Close()
	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return Metadata{}, err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	header, err := tarReader.Next()
	if err != nil {
		return Metadata{}, err
	}
	if header.Name != "metadata.json" || header.Size < 0 || header.Size > 16*1024 {
		return Metadata{}, fmt.Errorf("invalid archive metadata entry")
	}
	var metadata Metadata
	if err := json.NewDecoder(tarReader).Decode(&metadata); err != nil {
		return Metadata{}, err
	}
	if metadata.Application != application || metadata.Version != version || metadata.CreatedAt == "" {
		return Metadata{}, fmt.Errorf("archive metadata does not match its version")
	}
	return metadata, nil
}

// Delete removes one validated completed archive and leaves its application directory intact.
func (store Store) Delete(application, version string) error {
	path := filepath.Join(store.Root, application, version+".tar.gz")
	archive, err := store.OpenVersion(application, version)
	if err != nil {
		return err
	}
	if err := archive.Close(); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("backup version is not a regular archive")
	}
	return os.Remove(path)
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
