package authtoken

import "errors"

var (
	errMissingSigningKey      = errors.New("auth token signing key is required")
	errMissingVerificationKey = errors.New("auth token verification key is required")
	errInvalidSigningKey      = errors.New("auth token signing key has invalid length")
	errInvalidVerificationKey = errors.New("auth token verification key has invalid length")
	errMissingIssuer          = errors.New("auth token issuer is required")
	errMissingAudience        = errors.New("auth token audience is required")
	errInvalidAccessLifetime  = errors.New("auth token access lifetime must be greater than zero")
	errInvalidClockSkew       = errors.New("auth token clock skew must not be negative")
	errMissingSubject         = errors.New("auth token subject is required")
	errMissingLogin           = errors.New("auth token login is required")
	errUnexpectedSigningAlg   = errors.New("unexpected auth token signing algorithm")
	errMissingPEMBlock        = errors.New("missing PEM block")
	errUnsupportedSigningKey  = errors.New("unsupported auth token signing key type")
	errUnsupportedPublicKey   = errors.New("unsupported auth token public key type")
)
