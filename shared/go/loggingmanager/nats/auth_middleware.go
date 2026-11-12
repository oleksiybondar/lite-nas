package nats

import (
	"context"
	"errors"
	"fmt"

	"lite-nas/shared/authtoken"
	sharedmessaging "lite-nas/shared/messaging"
	"lite-nas/shared/roleauth"

	"github.com/go-playground/validator/v10"
)

var (
	errAccessTokenMissing = errors.New("access token is required")
	errAccessTokenDenied  = errors.New("access token is invalid")
	errInsufficientRole   = errors.New("insufficient role")
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

type accessClaimsContextKey struct{}

// AuthorizationPolicy defines role requirements per messaging subject.
type AuthorizationPolicy struct {
	RPCRolesBySubject          map[string][]string
	SubscriptionRolesBySubject map[string][]string
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
		claims, err := validateAccessToken(verifier, envelope)
		if err != nil {
			return err
		}
		return next(context.WithValue(ctx, accessClaimsContextKey{}, claims), envelope)
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
		claims, err := validateAccessToken(verifier, envelope)
		if err != nil {
			return nil, err
		}
		return next(context.WithValue(ctx, accessClaimsContextKey{}, claims), envelope)
	}
}

// NewRoleAuthorizationSubscriptionMiddleware enforces role requirements for
// selected subscription subjects.
func NewRoleAuthorizationSubscriptionMiddleware(
	policy AuthorizationPolicy,
) sharedmessaging.SubscriptionMiddleware {
	return func(
		ctx context.Context,
		envelope sharedmessaging.Envelope,
		next sharedmessaging.MessageNext,
	) error {
		requiredRoles, protected := policy.SubscriptionRolesBySubject[envelope.Subject]
		claims := claimsFromContext(ctx)
		if protected && !roleauth.HasAnyRole(claims.Roles, requiredRoles) {
			return insufficientRoleError(envelope.Subject, claims, requiredRoles)
		}
		return next(ctx, envelope)
	}
}

// NewRoleAuthorizationRPCMiddleware enforces role requirements for selected
// RPC subjects.
func NewRoleAuthorizationRPCMiddleware(
	policy AuthorizationPolicy,
) sharedmessaging.RPCMiddleware {
	return func(
		ctx context.Context,
		envelope sharedmessaging.Envelope,
		next sharedmessaging.RPCNext,
	) (any, error) {
		requiredRoles, protected := policy.RPCRolesBySubject[envelope.Subject]
		claims := claimsFromContext(ctx)
		if protected && !roleauth.HasAnyRole(claims.Roles, requiredRoles) {
			return nil, insufficientRoleError(envelope.Subject, claims, requiredRoles)
		}
		return next(ctx, envelope)
	}
}

// insufficientRoleError preserves the authorization sentinel and adds the
// principal context needed to identify which caller needs a role correction.
func insufficientRoleError(subject string, claims authtoken.AccessClaims, requiredRoles []string) error {
	return fmt.Errorf("%w: subject=%q login=%q roles=%q required_roles=%q", errInsufficientRole, subject, claims.Login, claims.Roles, requiredRoles)
}

// validateAccessToken decodes and validates one access token payload against
// the local JWT verifier.
func validateAccessToken(
	verifier accessTokenVerifier,
	envelope sharedmessaging.Envelope,
) (authtoken.AccessClaims, error) {
	input, err := decodePayload[accessTokenPayload](envelope)
	if err != nil {
		return authtoken.AccessClaims{}, err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(input); err != nil {
		return authtoken.AccessClaims{}, errAccessTokenMissing
	}

	claims, err := verifier.Verify(input.AccessToken)
	if err != nil {
		return authtoken.AccessClaims{}, errAccessTokenDenied
	}

	return claims, nil
}

func claimsFromContext(ctx context.Context) authtoken.AccessClaims {
	claims, _ := ctx.Value(accessClaimsContextKey{}).(authtoken.AccessClaims)
	return claims
}
