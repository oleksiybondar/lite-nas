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
)

// RestoreMode determines whether absent entries under selected directories are removed.
type RestoreMode string

const (
	// ReplaceMode makes selected directory trees match the archive.
	ReplaceMode RestoreMode = "replace"
	// MergeMode keeps live entries absent from the archive.
	MergeMode RestoreMode = "merge"
)

// stagedEntry records one validated archive member before any host mutation.
type stagedEntry struct {
	header *tar.Header
	path   string
	stage  string
}

const (
	maxArchiveFileSize  int64 = 128 << 20
	maxArchiveTotalSize int64 = 1 << 30
)

// Restore applies a completed archive only to paths owned by the current spec.
func (store Store) Restore(spec Spec, version string, mode RestoreMode) error {
	if mode != ReplaceMode && mode != MergeMode {
		return fmt.Errorf("invalid restore mode")
	}
	archive, err := store.OpenVersion(spec.Application, version)
	if err != nil {
		return err
	}
	defer archive.Close()
	stage, err := os.MkdirTemp(store.Root, ".restore-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	entries, err := stageArchive(archive, stage)
	if err != nil {
		return err
	}
	if err := validateRestoreEntries(entries, spec); err != nil {
		return err
	}
	for _, entry := range entries {
		if err := rejectSymlinkParents(entry.path); err != nil {
			return err
		}
	}
	for _, selection := range spec.Selections {
		if !selection.Directory {
			continue
		}
		if info, err := os.Lstat(selection.Path); err == nil && !info.IsDir() {
			return fmt.Errorf("selected directory has another live type: %s", selection.Path)
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if mode == ReplaceMode {
		if err := removeAbsentDirectoryEntries(entries, spec); err != nil {
			return err
		}
	}
	return applyEntries(entries)
}

// stageArchive fully validates and copies regular content into private staging.
func stageArchive(source io.Reader, stage string) ([]stagedEntry, error) {
	gzipReader, err := gzip.NewReader(source)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	entries := make([]stagedEntry, 0)
	seen := make(map[string]bool)
	var totalSize int64
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			if _, err := io.Copy(io.Discard, gzipReader); err != nil {
				return nil, err
			}
			return entries, nil
		}
		if err != nil {
			return nil, err
		}
		if header.Name == "" || filepath.IsAbs(header.Name) || filepath.Clean(header.Name) != header.Name || header.Name == ".." || strings.HasPrefix(header.Name, "../") {
			return nil, fmt.Errorf("unsafe archive path %q", header.Name)
		}
		path := "/" + header.Name
		if seen[path] {
			return nil, fmt.Errorf("duplicate archive path %s", path)
		}
		seen[path] = true
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir && header.Typeflag != tar.TypeSymlink {
			return nil, fmt.Errorf("unsupported archive type at %s", path)
		}
		if header.Size < 0 || header.Size > maxArchiveFileSize || totalSize > maxArchiveTotalSize-header.Size {
			return nil, fmt.Errorf("archive size limit exceeded")
		}
		totalSize += header.Size
		if header.Name == "metadata.json" {
			if header.Typeflag != tar.TypeReg || header.Size > 4096 {
				return nil, fmt.Errorf("invalid archive metadata")
			}
			var metadata Metadata
			if err := json.NewDecoder(tarReader).Decode(&metadata); err != nil {
				return nil, err
			}
			if metadata.Application == "" || metadata.Version == "" || metadata.CreatedAt == "" {
				return nil, fmt.Errorf("incomplete archive metadata")
			}
			continue
		}
		staged := stagedEntry{header: header, path: path}
		if header.Typeflag == tar.TypeReg {
			file, err := os.CreateTemp(stage, "file-*")
			if err != nil {
				return nil, err
			}
			staged.stage = file.Name()
			_, copyErr := io.Copy(file, tarReader)
			closeErr := file.Close()
			if copyErr != nil {
				return nil, copyErr
			}
			if closeErr != nil {
				return nil, closeErr
			}
		}
		entries = append(entries, staged)
	}
}

// selectedPath checks the current spec; a selected directory owns descendants.
func selectedPath(spec Spec, path string) bool {
	if path == spec.SourcePath {
		return true
	}
	for _, selection := range spec.Selections {
		if path == selection.Path || selection.Directory && strings.HasPrefix(path, selection.Path+"/") {
			return true
		}
	}
	return false
}

// validateRestoreEntries rejects symlink targets outside the staged selection.
func validateRestoreEntries(entries []stagedEntry, spec Spec) error {
	files := make(map[string]bool)
	linkedTargets := make(map[string]bool)
	links := make(map[string]string)
	present := make(map[string]bool)
	types := make(map[string]byte)
	for _, entry := range entries {
		present[entry.path] = true
		types[entry.path] = entry.header.Typeflag
		if entry.header.Typeflag == tar.TypeReg {
			files[entry.path] = true
		} else if entry.header.Typeflag == tar.TypeSymlink {
			target := entry.header.Linkname
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(entry.path), target)
			}
			links[entry.path] = filepath.Clean(target)
		}
	}
	for _, entry := range entries {
		for parent := filepath.Dir(entry.path); parent != "/"; parent = filepath.Dir(parent) {
			if _, linked := links[parent]; linked {
				return fmt.Errorf("archive member has symlink parent: %s", entry.path)
			}
		}
		if entry.header.Typeflag != tar.TypeSymlink {
			continue
		}
		target := links[entry.path]
		if !selectedPath(spec, entry.path) && !linkedTargets[entry.path] {
			return fmt.Errorf("symlink target is not a selected regular file: %s", entry.path)
		}
		for steps := 0; steps <= len(links); steps++ {
			linkedTargets[target] = true
			if files[target] {
				break
			}
			next, ok := links[target]
			if !ok || steps == len(links) {
				return fmt.Errorf("symlink chain does not end at an archived file: %s", entry.path)
			}
			target = next
		}
	}
	for _, entry := range entries {
		if !selectedPath(spec, entry.path) && !linkedTargets[entry.path] {
			return fmt.Errorf("unselected archive path %s", entry.path)
		}
	}
	for _, selection := range spec.Selections {
		if !present[selection.Path] {
			return fmt.Errorf("selected path missing from archive: %s", selection.Path)
		}
		if selection.Directory && types[selection.Path] != tar.TypeDir || !selection.Directory && types[selection.Path] != tar.TypeReg && types[selection.Path] != tar.TypeSymlink {
			return fmt.Errorf("selected path has wrong archive type: %s", selection.Path)
		}
	}
	if spec.SourcePath != "" && types[spec.SourcePath] != tar.TypeReg {
		return fmt.Errorf("spec missing from archive")
	}
	return nil
}

