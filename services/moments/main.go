package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lite-nas/shared/authtoken"
	"lite-nas/shared/browserauth"
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
	"lite-nas/shared/modules"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

const configPath = "/etc/lite-nas/moments.conf"

// serviceConfig holds the existing LiteNAS infrastructure settings for Moments.
type serviceConfig struct {
	HTTP         sharedconfig.HTTPConfig
	Messaging    sharedconfig.MessagingConfig
	AuthTokens   sharedconfig.AuthTokenConfig
	Logging      sharedconfig.LoggingConfig
	CookieDomain string
}

// loadConfig reads the service listener, messaging, and browser cookie scope.
func loadConfig(path string) (serviceConfig, error) {
	reader, err := fileio.NewFileReader(path)
	if err != nil {
		return serviceConfig{}, err
	}
	iniFile, err := sharedconfig.LoadINI(reader)
	if err != nil {
		return serviceConfig{}, err
	}
	authConfig, err := sharedconfig.LoadSharedAuthTokenConfig(iniFile)
	if err != nil {
		return serviceConfig{}, err
	}
	httpConfig, err := sharedconfig.LoadHTTPConfig(iniFile)
	if err != nil {
		return serviceConfig{}, err
	}
	if err := validateListener(httpConfig.Address); err != nil {
		return serviceConfig{}, err
	}
	domain := iniFile.Section("http").Key("cookie_domain").String()
	if err := browserauth.ValidateCookieDomain(domain); err != nil {
		return serviceConfig{}, err
	}
	return serviceConfig{HTTP: httpConfig, Messaging: authConfig.Messaging, AuthTokens: authConfig.AuthTokens, Logging: authConfig.Logging, CookieDomain: domain}, nil
}

// validateListener keeps the Moments API behind the local Nginx proxy.
func validateListener(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	if host != "127.0.0.1" && host != "::1" {
		return errors.New("Moments HTTP listener must be loopback")
	}
	return nil
}

// main starts the loopback Moments API and exits on startup or listener errors.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintf(os.Stderr, "moments: %v\n", err)
		os.Exit(1)
	}
}

// run assembles the shared auth client and serves the Moments skeleton API.
func run(ctx context.Context) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	certificate, err := os.ReadFile(cfg.AuthTokens.VerificationCert)
	if err != nil {
		return err
	}
	key, err := authtoken.ParseEd25519CertificatePublicKeyPEM(certificate)
	if err != nil {
		return err
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{Issuer: cfg.AuthTokens.Issuer, Audience: cfg.AuthTokens.Audience, ClockSkew: cfg.AuthTokens.ClockSkew}, key)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientInfra("lite-nas-moments", cfg.Logging, cfg.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           newRouter(browserauth.NewAuthService(infra.Client, verifier), cfg.CookieDomain),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	results := make(chan error, 1)
	go func() { results <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return ctx.Err()
	case err := <-results:
		return err
	}
}

// newRouter mounts the shared browser session flow for the Moments shell.
func newRouter(auth browserauth.AuthService, cookieDomain string) http.Handler {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("LiteNAS Moments", "0.1.0"))
	controller := authController{service: auth, cookieDomain: cookieDomain}
	huma.Post(api, "/api/auth/login", controller.login)
	huma.Post(api, "/api/auth/refresh", controller.refresh)
	huma.Post(api, "/api/auth/logout", controller.logout)
	huma.Get(api, "/api/auth/me", controller.me)
	return browserauth.SameOrigin(router)
}
