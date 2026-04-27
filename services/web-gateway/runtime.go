package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"lite-nas/services/web-gateway/modules"
	"lite-nas/services/web-gateway/routes"
)

const (
	packagedConfigPath  = "/etc/lite-nas/web-gateway.conf"
	packagedAssetRoot   = "/usr/share/lite-nas/web-gateway/assets"
	serviceName         = "web-gateway"
	apiVersion          = "0.1.0"
	httpReadTimeout     = 10 * time.Second
	httpReadHeaderLimit = 5 * time.Second
	httpWriteTimeout    = 15 * time.Second
	httpIdleTimeout     = 30 * time.Second
	httpShutdownTimeout = 10 * time.Second
)

// run assembles the gateway runtime, starts the HTTP server, and blocks until
// the process shuts down.
//
// Parameters:
//   - ctx: process-lifetime context cancelled by OS signal handling
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	serviceModule := modules.NewServicesModule(infra.Client)
	fileModule, err := modules.NewFilesModule(packagedAssetRoot)
	if err != nil {
		return err
	}
	controllerModule := modules.NewControllersModule(fileModule.Static, infra.Logger, serviceModule)

	handler := routes.NewRouter(
		serviceName,
		apiVersion,
		controllerModule,
	)

	server := &http.Server{
		Addr:              infra.Config.HTTP.Address,
		Handler:           handler,
		ReadHeaderTimeout: httpReadHeaderLimit,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		IdleTimeout:       httpIdleTimeout,
	}

	return serveHTTP(ctx, server, infra.Logger)
}

// serveHTTP runs the HTTP server and coordinates graceful shutdown.
//
// Parameters:
//   - ctx: cancellation source used to trigger graceful shutdown
//   - server: configured HTTP server instance to run
//   - log: logger used for lifecycle messages and shutdown failures
func serveHTTP(
	ctx context.Context,
	server *http.Server,
	log interface {
		Info(string, ...any)
		Error(string, ...any)
	},
) error {
	errCh := make(chan error, 1)

	go runHTTPServer(server, log, errCh)

	select {
	case <-ctx.Done():
		if err := shutdownHTTPServer(server, log); err != nil {
			return err
		}

		if err := <-errCh; err != nil {
			return err
		}

		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func runHTTPServer(
	server *http.Server,
	log interface {
		Info(string, ...any)
		Error(string, ...any)
	},
	errCh chan<- error,
) {
	log.Info("web gateway started", "address", server.Addr)
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		errCh <- err
		return
	}

	errCh <- nil
}

func shutdownHTTPServer(
	server *http.Server,
	log interface {
		Info(string, ...any)
		Error(string, ...any)
	},
) error {
	log.Info("web gateway stopping")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}
