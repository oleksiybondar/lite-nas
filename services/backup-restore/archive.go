package backuprestore

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

// WriteArchive captures selected files, full directory trees, and file symlink targets.
// Entries use absolute host paths without the leading slash.
func WriteArchive(writer io.Writer, spec Spec) error {
	gzipWriter := gzip.NewWriter(writer)
	tarWriter := tar.NewWriter(gzipWriter)
	seen := make(map[string]bool)
	var totalSize int64
	for _, selection := range spec.Selections {
		if err := archivePath(tarWriter, selection.Path, selection.Directory, seen, &totalSize); err != nil {
			_ = tarWriter.Close()
			_ = gzipWriter.Close()
			return err
		}
	}
	if spec.SourcePath != "" {
		if err := archivePath(tarWriter, spec.SourcePath, false, seen, &totalSize); err != nil {
			_ = tarWriter.Close()
			_ = gzipWriter.Close()
			return err
		}
	}
	if err := tarWriter.Close(); err != nil {
		_ = gzipWriter.Close()
		return err
	}
	return gzipWriter.Close()
}

// archivePath captures a configured path without following directory symlinks.
func archivePath(writer *tar.Writer, path string, recursive bool, seen map[string]bool, totalSize *int64) error {
	if seen[path] {
		return nil
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.IsDir() != recursive && (recursive || !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0) {
		return fmt.Errorf("unexpected file type at %s", path)
	}
	if !info.IsDir() && !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("unsupported file type at %s", path)
	}
	if info.Mode().IsRegular() {
		if info.Size() < 0 || info.Size() > maxArchiveFileSize || *totalSize > maxArchiveTotalSize-info.Size() {
			return fmt.Errorf("source size limit exceeded at %s", path)
		}
		*totalSize += info.Size()
	}
	link := ""
	if info.Mode()&os.ModeSymlink != 0 {
		link, err = os.Readlink(path)
		if err != nil {
			return err
		}
	}
	header, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return err
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		header.Uid = int(stat.Uid)
		header.Gid = int(stat.Gid)
	}
	header.Name = strings.TrimPrefix(path, "/")
	header.Format = tar.FormatPAX
	if err := writer.WriteHeader(header); err != nil {
		return err
	}
	seen[path] = true
	if info.Mode().IsRegular() {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}
	if info.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}
		resolvedInfo, err := os.Stat(resolved)
		if err != nil || !resolvedInfo.Mode().IsRegular() {
			return fmt.Errorf("symlink target is not a regular file: %s", path)
		}
		next := link
		if !filepath.IsAbs(next) {
			next = filepath.Join(filepath.Dir(path), next)
		}
		return archivePath(writer, filepath.Clean(next), false, seen, totalSize)
	}
	if recursive {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := archivePath(writer, filepath.Join(path, entry.Name()), entry.IsDir(), seen, totalSize); err != nil {
				return err
			}
		}
	}
	return nil
}

// ArchivePaths lists archive paths without extracting or writing host files.
func ArchivePaths(reader io.Reader) ([]string, error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	var paths []string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			if _, err := io.Copy(io.Discard, gzipReader); err != nil {
				return nil, err
			}
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Name == "" || header.Name == "." || header.Name == ".." || filepath.IsAbs(header.Name) || filepath.Clean(header.Name) != header.Name || strings.HasPrefix(header.Name, "../") {
			return nil, fmt.Errorf("unsafe archive path %q", header.Name)
		}
		paths = append(paths, header.Name)
	}
	sort.Strings(paths)
	return paths, nil
}
