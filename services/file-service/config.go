package main

import (
	"errors"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"lite-nas/services/file-service/filesystem"
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileaccess"
	"lite-nas/shared/fileio"
)

// serviceConfig describes only settings consumed by File Service.
type serviceConfig struct {
	Address             string
	Messaging           sharedconfig.MessagingConfig
	Logging             sharedconfig.LoggingConfig
	WriteRoots          []string
	ProtectedPaths      []string
	ReadRules           filesystem.ReadRules
	ListingIgnoredNames []string
	MetadataDir         string
	ChunkSize           int64
	DocsEnabled         bool
	UploadTTL           time.Duration
	CleanupEvery        time.Duration
}

// loadServiceConfig validates the service's INI boundary before runtime setup.
func loadServiceConfig(reader fileio.Reader) (serviceConfig, error) {
	file, err := sharedconfig.LoadINI(reader)
	if err != nil {
		return serviceConfig{}, err
	}
	messaging, err := sharedconfig.LoadMessagingConfig(file)
	if err != nil {
		return serviceConfig{}, err
	}
	logging, err := sharedconfig.LoadLoggingConfig(file)
	if err != nil {
		return serviceConfig{}, err
	}
	address := file.Section("http").Key("address").MustString("127.0.0.1:8082")
	host, portText, err := net.SplitHostPort(address)
	port, portErr := strconv.Atoi(portText)
	if err != nil || portErr != nil || port < 1 || port > 65535 ||
		net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		return serviceConfig{}, errors.New("HTTP address must bind to a loopback IP and port")
	}
	chunkMiB := 1
	if key, err := file.Section("uploads").GetKey("chunk_size_mib"); err == nil {
		chunkMiB, err = key.Int()
		if err != nil {
			return serviceConfig{}, errors.New("chunk_size_mib must be an integer")
		}
	}
	if chunkMiB < 1 || chunkMiB > 16 {
		return serviceConfig{}, errors.New("chunk_size_mib must be between 1 and 16")
	}
	metadataDir := file.Section("uploads").Key("metadata_dir").MustString("/var/lib/lite-nas/file-service/uploads")
	if fileaccess.ValidateHostPath(metadataDir) != nil {
		return serviceConfig{}, errors.New("metadata_dir must be a clean absolute path")
	}
	writeRoots := configPaths(file.Section("filesystem").Key("write_roots").String())
	readRoots := configPaths(file.Section("filesystem").Key("read_roots").String())
	protected := configPaths(file.Section("filesystem").Key("protected_paths").String())
	readRules := filesystem.ReadRules{
		NonAdminDenied:       configPaths(file.Section("filesystem").Key("non_admin_denied_paths").String()),
		AllDenied:            configPaths(file.Section("filesystem").Key("all_denied_paths").String()),
		KnownPathDirectories: configPaths(file.Section("filesystem").Key("known_path_directories").String()),
	}
	listingIgnoredNames, err := parseListingIgnoredNames(file.Section("filesystem").Key("listing_ignored_names").String())
	if err != nil {
		return serviceConfig{}, err
	}
	if len(writeRoots) == 0 {
		return serviceConfig{}, errors.New("filesystem.write_roots is required")
	}
	readRules.ReadRoots = readRoots
	for _, root := range append(append(append([]string(nil), writeRoots...), readRoots...), protected...) {
		if fileaccess.ValidateHostPath(root) != nil || filepath.Clean(root) != root {
			return serviceConfig{}, errors.New("filesystem roots must be clean absolute paths")
		}
	}
	for _, pattern := range append(append(append([]string{}, readRules.NonAdminDenied...), readRules.AllDenied...), readRules.KnownPathDirectories...) {
		if err := filesystem.ValidatePathMask(pattern); err != nil {
			return serviceConfig{}, err
		}
	}
	docsEnabled := true
	if key, err := file.Section("api_docs").GetKey("enabled"); err == nil {
		docsEnabled, err = key.Bool()
		if err != nil {
			return serviceConfig{}, errors.New("api_docs.enabled must be boolean")
		}
	}
	return serviceConfig{
		Address: address, Messaging: messaging, Logging: logging,
		WriteRoots: writeRoots, ProtectedPaths: protected,
		ReadRules: readRules, ListingIgnoredNames: listingIgnoredNames,
		MetadataDir: metadataDir, ChunkSize: int64(chunkMiB) << 20,
		DocsEnabled: docsEnabled,
		UploadTTL:   24 * time.Hour, CleanupEvery: 3 * time.Hour,
	}, nil
}

// parseListingIgnoredNames validates comma-separated basename patterns for directory listings.
func parseListingIgnoredNames(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	patterns := make([]string, 0)
	for _, part := range strings.Split(value, ",") {
		pattern := strings.TrimSpace(part)
		if pattern == "" || pattern == "." || pattern == ".." || len(pattern) > 255 ||
			strings.ContainsAny(pattern, "/\\") {
			return nil, errors.New("listing_ignored_names must contain valid basename patterns")
		}
		if _, err := filepath.Match(pattern, ""); err != nil {
			return nil, errors.New("listing_ignored_names contains an invalid pattern")
		}
		patterns = append(patterns, strings.ToLower(pattern))
	}
	return patterns, nil
}

// configPaths parses the comma-separated path convention used in INI values.
func configPaths(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	paths := make([]string, 0)
	for _, value := range strings.Split(value, ",") {
		paths = append(paths, strings.TrimSpace(value))
	}
	return paths
}
