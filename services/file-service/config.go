package main

import (
	"errors"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileaccess"
	"lite-nas/shared/fileio"
)

// serviceConfig describes only settings consumed by File Service.
type serviceConfig struct {
	Address        string
	Messaging      sharedconfig.MessagingConfig
	Logging        sharedconfig.LoggingConfig
	WriteRoots     []string
	ProtectedPaths []string
	MetadataDir    string
	ChunkSize      int64
	DocsEnabled    bool
	UploadTTL      time.Duration
	CleanupEvery   time.Duration
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
	writeRoots := configPaths(file.Section("filesystem").Key("write_roots").MustString("/home"))
	protected := configPaths(file.Section("filesystem").Key("protected_paths").String())
	for _, root := range append(append([]string(nil), writeRoots...), protected...) {
		if fileaccess.ValidateHostPath(root) != nil || filepath.Clean(root) != root {
			return serviceConfig{}, errors.New("filesystem roots must be clean absolute paths")
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
		MetadataDir: metadataDir, ChunkSize: int64(chunkMiB) << 20,
		DocsEnabled: docsEnabled,
		UploadTTL:   24 * time.Hour, CleanupEvery: 3 * time.Hour,
	}, nil
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
