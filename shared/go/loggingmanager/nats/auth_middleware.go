package nats

import (
	"context"
	"errors"
	"strings"

	"lite-nas/shared/authtoken"
	sharedmessaging "lite-nas/shared/messaging"

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
		if protected && !hasAnyRole(claimsFromContext(ctx), requiredRoles) {
			return errInsufficientRole
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
		if protected && !hasAnyRole(claimsFromContext(ctx), requiredRoles) {
			return nil, errInsufficientRole
		}
		return next(ctx, envelope)
	}
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

func hasAnyRole(claims authtoken.AccessClaims, requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true
	}

	roleSet := buildNormalizedRoleSet(claims.Roles)
	return hasRequiredRole(roleSet, requiredRoles)
}

func buildNormalizedRoleSet(roles []string) map[string]struct{} {
	roleSet := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		key := normalizeRole(role)
		if key == "" {
			continue
		}
		roleSet[key] = struct{}{}
	}
	return roleSet
}

func hasRequiredRole(roleSet map[string]struct{}, requiredRoles []string) bool {
	for _, role := range requiredRoles {
		key := normalizeRole(role)
		if key == "" {
			continue
		}
		if _, ok := roleSet[key]; ok {
			return true
		}
	}
	return false
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}
