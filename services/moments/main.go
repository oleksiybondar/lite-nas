package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lite-nas/shared/authtoken"
	"lite-nas/shared/browserauth"
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileaccess"
	"lite-nas/shared/fileio"
	"lite-nas/shared/modules"
	"lite-nas/shared/swaggerui"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

const configPath = "/etc/lite-nas/moments.conf"

// serviceConfig holds the existing LiteNAS infrastructure settings for Moments.
type serviceConfig struct {
	HTTP           sharedconfig.HTTPConfig
	Messaging      sharedconfig.MessagingConfig
	AuthTokens     sharedconfig.AuthTokenConfig
	Logging        sharedconfig.LoggingConfig
	CookieDomain   string
	DatabasePath   string
	FileServiceURL string
	ThumbnailDir   string
	WorkingDir     string
	PlaceDataPath  string
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
	databasePath := iniFile.Section("storage").Key("database_path").MustString("/var/lib/lite-nas/moments/moments.db")
	if fileaccess.ValidateHostPath(databasePath) != nil || databasePath == "/" {
		return serviceConfig{}, errors.New("Moments database path must be clean and absolute")
	}
	fileServiceURL := iniFile.Section("storage").Key("file_service_url").MustString("http://127.0.0.1:8082")
	thumbnailDir := iniFile.Section("storage").Key("thumbnail_dir").MustString("/var/lib/lite-nas/moments/thumbnails")
	if fileaccess.ValidateHostPath(thumbnailDir) != nil || thumbnailDir == "/" {
		return serviceConfig{}, errors.New("thumbnail_dir must be clean and absolute")
	}
	workingDir := iniFile.Section("storage").Key("working_dir").MustString("/home")
	if fileaccess.ValidateHostPath(workingDir) != nil {
		return serviceConfig{}, errors.New("working_dir must be clean and absolute")
	}
	placeDataPath := iniFile.Section("storage").Key("place_data_path").MustString("/usr/share/lite-nas/moments/cities1000.zip")
	if fileaccess.ValidateHostPath(placeDataPath) != nil {
		return serviceConfig{}, errors.New("place_data_path must be clean and absolute")
	}
	parsed, err := url.Parse(fileServiceURL)
	if err != nil || parsed.Scheme != "http" || parsed.Path != "" || parsed.RawQuery != "" || parsed.User != nil || parsed.Fragment != "" {
		return serviceConfig{}, errors.New("file_service_url must be a loopback HTTP origin")
	}
	if err := validateListener(parsed.Host); err != nil {
		return serviceConfig{}, errors.New("file_service_url must be a loopback HTTP origin")
	}
	return serviceConfig{HTTP: httpConfig, Messaging: authConfig.Messaging, AuthTokens: authConfig.AuthTokens, Logging: authConfig.Logging, CookieDomain: domain, DatabasePath: databasePath, FileServiceURL: fileServiceURL, ThumbnailDir: thumbnailDir, WorkingDir: workingDir, PlaceDataPath: placeDataPath}, nil
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
	if err := run(ctx, configPath); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintf(os.Stderr, "moments: %v\n", err)
		os.Exit(1)
	}
}

// run assembles the shared auth client and serves the Moments skeleton API.
func run(ctx context.Context, path string) error {
	cfg, err := loadConfig(path)
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
	store, err := openGalleryStore(cfg.DatabasePath)
	if err != nil {
		return err
	}
	defer store.close()
	places, err := loadCityIndex(cfg.PlaceDataPath)
	if err != nil {
		return fmt.Errorf("load offline city data: %w", err)
	}
	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           newRouterWithRuntime(browserauth.NewAuthService(infra.Client, verifier), cfg.CookieDomain, store, newFileServiceClient(cfg.FileServiceURL), cfg.ThumbnailDir, cfg.WorkingDir, infra.Client, places),
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
	return newRouterWithStore(auth, cookieDomain, nil)
}

// newRouterWithStore mounts authenticated gallery metadata alongside the session API.
func newRouterWithStore(auth browserauth.AuthService, cookieDomain string, store *galleryStore) http.Handler {
	return newRouterWithDependencies(auth, cookieDomain, store, "http://127.0.0.1:8082")
}

// newRouterWithDependencies binds the metadata store and internal File Service origin.
func newRouterWithDependencies(auth browserauth.AuthService, cookieDomain string, store *galleryStore, fileServiceURL string) http.Handler {
	return newRouterWithClient(auth, cookieDomain, store, newFileServiceClient(fileServiceURL))
}

// newRouterWithClient keeps File Service transport injectable at the HTTP boundary.
func newRouterWithClient(auth browserauth.AuthService, cookieDomain string, store *galleryStore, files fileServiceClient) http.Handler {
	return newRouterWithRuntime(auth, cookieDomain, store, files, "", "/home", nil)
}

// newRouterWithRuntime mounts API routes with provisioned thumbnails and picker configuration.
func newRouterWithRuntime(auth browserauth.AuthService, cookieDomain string, store *galleryStore, files fileServiceClient, thumbnailDir, workingDir string, users subjectLookup, places ...*cityIndex) http.Handler {
	router := chi.NewMux()
	config := huma.DefaultConfig("LiteNAS Moments", "0.1.0")
	config.OpenAPIPath = "/api/openapi"
	config.DocsPath = ""
	api := humachi.New(router, config)
	router.Get("/api/docs", func(writer http.ResponseWriter, request *http.Request) {
		swaggerui.Serve(writer, request, "/api", "LiteNAS Moments API", "/api/openapi.json")
	})
	router.Get("/api/docs/*", func(writer http.ResponseWriter, request *http.Request) {
		swaggerui.Serve(writer, request, "/api", "LiteNAS Moments API", "/api/openapi.json")
	})
	controller := authController{service: auth, cookieDomain: cookieDomain}
	huma.Post(api, "/api/auth/login", controller.login)
	huma.Post(api, "/api/auth/refresh", controller.refresh)
	huma.Post(api, "/api/auth/logout", controller.logout)
	huma.Get(api, "/api/auth/me", controller.me)
	huma.Get(api, "/api/config", controller.config(workingDir))
	if store != nil {
		mountGalleryRoutes(api, galleryController{auth: auth, store: store, files: files, scans: newScanManager(store, auth, files, places...), thumbnailDir: thumbnailDir, users: users})
	}
	return browserauth.SameOrigin(router)
}
