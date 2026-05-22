package nats

import (
	"context"
	"errors"

	"lite-nas/shared/authtoken"
	sharedmessaging "lite-nas/shared/messaging"

	"github.com/go-playground/validator/v10"
)

var (
	errAccessTokenMissing = errors.New("access token is required")
	errAccessTokenDenied  = errors.New("access token is invalid")
)

// accessTokenVerifier captures the local JWT verifier dependency used by
// middleware.
type accessTokenVerifier interface {
	Verify(tokenText string) (authtoken.AccessClaims, error)
}

// accessTokenPayload is the minimal request shape needed for token validation.
type accessTokenPayload struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// NewAccessTokenValidationSubscriptionMiddleware validates the access token
// from each subscription payload through local JWT verification before business
// handlers.
func NewAccessTokenValidationSubscriptionMiddleware(
	verifier accessTokenVerifier,
) sharedmessaging.SubscriptionMiddleware {
	return func(
		ctx context.Context,
		envelope sharedmessaging.Envelope,
		next sharedmessaging.MessageNext,
	) error {
		if err := validateAccessToken(verifier, envelope); err != nil {
			return err
		}
		return next(ctx, envelope)
	}
}

// NewAccessTokenValidationRPCMiddleware validates the access token from each
// RPC payload through local JWT verification before business handlers.
func NewAccessTokenValidationRPCMiddleware(
	verifier accessTokenVerifier,
) sharedmessaging.RPCMiddleware {
	return func(
		ctx context.Context,
		envelope sharedmessaging.Envelope,
		next sharedmessaging.RPCNext,
	) (any, error) {
		if err := validateAccessToken(verifier, envelope); err != nil {
			return nil, err
		}
		return next(ctx, envelope)
	}
}

// validateAccessToken decodes and validates one access token payload against
// the local JWT verifier.
func validateAccessToken(
	verifier accessTokenVerifier,
	envelope sharedmessaging.Envelope,
) error {
	input, err := decodePayload[accessTokenPayload](envelope)
	if err != nil {
		return err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(input); err != nil {
		return errAccessTokenMissing
	}

	if _, err := verifier.Verify(input.AccessToken); err != nil {
		return errAccessTokenDenied
	}

	return nil
}
