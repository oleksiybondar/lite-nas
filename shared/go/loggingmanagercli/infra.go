package loggingmanagercli

import (
	"context"
	"time"

	sharedconfig "lite-nas/shared/config"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedfileio "lite-nas/shared/fileio"
	"lite-nas/shared/loggingmanagerauth"
	sharedmodules "lite-nas/shared/modules"
)

const defaultAuthRefreshInterval = 24 * time.Hour

// LoadInfra constructs messaging client infra for a logging-manager CLI app.
func LoadInfra(configPath string, appName string) (func(), MessagingClient, error) {
	cfg, err := loadConfig(configPath, appName)
	if err != nil {
		return nil, nil, err
	}

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

	if err := loginTokenManager(core); err != nil {
		return nil, nil, err
	}

	return core.Close, loggingmanagerauth.NewClient(core.Client, core.AuthTokenManager, withAccessToken), nil
}

func loadConfig(configPath string, appName string) (sharedconfig.SharedConfig, error) {
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return sharedconfig.SharedConfig{}, err
	}

	cfgFile, err := sharedconfig.LoadINI(cfgReader)
	if err != nil {
		return sharedconfig.SharedConfig{}, err
	}

	cfg, err := sharedconfig.LoadSharedConfig(cfgFile)
	if err != nil {
		return sharedconfig.SharedConfig{}, err
	}

	authServiceName := cfg.Auth.ServiceName
	if authServiceName == "" {
		authServiceName = appName
	}
	cfg.Auth.ServiceName = authServiceName
	return cfg, nil
}

func loginTokenManager(core sharedmodules.CoreClientAuthInfra) error {
	if err := core.AuthTokenManager.Login(context.Background()); err != nil {
		core.Close()
		return err
	}
	return nil
}

func withAccessToken(payload any, accessToken string) any {
	if updated, ok := injectAccessTokenMutationPayload(payload, accessToken); ok {
		return updated
	}
	if updated, ok := injectAccessTokenReadPayload(payload, accessToken); ok {
		return updated
	}
	return payload
}

func injectAccessTokenMutationPayload(payload any, accessToken string) (any, bool) {
	switch typed := payload.(type) {
	case loggingmanagercontract.AlertPayload:
		typed.AccessToken = accessToken
		return typed, true
	case loggingmanagercontract.AlertOccurrencePayload:
		typed.AccessToken = accessToken
		return typed, true
	case loggingmanagercontract.UpdateAlertStateInput:
		typed.AccessToken = accessToken
		return typed, true
	case loggingmanagercontract.AcknowledgeAlertInput:
		typed.AccessToken = accessToken
		return typed, true
	case loggingmanagercontract.MuteAlertInput:
		typed.AccessToken = accessToken
		return typed, true
	default:
		return nil, false
	}
}

func injectAccessTokenReadPayload(payload any, accessToken string) (any, bool) {
	switch typed := payload.(type) {
	case loggingmanagercontract.ListAlertsInput:
		typed.AccessToken = accessToken
		return typed, true
	case loggingmanagercontract.GetAlertInput:
		typed.AccessToken = accessToken
		return typed, true
	default:
		return nil, false
	}
}
