package nats

import (
	"context"
	"errors"
	"testing"

	"lite-nas/shared/authtoken"
	sharedmessaging "lite-nas/shared/messaging"
	"lite-nas/shared/roleauth"
)

var errUnexpectedNextCall = errors.New("next should not be called")

// middlewareVerifierStub is a test double for access-token verifier calls.
type middlewareVerifierStub struct {
	err    error
	claims authtoken.AccessClaims
}

// Verify returns configured verification error.
func (stub middlewareVerifierStub) Verify(tokenText string) (authtoken.AccessClaims, error) {
	if tokenText == "" {
		return authtoken.AccessClaims{}, errors.New("unexpected empty token")
	}
	if stub.err != nil {
		return authtoken.AccessClaims{}, stub.err
	}
	return stub.claims, nil
}

func TestSubscriptionMiddlewareCallsNextWhenTokenIsValid(t *testing.T) {
	t.Parallel()

	middleware := NewAccessTokenValidationSubscriptionMiddleware(middlewareVerifierStub{})

	nextCalled := false
	err := middleware(
		context.Background(),
		sharedmessaging.Envelope{Payload: []byte(`{"access_token":"token"}`)},
		func(context.Context, sharedmessaging.Envelope) error {
			nextCalled = true
			return nil
		},
	)
	if err != nil {
		t.Fatalf("middleware() error = %v", err)
	}
	if !nextCalled {
		t.Fatal("next was not called")
	}
}

func TestSubscriptionMiddlewareRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	middleware := NewAccessTokenValidationSubscriptionMiddleware(
		middlewareVerifierStub{err: errors.New("invalid token")},
	)

	err := middleware(
		context.Background(),
		sharedmessaging.Envelope{Payload: []byte(`{"access_token":"token"}`)},
		func(context.Context, sharedmessaging.Envelope) error {
			t.Fatal("next should not be called")
			return nil
		},
	)
	if !errors.Is(err, errAccessTokenDenied) {
		t.Fatalf("middleware() error = %v, want errAccessTokenDenied", err)
	}
}

func TestSubscriptionMiddlewareRejectsMissingToken(t *testing.T) {
	t.Parallel()

	middleware := NewAccessTokenValidationSubscriptionMiddleware(
		middlewareVerifierStub{},
	)

	err := middleware(
		context.Background(),
		sharedmessaging.Envelope{Payload: []byte(`{}`)},
		func(context.Context, sharedmessaging.Envelope) error {
			t.Fatal("next should not be called")
			return nil
		},
	)
	if !errors.Is(err, errAccessTokenMissing) {
		t.Fatalf("middleware() error = %v, want errAccessTokenMissing", err)
	}
}

func TestRPCMiddlewareReturnsNextResponseWhenTokenIsValid(t *testing.T) {
	t.Parallel()

	middleware := NewAccessTokenValidationRPCMiddleware(middlewareVerifierStub{})

	result, err := middleware(
		context.Background(),
		sharedmessaging.Envelope{Payload: []byte(`{"access_token":"token"}`)},
		func(context.Context, sharedmessaging.Envelope) (any, error) {
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("middleware() error = %v", err)
	}
	if result != "ok" {
		t.Fatalf("result = %v, want ok", result)
	}
}

func TestRPCMiddlewareRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	middleware := NewAccessTokenValidationRPCMiddleware(
		middlewareVerifierStub{err: errors.New("invalid token")},
	)

	_, err := middleware(
		context.Background(),
		sharedmessaging.Envelope{Payload: []byte(`{"access_token":"token"}`)},
		func(context.Context, sharedmessaging.Envelope) (any, error) {
			t.Fatal("next should not be called")
			return nil, errUnexpectedNextCall
		},
	)
	if !errors.Is(err, errAccessTokenDenied) {
		t.Fatalf("middleware() error = %v, want errAccessTokenDenied", err)
	}
}

func TestRoleAuthorizationRPCMiddlewareAllowsMatchingRole(t *testing.T) {
	t.Parallel()

	middleware := NewRoleAuthorizationRPCMiddleware(AuthorizationPolicy{
		RPCRolesBySubject: map[string][]string{
			"subject.protected": roleauth.AllowedRoles(roleauth.RequirementOperator),
		},
	})

	claims := authtoken.AccessClaims{Roles: []string{roleauth.RoleOperator}}
	result, err := middleware(
		context.WithValue(context.Background(), accessClaimsContextKey{}, claims),
		sharedmessaging.Envelope{Subject: "subject.protected"},
		func(context.Context, sharedmessaging.Envelope) (any, error) {
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("middleware() error = %v", err)
	}
	if result != "ok" {
		t.Fatalf("result = %v, want ok", result)
	}
}

func TestRoleAuthorizationRPCMiddlewareRejectsMissingRole(t *testing.T) {
	t.Parallel()

	middleware := NewRoleAuthorizationRPCMiddleware(AuthorizationPolicy{
		RPCRolesBySubject: map[string][]string{
			"subject.protected": roleauth.AllowedRoles(roleauth.RequirementOperator),
		},
	})

	_, err := middleware(
		context.WithValue(context.Background(), accessClaimsContextKey{}, authtoken.AccessClaims{Roles: []string{"viewer"}}),
		sharedmessaging.Envelope{Subject: "subject.protected"},
		func(context.Context, sharedmessaging.Envelope) (any, error) {
			return nil, errUnexpectedNextCall
		},
	)
	if !errors.Is(err, errInsufficientRole) {
		t.Fatalf("middleware() error = %v, want errInsufficientRole", err)
	}
}
