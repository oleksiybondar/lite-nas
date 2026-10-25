// Command stream-service serves authorized original media over loopback HTTP.
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

	"lite-nas/shared/fileaccess"
	"lite-nas/shared/fileio"
	"lite-nas/shared/modules"
)

const packagedConfigPath = "/etc/lite-nas/stream-service.conf"

// main runs the service until it receives a termination signal.
func main() {
	if len(os.Args) > 1 && os.Args[1] == "--read-file" {
		if err := runUserReader(os.Args[2:]); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "stream-service reader failed")
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--transcode-file" {
		if err := runUserTranscoder(os.Args[2:]); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "stream-service transcoder failed")
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--preview-file" {
		if err := runUserPreview(os.Args[2:]); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "stream-service preview failed")
			os.Exit(1)
		}
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, packagedConfigPath); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintf(os.Stderr, "stream-service: %v\n", err)
		os.Exit(1)
	}
}

// run sets up TLS-backed NATS clients and the loopback media server.
func run(ctx context.Context, configPath string) error {
	if os.Geteuid() == 0 {
		return errors.New("stream-service must run as a non-root user")
	}
	reader, err := fileio.NewFileReader(configPath)
	if err != nil {
		return err
	}
	cfg, err := loadServiceConfig(reader)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientInfra("stream-service", cfg.Logging, cfg.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	auth, err := fileaccess.NewAuthorizer(infra.Client)
	if err != nil {
		return err
	}
	service := &streamService{cfg: cfg, auth: auth, open: openAuthorizedFile, readAsUser: userReader, transcodes: make(chan struct{}, 2)}
	service.transcodeFile = service.transcode
	service.previewFile = service.preview
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
