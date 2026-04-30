package authtoken

import (
	"crypto/ed25519"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Principal contains the authenticated identity and optional authorization
// labels copied into an access token.
type Principal struct {
	// Subject is the stable principal identifier written to the JWT "sub"
	// claim. Prefer a stable host identity such as UID when available.
	Subject string
	// Login is the authenticated host login name written to the custom
	// "login" claim.
	Login string
	// Scopes contains optional fine-grained authorization labels copied into
	// the token. Empty scopes mean authenticated identity only.
	Scopes []string
	// Roles contains optional coarse authorization labels copied into the
	// token. Empty roles mean authenticated identity only.
	Roles []string
}

// IssuerOptions contains token policy values needed to sign access tokens.
type IssuerOptions struct {
	// Issuer is written to the JWT "iss" claim and must match verifier
	// expectations.
	Issuer string
	// Audience is written to the JWT "aud" claim and names the API surface
	// that should accept the token.
	Audience string
	// AccessLifetime is added to the issue time to compute the JWT "exp"
	// claim. It must be greater than zero.
	AccessLifetime time.Duration
}

// VerifierOptions contains token policy values needed to verify access tokens.
type VerifierOptions struct {
	// Issuer is the required JWT "iss" claim value.
	Issuer string
	// Audience is the required JWT "aud" claim value.
	Audience string
	// ClockSkew is the allowed time leeway for time-based JWT validation.
	// It must not be negative.
	ClockSkew time.Duration
}

// Issuer signs LiteNAS JWT access tokens.
//
// An Issuer is immutable after construction and is safe to share between
// request handlers as long as the private key material is managed by the caller.
type Issuer struct {
	key            ed25519.PrivateKey
	issuer         string
	audience       string
	accessLifetime time.Duration
	now            func() time.Time
}

// Verifier validates LiteNAS JWT access tokens.
//
// A Verifier checks the EdDSA signature, expected issuer, expected audience,
// issued-at time, expiration, and configured clock skew.
type Verifier struct {
	key      ed25519.PublicKey
	issuer   string
	audience string
	leeway   time.Duration
	now      func() time.Time
}

// NewIssuer creates an access-token issuer from parsed key material.
//
// Parameters:
//   - options: issuer, audience, and access-token lifetime policy.
//   - key: Ed25519 private key used to sign JWTs with EdDSA.
//
// The function validates option values and key length. It does not read files,
// parse PEM data, or depend on service configuration.
func NewIssuer(options IssuerOptions, key ed25519.PrivateKey) (Issuer, error) {
	if err := validateIssuerOptions(options); err != nil {
		return Issuer{}, err
	}
	if len(key) == 0 {
		return Issuer{}, errMissingSigningKey
	}
	if len(key) != ed25519.PrivateKeySize {
		return Issuer{}, errInvalidSigningKey
	}

	return Issuer{
		key:            key,
		issuer:         options.Issuer,
		audience:       options.Audience,
		accessLifetime: options.AccessLifetime,
		now:            time.Now,
	}, nil
}

// NewVerifier creates an access-token verifier from parsed key material.
//
// Parameters:
//   - options: required issuer, required audience, and clock-skew policy.
//   - key: Ed25519 public key used to verify EdDSA JWT signatures.
//
// The function validates option values and key length. It does not read files,
// parse PEM data, or depend on service configuration.
func NewVerifier(options VerifierOptions, key ed25519.PublicKey) (Verifier, error) {
	if err := validateVerifierOptions(options); err != nil {
		return Verifier{}, err
	}
	if len(key) == 0 {
		return Verifier{}, errMissingVerificationKey
	}
	if len(key) != ed25519.PublicKeySize {
		return Verifier{}, errInvalidVerificationKey
	}

	return Verifier{
		key:      key,
		issuer:   options.Issuer,
		audience: options.Audience,
		leeway:   options.ClockSkew,
		now:      time.Now,
	}, nil
}

// Issue signs a short-lived access token for a principal.
//
// Parameters:
//   - principal: authenticated identity and optional authorization labels to
//     copy into the access token.
//
// It returns the compact JWT string plus the claims used to produce it. The
// claims return value is intended for response metadata, logging, and tests; the
// signed token remains the authoritative bearer value.
func (i Issuer) Issue(principal Principal) (string, AccessClaims, error) {
	if principal.Subject == "" {
		return "", AccessClaims{}, errMissingSubject
	}
	if principal.Login == "" {
		return "", AccessClaims{}, errMissingLogin
	}

	now := i.now().UTC()
	tokenID, err := newTokenID()
	if err != nil {
		return "", AccessClaims{}, err
	}

	claims := newAccessClaims(accessClaimOptions{
		Issuer:    i.issuer,
		Audience:  i.audience,
		IssuedAt:  now,
		ExpiresAt: now.Add(i.accessLifetime),
		TokenID:   tokenID,
		Principal: principal,
	})

	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(i.key)
	if err != nil {
		return "", AccessClaims{}, err
	}

	return signedToken, claims, nil
}

// Verify validates and decodes a signed access token.
//
// Parameters:
//   - tokenText: compact JWT access token string received from a caller.
//
// Verification fails closed when the token is malformed, signed with an
// unexpected algorithm, signed by a different key, expired, missing a required
// issued-at or expiration claim, or intended for a different issuer/audience.
func (v Verifier) Verify(tokenText string) (AccessClaims, error) {
	claims := AccessClaims{}
	parser := jwt.NewParser(
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(v.issuer),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(v.leeway),
		jwt.WithTimeFunc(v.now),
	)

	token, err := parser.ParseWithClaims(tokenText, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodEdDSA {
			return nil, errUnexpectedSigningAlg
		}

		return v.key, nil
	})
	if err != nil {
		return AccessClaims{}, err
	}

	if !token.Valid {
		return AccessClaims{}, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func (i Issuer) withClock(now func() time.Time) Issuer {
	i.now = now
	return i
}

func (v Verifier) withClock(now func() time.Time) Verifier {
	v.now = now
	return v
}

func validateIssuerOptions(options IssuerOptions) error {
	switch {
	case options.Issuer == "":
		return errMissingIssuer
	case options.Audience == "":
		return errMissingAudience
	case options.AccessLifetime <= 0:
		return errInvalidAccessLifetime
	default:
		return nil
	}
}

func validateVerifierOptions(options VerifierOptions) error {
	switch {
	case options.Issuer == "":
		return errMissingIssuer
	case options.Audience == "":
		return errMissingAudience
	case options.ClockSkew < 0:
		return errInvalidClockSkew
	default:
		return nil
	}
}
