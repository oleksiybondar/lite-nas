package backuprestore

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"syscall"
)

// Difference reports one path that differs between a stored version and live state.
type Difference struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

// Diff compares a validated archive to current host files without exposing contents.
func (store Store) Diff(spec Spec, version string) ([]Difference, error) {
	archive, err := store.OpenVersion(spec.Application, version)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	stage, err := os.MkdirTemp(store.Root, ".diff-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stage)
	entries, err := stageArchive(archive, stage)
	if err != nil {
		return nil, err
	}
	if err := validateRestoreEntries(entries, spec); err != nil {
		return nil, err
	}
	if spec.Plugin == "ufw" || spec.Plugin == "tailscale" {
		var current []byte
		var err error
		virtualPath := ufwArchivePath
		if spec.Plugin == "ufw" {
			current, err = store.UFW.Export()
		} else {
			virtualPath = tailscaleArchivePath
			current, err = store.Tailscale.Export()
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.path == virtualPath {
				archived, err := os.ReadFile(entry.stage)
				if err != nil {
					return nil, err
				}
				if bytes.Equal(archived, current) {
					return []Difference{}, nil
				}
				return []Difference{{Path: virtualPath, Status: "changed"}}, nil
			}
		}
		return nil, fmt.Errorf("plugin JSON missing from snapshot")
	}
	present := make(map[string]bool, len(entries))
	differences := make([]Difference, 0)
	for _, entry := range entries {
		present[entry.path] = true
		status, err := compareEntry(entry)
		if err != nil {
			return nil, err
		}
		if status != "same" {
			differences = append(differences, Difference{Path: entry.path, Status: status})
		}
	}
	for _, selection := range spec.Selections {
		if !selection.Directory {
			continue
		}
		err := filepath.WalkDir(selection.Path, func(path string, _ os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !present[path] {
				differences = append(differences, Difference{Path: path, Status: "live-only"})
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	sort.Slice(differences, func(i, j int) bool { return differences[i].Path < differences[j].Path })
	return differences, nil
}

// compareEntry reports missing, changed, or same for a staged archive member.
func compareEntry(entry stagedEntry) (string, error) {
	info, err := os.Lstat(entry.path)
	if os.IsNotExist(err) {
		return "snapshot-only", nil
	}
	if err != nil {
		return "", err
	}
	switch entry.header.Typeflag {
	case tar.TypeDir:
		if info.IsDir() && sameMetadata(info, entry.header) {
			return "same", nil
		}
	case tar.TypeSymlink:
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(entry.path)
			if err != nil {
				return "", err
			}
			if target == entry.header.Linkname {
				return "same", nil
			}
		}
	case tar.TypeReg:
		if info.Mode().IsRegular() && info.Size() == entry.header.Size {
			archived, err := os.Open(entry.stage)
			if err != nil {
				return "", err
			}
			defer archived.Close()
			live, err := os.Open(entry.path)
			if err != nil {
				return "", err
			}
			defer live.Close()
			liveHash := sha256.New()
			if _, err := io.Copy(liveHash, live); err != nil {
				return "", err
			}
			archiveHash := sha256.New()
			if _, err := io.Copy(archiveHash, archived); err != nil {
				return "", err
			}
			if bytes.Equal(liveHash.Sum(nil), archiveHash.Sum(nil)) && sameMetadata(info, entry.header) {
				return "same", nil
			}
		}
	default:
		return "", fmt.Errorf("unsupported archive entry")
	}
	return "changed", nil
}

// sameMetadata compares restorable ownership, mode, and modification time.
func sameMetadata(info os.FileInfo, header *tar.Header) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	return info.Mode().Perm() == os.FileMode(header.Mode).Perm() &&
		int(stat.Uid) == header.Uid && int(stat.Gid) == header.Gid &&
		info.ModTime().Equal(header.ModTime)
}
