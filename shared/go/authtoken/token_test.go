package authtoken

import (
	"crypto/ed25519"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"lite-nas/shared/testutil/authtokentest"
	"lite-nas/shared/testutil/testcasetest"
)

func TestIssuerIssueAndVerifierVerifyAccessToken(t *testing.T) {
	t.Parallel()

	publicKey, privateKey := authtokentest.MustGenerateEd25519Key(t)
	issuedAt := time.Unix(1000, 0)
	issuer := mustNewIssuer(t, privateKey).withClock(func() time.Time { return issuedAt })
	verifier := mustNewVerifier(t, publicKey).withClock(func() time.Time { return issuedAt.Add(time.Minute) })

	tokenText, issuedClaims, err := issuer.Issue(Principal{
		Subject: "1000",
		Login:   "alice",
		Scopes:  []string{"monitoring.read"},
		Roles:   []string{"operator"},
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	verifiedClaims, err := verifier.Verify(tokenText)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	assertAccessClaims(t, verifiedClaims, issuedClaims)
}

func TestIssuerIssueRejectsMissingPrincipalFields(t *testing.T) {
	t.Parallel()

	_, privateKey := authtokentest.MustGenerateEd25519Key(t)
	issuer := mustNewIssuer(t, privateKey)

	testCases := []struct {
		name      string
		principal Principal
		wantErr   error
	}{
		{name: "missing subject", principal: Principal{Login: "alice"}, wantErr: errMissingSubject},
		{name: "missing login", principal: Principal{Subject: "1000"}, wantErr: errMissingLogin},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if _, _, err := issuer.Issue(testCase.principal); !errors.Is(err, testCase.wantErr) {
				t.Fatalf("Issue() error = %v, want %v", err, testCase.wantErr)
			}
		})
	}
}

func TestVerifierRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	publicKey, privateKey := authtokentest.MustGenerateEd25519Key(t)
	issuedAt := time.Unix(1000, 0)
	issuer := mustNewIssuer(t, privateKey).withClock(func() time.Time { return issuedAt })
	verifier := mustNewVerifier(t, publicKey).withClock(func() time.Time { return issuedAt.Add(20 * time.Minute) })

	tokenText := mustIssueToken(t, issuer)
	if _, err := verifier.Verify(tokenText); err == nil {
		t.Fatal("Verify() error = nil, want expiry error")
	}
}

func TestVerifierRejectsWrongAudience(t *testing.T) {
	t.Parallel()

	publicKey, privateKey := authtokentest.MustGenerateEd25519Key(t)
	issuer := mustNewIssuer(t, privateKey)
	cfg := authTokenConfigFixture()
	cfg.Audience = "other-audience"
	verifier := mustNewVerifierWithConfig(t, cfg, publicKey)

	tokenText := mustIssueToken(t, issuer)
	if _, err := verifier.Verify(tokenText); err == nil {
		t.Fatal("Verify() error = nil, want audience error")
	}
}

func TestVerifierRejectsUnexpectedSigningAlgorithm(t *testing.T) {
	t.Parallel()

	publicKey, _ := authtokentest.MustGenerateEd25519Key(t)
	verifier := mustNewVerifier(t, publicKey)
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "lite-nas-auth",
			Subject:   "1000",
			Audience:  jwt.ClaimStrings{"lite-nas-management-api"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Login: "alice",
	}

	tokenText, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err = verifier.Verify(tokenText); !errors.Is(err, errUnexpectedSigningAlg) {
		t.Fatalf("Verify() error = %v, want %v", err, errUnexpectedSigningAlg)
	}
}

func mustNewIssuer(t *testing.T, key ed25519.PrivateKey) Issuer {
	t.Helper()

	return mustNewIssuerWithConfig(t, authTokenConfigFixture(), key)
}

func mustNewIssuerWithConfig(t *testing.T, options IssuerOptions, key ed25519.PrivateKey) Issuer {
	t.Helper()

	issuer, err := NewIssuer(options, key)
	if err != nil {
		t.Fatalf("NewIssuer() error = %v", err)
	}

	return issuer
}

func mustNewVerifier(t *testing.T, key ed25519.PublicKey) Verifier {
	t.Helper()

	return mustNewVerifierWithConfig(t, authTokenConfigFixture(), key)
}

func mustNewVerifierWithConfig(t *testing.T, options IssuerOptions, key ed25519.PublicKey) Verifier {
	t.Helper()

	verifier, err := NewVerifier(verifierOptionsFromIssuerOptions(options), key)
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}

	return verifier
}

func mustIssueToken(t *testing.T, issuer Issuer) string {
	t.Helper()

	tokenText, _, err := issuer.Issue(Principal{Subject: "1000", Login: "alice"})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	return tokenText
}

func authTokenConfigFixture() IssuerOptions {
	return IssuerOptions{
		Issuer:         "lite-nas-auth",
		Audience:       "lite-nas-management-api",
		AccessLifetime: 15 * time.Minute,
	}
}

func verifierOptionsFromIssuerOptions(options IssuerOptions) VerifierOptions {
	return VerifierOptions{
		Issuer:    options.Issuer,
		Audience:  options.Audience,
		ClockSkew: 30 * time.Second,
	}
}

func assertAccessClaims(t *testing.T, got AccessClaims, want AccessClaims) {
	t.Helper()

	testCases := []testcasetest.FieldCase[AccessClaims]{
		{Name: "issuer", Got: func(claims AccessClaims) any { return claims.Issuer }, Want: want.Issuer},
		{Name: "subject", Got: func(claims AccessClaims) any { return claims.Subject }, Want: want.Subject},
		{Name: "id", Got: func(claims AccessClaims) any { return claims.ID }, Want: want.ID},
		{Name: "login", Got: func(claims AccessClaims) any { return claims.Login }, Want: want.Login},
		{Name: "scopes", Got: func(claims AccessClaims) any { return claims.Scopes }, Want: want.Scopes},
		{Name: "roles", Got: func(claims AccessClaims) any { return claims.Roles }, Want: want.Roles},
	}

	testcasetest.RunFieldCases(t, func(*testing.T) AccessClaims { return got }, testCases)
}
