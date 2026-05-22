package modules

import (
	"context"
	"time"

	serviceconfig "lite-nas/services/resources-monitor/config"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedfileio "lite-nas/shared/fileio"
	"lite-nas/shared/messaging"
	sharedmodules "lite-nas/shared/modules"
)

const defaultAuthRefreshInterval = 24 * time.Hour

// Infra groups runtime infrastructure dependencies for resources-monitor.
type Infra struct {
	sharedmodules.CoreClientAuthInfra
	Config serviceconfig.Config
}

// NewInfraModule loads service config and constructs shared runtime
// infrastructure.
func NewInfraModule(configPath string, serviceName string) (Infra, error) {
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return Infra{}, err
	}

	cfg, err := serviceconfig.LoadConfig(cfgReader)
	if err != nil {
		return Infra{}, err
	}

	authServiceName := cfg.Auth.ServiceName
	if authServiceName == "" {
		authServiceName = serviceName
	}
	cfg.Auth.ServiceName = authServiceName

	core, err := sharedmodules.NewCoreClientAuthInfra(
		serviceName,
		cfg.Logging,
		cfg.Messaging,
		cfg.Auth,
		defaultAuthRefreshInterval,
	)
	if err != nil {
		return Infra{}, err
	}

	server, err := messaging.NewServer(cfg.Messaging, core.Logger, messaging.NewJSONCodec())
	if err != nil {
		core.Close()
		return Infra{}, err
	}
	core.Server = server
	core.Client = authTokenClient{
		client:       core.Client,
		tokenManager: core.AuthTokenManager,
	}

	return Infra{
		CoreClientAuthInfra: core,
		Config:              cfg,
	}, nil
}

type authTokenClient struct {
	client       messaging.Client
	tokenManager interface {
		Token() (string, time.Time, error)
		Refresh(context.Context) error
		Login(context.Context) error
	}
}

func (client authTokenClient) Publish(ctx context.Context, subject string, payload any) error {
	accessToken, err := client.currentToken(ctx)
	if err != nil {
		return err
	}
	return client.client.Publish(ctx, subject, withAccessToken(payload, accessToken))
}

func (client authTokenClient) Request(ctx context.Context, subject string, request any, response any) error {
	accessToken, err := client.currentToken(ctx)
	if err != nil {
		return err
	}
	return client.client.Request(ctx, subject, withAccessToken(request, accessToken), response)
}

func (client authTokenClient) Drain() error {
	return client.client.Drain()
}

func (client authTokenClient) Close() {
	client.client.Close()
}

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
