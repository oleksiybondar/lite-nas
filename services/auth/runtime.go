package main

import (
	"context"
	"os"
	"time"

	"lite-nas/services/auth/modules"
	"lite-nas/services/auth/sessions"
	"lite-nas/shared/authtoken"
	sharedconfig "lite-nas/shared/config"
)

const (
	packagedConfigPath = "/etc/lite-nas/auth.conf"
	serviceName        = "auth-service"
	pamServiceName     = "litenas-auth"
)

// run assembles the auth-service runtime and keeps the process alive until
// shutdown while the service contract surface is still being built out.
//
// Parameters:
//   - ctx: process-lifetime context cancelled by OS signal handling
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	tokenRuntime, err := newAuthTokenRuntime(infra.Config.AuthTokens)
	if err != nil {
		return err
	}

	authModule, err := modules.NewAuthModule(pamServiceName)
	if err != nil {
		return err
	}

	runtimeDeps := authRuntime{
		Auth:         authModule,
		Tokens:       tokenRuntime,
		RefreshStore: newRefreshStore(infra.Config.AuthTokens),
	}

	infra.Logger.Info(
		"auth service started",
		"config", packagedConfigPath,
		"pam_service", runtimeDeps.Auth.ServiceName,
		"token_runtime_ready", true,
		"refresh_sessions", runtimeDeps.RefreshStore.Len(),
	)

	<-ctx.Done()

	infra.Logger.Info("auth service stopping")
	return ctx.Err()
}

type authRuntime struct {
	Auth         modules.Auth
	Tokens       authTokenRuntime
	RefreshStore *sessions.Store
}

type authTokenRuntime struct {
	Issuer   authtoken.Issuer
	Verifier authtoken.Verifier
}

func newAuthTokenRuntime(cfg sharedconfig.AuthTokenConfig) (authTokenRuntime, error) {
	issuer, err := newAuthTokenIssuer(cfg)
	if err != nil {
		return authTokenRuntime{}, err
	}

	verifier, err := newAuthTokenVerifier(cfg)
	if err != nil {
		return authTokenRuntime{}, err
	}

	return authTokenRuntime{
		Issuer:   issuer,
		Verifier: verifier,
	}, nil
}

func newAuthTokenIssuer(cfg sharedconfig.AuthTokenConfig) (authtoken.Issuer, error) {
	signingKeyData, err := os.ReadFile(cfg.SigningKey) // #nosec G304 -- path comes from service config
	if err != nil {
		return authtoken.Issuer{}, err
	}

	signingKey, err := authtoken.ParseEd25519PrivateKeyPEM(signingKeyData)
	if err != nil {
		return authtoken.Issuer{}, err
	}

	return authtoken.NewIssuer(issuerOptions(cfg), signingKey)
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

	return authtoken.NewVerifier(verifierOptions(cfg), verificationKey)
}

func issuerOptions(cfg sharedconfig.AuthTokenConfig) authtoken.IssuerOptions {
	return authtoken.IssuerOptions{
		Issuer:         cfg.Issuer,
		Audience:       cfg.Audience,
		AccessLifetime: cfg.AccessLifetime,
	}
}

func verifierOptions(cfg sharedconfig.AuthTokenConfig) authtoken.VerifierOptions {
	return authtoken.VerifierOptions{
		Issuer:    cfg.Issuer,
		Audience:  cfg.Audience,
		ClockSkew: cfg.ClockSkew,
	}
}

func newRefreshStore(cfg sharedconfig.AuthTokenConfig) *sessions.Store {
	return sessions.NewStore(
		time.Now,
		sessions.StoreOptions{EnforceClientIP: cfg.EnforceRefreshClientIP},
	)
}
