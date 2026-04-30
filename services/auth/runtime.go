package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"time"

	"lite-nas/services/auth/modules"
	"lite-nas/services/auth/pamauth"
	"lite-nas/services/auth/sessions"
	"lite-nas/shared/authtoken"
	sharedconfig "lite-nas/shared/config"
	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
)

const (
	packagedConfigPath = "/etc/lite-nas/auth.conf"
	serviceName        = "auth-service"
	pamServiceName     = "litenas-auth"
	refreshTokenTTL    = 24 * time.Hour
	sessionIDBytes     = 16
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
	if err := registerRPCHandlers(infra.Server, runtimeDeps); err != nil {
		return err
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

func registerRPCHandlers(server messaging.Server, runtimeDeps authRuntime) error {
	if err := server.RegisterRPC(authcontract.LoginRPCSubject, func(_ context.Context, envelope messaging.Envelope) (any, error) {
		return handleLoginRPC(runtimeDeps, envelope)
	}); err != nil {
		return err
	}

	if err := server.RegisterRPC(authcontract.RefreshRPCSubject, func(_ context.Context, envelope messaging.Envelope) (any, error) {
		return handleRefreshRPC(runtimeDeps, envelope)
	}); err != nil {
		return err
	}

	if err := server.RegisterRPC(authcontract.LogoutRPCSubject, func(_ context.Context, envelope messaging.Envelope) (any, error) {
		return handleLogoutRPC(runtimeDeps, envelope)
	}); err != nil {
		return err
	}

	if err := server.RegisterRPC(authcontract.ValidateAccessTokenRPCSubject, func(_ context.Context, envelope messaging.Envelope) (any, error) {
		return handleValidateAccessTokenRPC(runtimeDeps, envelope)
	}); err != nil {
		return err
	}

	return nil
}

func handleLoginRPC(runtimeDeps authRuntime, envelope messaging.Envelope) (authcontract.LoginResponse, error) {
	var request authcontract.LoginRequest
	if !decodeRPCRequest(envelope, &request) {
		return authcontract.LoginResponse{Status: authcontract.StatusDenied}, nil
	}

	result, authenticated := authenticatePrincipal(runtimeDeps, request)
	if !authenticated {
		return loginResponseFromPAM(result), nil
	}

	accessToken, issued := issueAccessToken(runtimeDeps, result.Username, result.Username)
	if !issued {
		return authcontract.LoginResponse{Status: authcontract.StatusDenied}, nil
	}

	refreshToken, created := createRefreshToken(runtimeDeps, request, result.Username)
	if !created {
		return authcontract.LoginResponse{Status: authcontract.StatusDenied}, nil
	}

	response := loginResponseFromPAM(result)
	response.AccessToken = accessToken
	response.RefreshToken = refreshToken
	return response, nil
}

func authenticatePrincipal(runtimeDeps authRuntime, request authcontract.LoginRequest) (pamauth.Result, bool) {
	result, err := runtimeDeps.Auth.Authenticator.Authenticate(pamauth.AuthenticateRequest{
		Username: request.Username,
		Password: request.Password,
	})
	if err != nil {
		return result, false
	}
	if result.Code != pamauth.OutcomeAuthenticated {
		return result, false
	}

	return result, true
}

func issueAccessToken(runtimeDeps authRuntime, subject string, login string) (string, bool) {
	accessToken, _, err := runtimeDeps.Tokens.Issuer.Issue(authtoken.Principal{
		Subject: subject,
		Login:   login,
	})
	if err != nil {
		return "", false
	}

	return accessToken, true
}

func createRefreshToken(runtimeDeps authRuntime, request authcontract.LoginRequest, username string) (string, bool) {
	refreshToken, _, err := runtimeDeps.RefreshStore.Create(sessions.CreateInput{
		SessionID: newSessionID(),
		Subject:   username,
		Login:     username,
		Context:   refreshContext(request.ClientIP, request.UserAgent),
		TTL:       refreshTokenTTL,
	})
	if err != nil {
		return "", false
	}

	return refreshToken.Value, true
}

func handleRefreshRPC(runtimeDeps authRuntime, envelope messaging.Envelope) (authcontract.RefreshResponse, error) {
	var request authcontract.RefreshRequest
	if !decodeRPCRequest(envelope, &request) {
		return authcontract.RefreshResponse{}, nil
	}

	refreshToken, record, rotated := rotateRefreshToken(runtimeDeps, request)
	if !rotated {
		return authcontract.RefreshResponse{}, nil
	}

	accessToken, issued := issueAccessToken(runtimeDeps, record.Subject, record.Login)
	if !issued {
		return authcontract.RefreshResponse{}, nil
	}

	return authcontract.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Value,
	}, nil
}

func rotateRefreshToken(
	runtimeDeps authRuntime,
	request authcontract.RefreshRequest,
) (sessions.RefreshToken, sessions.RefreshRecord, bool) {
	refreshToken, record, err := runtimeDeps.RefreshStore.Rotate(request.RefreshToken, refreshContext(request.ClientIP, request.UserAgent))
	if err != nil {
		return sessions.RefreshToken{}, sessions.RefreshRecord{}, false
	}

	return refreshToken, record, true
}

func handleLogoutRPC(runtimeDeps authRuntime, envelope messaging.Envelope) (authcontract.LogoutResponse, error) {
	var request authcontract.LogoutRequest
	if !decodeRPCRequest(envelope, &request) {
		return authcontract.LogoutResponse{}, nil
	}

	if !revokeRefreshToken(runtimeDeps, request.RefreshToken) {
		return authcontract.LogoutResponse{}, nil
	}

	return authcontract.LogoutResponse{LoggedOut: true}, nil
}

func revokeRefreshToken(runtimeDeps authRuntime, refreshToken string) bool {
	return runtimeDeps.RefreshStore.Revoke(refreshToken) == nil
}

func handleValidateAccessTokenRPC(runtimeDeps authRuntime, envelope messaging.Envelope) (authcontract.ValidateAccessTokenResponse, error) {
	var request authcontract.ValidateAccessTokenRequest
	if !decodeRPCRequest(envelope, &request) {
		return authcontract.ValidateAccessTokenResponse{Valid: false, Status: authcontract.StatusDenied}, nil
	}

	claims, verified := verifyAccessToken(runtimeDeps, request.AccessToken)
	if !verified {
		return authcontract.ValidateAccessTokenResponse{Valid: false, Status: authcontract.StatusDenied}, nil
	}

	return authcontract.ValidateAccessTokenResponse{
		Valid:    true,
		Status:   authcontract.StatusAuthenticated,
		Username: claims.Login,
	}, nil
}

func verifyAccessToken(runtimeDeps authRuntime, accessToken string) (authtoken.AccessClaims, bool) {
	claims, err := runtimeDeps.Tokens.Verifier.Verify(accessToken)
	if err != nil {
		return authtoken.AccessClaims{}, false
	}

	return claims, true
}

func decodeRPCRequest(envelope messaging.Envelope, request any) bool {
	return json.Unmarshal(envelope.Payload, request) == nil
}

func loginResponseFromPAM(result pamauth.Result) authcontract.LoginResponse {
	return authcontract.LoginResponse{
		Status:            authStatusFromPAM(result.Code),
		Username:          result.Username,
		Messages:          authMessagesFromPAM(result.Messages),
		CanChangePassword: result.CanChangePassword,
	}
}

func authStatusFromPAM(outcome pamauth.OutcomeCode) authcontract.Status {
	switch outcome {
	case pamauth.OutcomeAuthenticated:
		return authcontract.StatusAuthenticated
	case pamauth.OutcomePasswordChangeNeeded:
		return authcontract.StatusPasswordChangeRequired
	default:
		return authcontract.StatusDenied
	}
}

func authMessagesFromPAM(messages []pamauth.Message) []authcontract.Message {
	if len(messages) == 0 {
		return nil
	}

	converted := make([]authcontract.Message, 0, len(messages))
	for _, message := range messages {
		converted = append(converted, authcontract.Message{
			Level: authMessageLevelFromPAM(message.Level),
			Text:  message.Text,
		})
	}

	return converted
}

func authMessageLevelFromPAM(level pamauth.MessageLevel) authcontract.MessageLevel {
	switch level {
	case pamauth.MessageLevelInfo:
		return authcontract.MessageLevelInfo
	case pamauth.MessageLevelWarn:
		return authcontract.MessageLevelWarn
	default:
		return authcontract.MessageLevelError
	}
}

func refreshContext(clientIP string, userAgent string) sessions.RefreshContext {
	return sessions.RefreshContext{
		ClientIP:  clientIP,
		UserAgent: userAgent,
	}
}

func newSessionID() string {
	data := make([]byte, sessionIDBytes)
	if _, err := rand.Read(data); err != nil {
		return time.Now().UTC().Format(time.RFC3339Nano)
	}

	return base64.RawURLEncoding.EncodeToString(data)
}
