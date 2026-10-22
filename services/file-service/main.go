// Command file-service serves authenticated file operations on a loopback HTTP port.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lite-nas/services/file-service/filesystem"
	"lite-nas/shared/fileaccess"
	"lite-nas/shared/fileio"
	"lite-nas/shared/modules"
)

const packagedConfigPath = "/etc/lite-nas/file-service.conf"

// main runs the service until an interrupt or termination signal arrives.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, packagedConfigPath); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintf(os.Stderr, "file-service: %v\n", err)
		os.Exit(1)
	}
}

// run builds the shared transport, transfer store, and HTTP server.
func run(ctx context.Context, configPath string) error {
	if os.Geteuid() == 0 {
		return errors.New("file-service must run as a non-root user")
	}
	reader, err := fileio.NewFileReader(configPath)
	if err != nil {
		return err
	}
	cfg, err := loadServiceConfig(reader)
	if err != nil {
		return err
	}
	policy, err := filesystem.NewPolicy(cfg.WriteRoots, cfg.ProtectedPaths)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientInfra("file-service", cfg.Logging, cfg.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	auth, err := fileaccess.NewAuthorizer(infra.Client)
	if err != nil {
		return err
	}
	service, err := newFileService(cfg, policy, auth)
	if err != nil {
		return err
	}
	service.logger = infra.Logger
	go service.cleanupLoop(ctx)
	server := &http.Server{
		Addr: cfg.Address, Handler: service,
		ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second,
	}
	stopped := make(chan error, 1)
	go func() { stopped <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			return err
		}
		return ctx.Err()
	case err := <-stopped:
		return err
	}
}
