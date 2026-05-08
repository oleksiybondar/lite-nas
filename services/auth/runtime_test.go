package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lite-nas/services/auth/sessions"
	"lite-nas/shared/authtoken"
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/testutil/authtokentest"
)

func TestNewAuthTokenRuntimeLoadsConfiguredKeys(t *testing.T) {
	t.Parallel()

	runtime, err := newAuthTokenRuntime(authTokenConfigFixture(t))
	if err != nil {
		t.Fatalf("newAuthTokenRuntime() error = %v", err)
	}

	tokenText, _, err := runtime.Issuer.Issue(authPrincipalFixture())
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if _, err = runtime.Verifier.Verify(tokenText); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestNewAuthTokenRuntimeReturnsIssuerError(t *testing.T) {
	t.Parallel()

	cfg := authTokenConfigFixture(t)
	cfg.SigningKey = filepath.Join(t.TempDir(), "missing.key")

	if _, err := newAuthTokenRuntime(cfg); err == nil {
		t.Fatal("newAuthTokenRuntime() error = nil, want signing key read error")
	}
}

func TestNewAuthTokenRuntimeReturnsVerifierError(t *testing.T) {
	t.Parallel()

	cfg := authTokenConfigFixture(t)
	cfg.VerificationCert = filepath.Join(t.TempDir(), "missing.crt")

	if _, err := newAuthTokenRuntime(cfg); err == nil {
		t.Fatal("newAuthTokenRuntime() error = nil, want verification cert read error")
	}
}

func TestNewAuthTokenIssuerRejectsInvalidSigningKey(t *testing.T) {
	t.Parallel()

	cfg := authTokenConfigFixture(t)
	writeRuntimeTestFile(t, cfg.SigningKey, []byte("not pem"))

	if _, err := newAuthTokenIssuer(cfg); err == nil {
		t.Fatal("newAuthTokenIssuer() error = nil, want invalid signing key error")
	}
}

func TestNewAuthTokenVerifierRejectsInvalidCertificate(t *testing.T) {
	t.Parallel()

	cfg := authTokenConfigFixture(t)
	writeRuntimeTestFile(t, cfg.VerificationCert, []byte("not pem"))

	if _, err := newAuthTokenVerifier(cfg); err == nil {
		t.Fatal("newAuthTokenVerifier() error = nil, want invalid certificate error")
	}
}

func TestNewAuthTokenIssuerReturnsConfiguredOptionError(t *testing.T) {
	t.Parallel()

	cfg := authTokenConfigFixture(t)
	cfg.Issuer = ""

	if _, err := newAuthTokenIssuer(cfg); err == nil {
		t.Fatal("newAuthTokenIssuer() error = nil, want invalid issuer option error")
	}
}

func TestNewAuthTokenVerifierReturnsConfiguredOptionError(t *testing.T) {
	t.Parallel()

	cfg := authTokenConfigFixture(t)
	cfg.Audience = ""

	if _, err := newAuthTokenVerifier(cfg); err == nil {
		t.Fatal("newAuthTokenVerifier() error = nil, want invalid audience option error")
	}
}

func TestNewAuthTokenIssuerReturnsReadError(t *testing.T) {
	t.Parallel()

	cfg := authTokenConfigFixture(t)
	cfg.SigningKey = filepath.Join(t.TempDir(), "missing.key")

	if _, err := newAuthTokenIssuer(cfg); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("newAuthTokenIssuer() error = %v, want missing file error", err)
	}
}

func TestNewAuthTokenVerifierReturnsReadError(t *testing.T) {
	t.Parallel()

	cfg := authTokenConfigFixture(t)
	cfg.VerificationCert = filepath.Join(t.TempDir(), "missing.crt")

	if _, err := newAuthTokenVerifier(cfg); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("newAuthTokenVerifier() error = %v, want missing file error", err)
	}
}

func TestNewRefreshStoreUsesConfiguredClientIPPolicy(t *testing.T) {
	t.Parallel()

	store := newRefreshStore(sharedconfig.AuthTokenConfig{EnforceRefreshClientIP: true})
	token, _, err := store.Create(refreshCreateInputFixture())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	context := refreshCreateInputFixture().Context
	context.ClientIP = "192.168.1.11"
	if _, _, err = store.Rotate(token.Value, context); err == nil {
		t.Fatal("Rotate() error = nil, want client IP mismatch error")
	}
}

func authTokenConfigFixture(t *testing.T) sharedconfig.AuthTokenConfig {
	t.Helper()

	publicKey, privateKey := authtokentest.MustGenerateEd25519Key(t)
	tempDir := t.TempDir()
	signingKeyPath := filepath.Join(tempDir, "token-signing.key")
	verificationCertPath := filepath.Join(tempDir, "token-signing.crt")
	writeRuntimeTestFile(t, signingKeyPath, authtokentest.MustMarshalEd25519PrivateKeyPEM(t, privateKey))
	writeRuntimeTestFile(t, verificationCertPath, authtokentest.MustCreateEd25519CertificatePEM(t, publicKey, privateKey))

	return sharedconfig.AuthTokenConfig{
		Issuer:           "lite-nas-auth",
		Audience:         "lite-nas-management-api",
		AccessLifetime:   15 * time.Minute,
		ClockSkew:        30 * time.Second,
		SigningKey:       signingKeyPath,
		VerificationCert: verificationCertPath,
	}
}

func writeRuntimeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
}

func authPrincipalFixture() authtoken.Principal {
	return authtoken.Principal{
		Subject: "1000",
		Login:   "alice",
	}
}

func refreshCreateInputFixture() sessions.CreateInput {
	return sessions.CreateInput{
		SessionID: "session-id",
		Subject:   "1000",
		Login:     "alice",
		Context: sessions.RefreshContext{
			ClientIP:  "192.168.1.10",
			UserAgent: "browser",
		},
		TTL: 30 * 24 * time.Hour,
	}
}
