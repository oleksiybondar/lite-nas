package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
	"lite-nas/services/web-gateway/routes"
	"lite-nas/services/web-gateway/services"
	"lite-nas/shared/authtoken"
	sharedconfig "lite-nas/shared/config"
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

	authVerifier, err := newAuthTokenVerifier(infra.Config.AuthTokens)
	if err != nil {
		return err
	}

	serviceModule := modules.NewServicesModule(infra.Client, authVerifier)
	fileModule, err := modules.NewFilesModule(packagedAssetRoot)
	if err != nil {
		return err
	}
	controllerModule := modules.NewControllersModule(fileModule.Static, infra.Logger, serviceModule)

	handler := routes.NewRouter(
		serviceName,
		apiVersion,
		controllerModule,
		middlewares.AuthenticationOptions{
			AccessCookieName:  services.AccessTokenCookieName,
			RefreshCookieName: services.RefreshTokenCookieName,
			Verifier:          authVerifier,
		},
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

func newAuthTokenVerifier(cfg sharedconfig.AuthTokenConfig) (authtoken.Verifier, error) {
	verificationCertData, err := os.ReadFile(cfg.VerificationCert) // #nosec G304 -- path comes from service config
	if err != nil {
		return authtoken.Verifier{}, err
	}

	verificationKey, err := authtoken.ParseEd25519CertificatePublicKeyPEM(verificationCertData)
	if err != nil {
		return authtoken.Verifier{}, err
	}

	return authtoken.NewVerifier(authTokenVerifierOptions(cfg), verificationKey)
}

func authTokenVerifierOptions(cfg sharedconfig.AuthTokenConfig) authtoken.VerifierOptions {
	return authtoken.VerifierOptions{
		Issuer:    cfg.Issuer,
		Audience:  cfg.Audience,
		ClockSkew: cfg.ClockSkew,
	}
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