// removeAbsentDirectoryEntries removes only entries beneath explicit directory roots.
func removeAbsentDirectoryEntries(entries []stagedEntry, spec Spec) error {
	present := make(map[string]bool, len(entries))
	for _, entry := range entries {
		present[entry.path] = true
	}
	for _, selection := range spec.Selections {
		if !selection.Directory {
			continue
		}
		var absent []string
		err := filepath.WalkDir(selection.Path, func(path string, _ os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path != selection.Path && !present[path] {
				absent = append(absent, path)
			}
			return nil
		})
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		sort.Slice(absent, func(i, j int) bool { return len(absent[i]) > len(absent[j]) })
		for _, path := range absent {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

// applyEntries restores regular files, links, then directory metadata.
func applyEntries(entries []stagedEntry) error {
	for _, entry := range entries {
		if err := rejectSymlinkParents(entry.path); err != nil {
			return err
		}
		if entry.header.Typeflag == tar.TypeDir {
			if info, err := os.Lstat(entry.path); err == nil && !info.IsDir() {
				return fmt.Errorf("restore directory path has another type: %s", entry.path)
			} else if err != nil && !os.IsNotExist(err) {
				return err
			}
			if err := os.MkdirAll(entry.path, 0o700); err != nil {
				return err
			}
		}
	}
	for _, entry := range entries {
		if entry.header.Typeflag != tar.TypeReg {
			continue
		}
		if err := restoreRegular(entry); err != nil {
			return err
		}
	}
	for _, entry := range entries {
		if entry.header.Typeflag != tar.TypeSymlink {
			continue
		}
		if err := os.Remove(entry.path); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := os.Symlink(entry.header.Linkname, entry.path); err != nil {
			return err
		}
		if err := os.Lchown(entry.path, entry.header.Uid, entry.header.Gid); err != nil {
			return err
		}
	}
	for _, entry := range entries {
		if entry.header.Typeflag == tar.TypeDir {
			if err := os.Chown(entry.path, entry.header.Uid, entry.header.Gid); err != nil {
				return err
			}
			if err := os.Chmod(entry.path, os.FileMode(entry.header.Mode)); err != nil {
				return err
			}
			if err := os.Chtimes(entry.path, entry.header.ModTime, entry.header.ModTime); err != nil {
				return err
			}
		}
	}
	return nil
}

// rejectSymlinkParents prevents an archive path from escaping through a live parent link.
func rejectSymlinkParents(path string) error {
	for parent := filepath.Dir(path); parent != "/" && parent != "."; parent = filepath.Dir(parent) {
		info, err := os.Lstat(parent)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink parent is not allowed: %s", parent)
		}
	}
	return nil
}

// restoreRegular writes a file through a temporary sibling and preserves metadata.
func restoreRegular(entry stagedEntry) error {
	if err := os.MkdirAll(filepath.Dir(entry.path), 0o700); err != nil {
		return err
	}
	input, err := os.Open(entry.stage)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.CreateTemp(filepath.Dir(entry.path), ".restore-*")
	if err != nil {
		return err
	}
	defer os.Remove(output.Name())
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return err
	}
	if err := output.Chown(entry.header.Uid, entry.header.Gid); err != nil {
		_ = output.Close()
		return err
	}
	if err := output.Chmod(os.FileMode(entry.header.Mode)); err != nil {
		_ = output.Close()
		return err
	}
	if err := output.Sync(); err != nil {
		_ = output.Close()
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	if err := os.Rename(output.Name(), entry.path); err != nil {
		return err
	}
	return os.Chtimes(entry.path, entry.header.ModTime, entry.header.ModTime)
}
