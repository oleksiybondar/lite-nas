package loggingmanagercli

import (
	"context"
	"time"

	sharedconfig "lite-nas/shared/config"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedfileio "lite-nas/shared/fileio"
	sharedmodules "lite-nas/shared/modules"
)

const defaultAuthRefreshInterval = 24 * time.Hour

// LoadInfra constructs messaging client infra for a logging-manager CLI app.
func LoadInfra(configPath string, appName string) (func(), MessagingClient, error) {
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return nil, nil, err
	}

	cfgFile, err := sharedconfig.LoadINI(cfgReader)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := sharedconfig.LoadSharedConfig(cfgFile)
	if err != nil {
		return nil, nil, err
	}

	authServiceName := cfg.Auth.ServiceName
	if authServiceName == "" {
		authServiceName = appName
	}
	cfg.Auth.ServiceName = authServiceName

	core, err := sharedmodules.NewCoreClientAuthInfra(
		appName,
		cfg.Logging,
		cfg.Messaging,
		cfg.Auth,
		defaultAuthRefreshInterval,
	)
	if err != nil {
		return nil, nil, err
	}

	if err := core.AuthTokenManager.Login(context.Background()); err != nil {
		core.Close()
		return nil, nil, err
	}

	return core.Close, authTokenClient{client: core.Client, tokenManager: core.AuthTokenManager}, nil
}

type authTokenClient struct {
	client       MessagingClient
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
	case loggingmanagercontract.ListAlertsInput:
		typed.AccessToken = accessToken
		return typed
	case loggingmanagercontract.GetAlertInput:
		typed.AccessToken = accessToken
		return typed
	case loggingmanagercontract.UpdateAlertStateInput:
		typed.AccessToken = accessToken
		return typed
	case loggingmanagercontract.AcknowledgeAlertInput:
		typed.AccessToken = accessToken
		return typed
	case loggingmanagercontract.MuteAlertInput:
		typed.AccessToken = accessToken
		return typed
	default:
		return payload
	}
}
