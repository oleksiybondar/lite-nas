package main

import (
	"errors"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileaccess"
	"lite-nas/shared/fileio"
)

// serviceConfig contains the transport and storage scope needed by Stream Service.
type serviceConfig struct {
	Address     string
	Messaging   sharedconfig.MessagingConfig
	Logging     sharedconfig.LoggingConfig
	ReadRoots   []string
	DeniedPaths []string
}

// loadServiceConfig validates the configuration before the HTTP listener starts.
func loadServiceConfig(reader fileio.Reader) (serviceConfig, error) {
	file, err := sharedconfig.LoadINI(reader)
	if err != nil {
		return serviceConfig{}, err
	}
	messaging, err := sharedconfig.LoadMessagingConfig(file)
	if err != nil {
		return serviceConfig{}, err
	}
	if !strings.HasPrefix(messaging.URL, "tls://") || messaging.ClientName != "stream-service" ||
		messaging.Timeout <= 0 || !filepath.IsAbs(messaging.CA) ||
		!filepath.IsAbs(messaging.Cert) || !filepath.IsAbs(messaging.Key) {
		return serviceConfig{}, errors.New("stream-service requires TLS NATS credentials")
	}
	logging, err := sharedconfig.LoadLoggingConfig(file)
	if err != nil {
		return serviceConfig{}, err
	}
	address := file.Section("http").Key("address").String()
	host, portText, err := net.SplitHostPort(address)
	port, portErr := strconv.Atoi(portText)
	if err != nil || portErr != nil || port < 1 || port > 65535 ||
		net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		return serviceConfig{}, errors.New("HTTP address must bind to a loopback IP and port")
	}
	paths := strings.Split(file.Section("filesystem").Key("read_roots").String(), ",")
	roots := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if fileaccess.ValidateHostPath(path) != nil || path == "/" {
			return serviceConfig{}, errors.New("read_roots must contain clean absolute storage paths")
		}
		roots = append(roots, path)
	}
	denied := strings.Split(file.Section("filesystem").Key("denied_paths").String(), ",")
	for i, path := range denied {
		path = strings.TrimSpace(path)
		if fileaccess.ValidateHostPath(path) != nil || path == "/" {
			return serviceConfig{}, errors.New("denied_paths must contain clean absolute paths")
		}
		denied[i] = path
	}
	return serviceConfig{Address: address, Messaging: messaging, Logging: logging, ReadRoots: roots, DeniedPaths: denied}, nil
}
