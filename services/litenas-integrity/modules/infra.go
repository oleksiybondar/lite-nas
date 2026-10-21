package modules

import (
	"context"
	"time"

	serviceconfig "lite-nas/services/litenas-integrity/config"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedfileio "lite-nas/shared/fileio"
	"lite-nas/shared/messaging"
	sharedmodules "lite-nas/shared/modules"
)

const defaultAuthRefreshInterval = 24 * time.Hour

// Infra groups runtime infrastructure dependencies for the integrity service.
type Infra struct {
	sharedmodules.CoreClientAuthInfra
	Config serviceconfig.Config
}

// NewInfraModule loads service config and constructs shared runtime
// infrastructure.
func NewInfraModule(configPath string, serviceName string) (Infra, error) {
	cfg, err := loadConfigWithDefaultAuthService(configPath, serviceName)
	if err != nil {
		return Infra{}, err
	}

	core, err := buildCoreAuthInfra(serviceName, cfg)
	if err != nil {
		return Infra{}, err
	}

	return Infra{
		CoreClientAuthInfra: core,
		Config:              cfg,
	}, nil
}

// loadConfigWithDefaultAuthService ensures service-token login defaults to the
// current service identity when the config omits an override.
func loadConfigWithDefaultAuthService(configPath string, serviceName string) (serviceconfig.Config, error) {
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return serviceconfig.Config{}, err
	}

	cfg, err := serviceconfig.LoadConfig(cfgReader)
	if err != nil {
		return serviceconfig.Config{}, err
	}

	authServiceName := cfg.Auth.ServiceName
	if authServiceName == "" {
		authServiceName = serviceName
	}
	cfg.Auth.ServiceName = authServiceName

	return cfg, nil
}

// buildCoreAuthInfra wires the shared logger, messaging client/server, and
// service-token manager for the integrity service.
func buildCoreAuthInfra(serviceName string, cfg serviceconfig.Config) (sharedmodules.CoreClientAuthInfra, error) {
	core, err := sharedmodules.NewCoreClientAuthInfra(
		serviceName,
		cfg.Logging,
		cfg.Messaging,
		cfg.Auth,
		defaultAuthRefreshInterval,
	)
	if err != nil {
		return sharedmodules.CoreClientAuthInfra{}, err
	}

	server, err := messaging.NewServer(cfg.Messaging, core.Logger, messaging.NewJSONCodec())
	if err != nil {
		core.Close()
		return sharedmodules.CoreClientAuthInfra{}, err
	}

	core.Server = server
	core.Client = authTokenClient{
		client:       core.Client,
		tokenManager: core.AuthTokenManager,
	}

	return core, nil
}

// authTokenClient injects the current service token into alert payloads while
// preserving the underlying generic messaging interface.
type authTokenClient struct {
	client       messaging.Client
	tokenManager interface {
		Token() (string, time.Time, error)
		Refresh(context.Context) error
		Login(context.Context) error
	}
}

// Publish adds the current service access token to supported alert payloads.
func (client authTokenClient) Publish(ctx context.Context, subject string, payload any) error {
	accessToken, err := client.currentToken(ctx)
	if err != nil {
		return err
	}
	return client.client.Publish(ctx, subject, withAccessToken(payload, accessToken))
}

// Request adds the current service access token to supported logging-manager
// mutation requests.
func (client authTokenClient) Request(ctx context.Context, subject string, request any, response any) error {
	accessToken, err := client.currentToken(ctx)
	if err != nil {
		return err
	}
	return client.client.Request(ctx, subject, withAccessToken(request, accessToken), response)
}

// Drain gracefully drains the wrapped client connection.
func (client authTokenClient) Drain() error {
	return client.client.Drain()
}

// Close immediately closes the wrapped client connection.
func (client authTokenClient) Close() {
	client.client.Close()
}

// currentToken returns the active service access token, refreshing or logging
// in when needed.
func (client authTokenClient) currentToken(ctx context.Context) (string, error) {
	accessToken, _, err := client.tokenManager.Token()
	if err == nil {
		return accessToken, nil
	}
	if refreshErr := client.tokenManager.Refresh(ctx); refreshErr == nil {
		accessToken, _, err = client.tokenManager.Token()
		if err == nil {
			return accessToken, nil
		}
	}
	if loginErr := client.tokenManager.Login(ctx); loginErr != nil {
		return "", loginErr
	}
	accessToken, _, err = client.tokenManager.Token()
	return accessToken, err
}

// withAccessToken copies supported alert contracts and injects the current
// service token into their auth field.
func withAccessToken(payload any, accessToken string) any {
	switch typed := payload.(type) {
	case loggingmanagercontract.AlertPayload:
		typed.AccessToken = accessToken
		return typed
	case loggingmanagercontract.AlertOccurrencePayload:
		typed.AccessToken = accessToken
		return typed
	case loggingmanagercontract.UpdateAlertStateInput:
		typed.AccessToken = accessToken
		return typed
	default:
		return payload
	}
}
