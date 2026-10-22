package fileaccess

import (
	"context"
	"errors"
	"testing"

	authcontract "lite-nas/shared/contracts/auth"
	rbaccontract "lite-nas/shared/contracts/rbac"
)

// requestFunc lets each test control one RPC boundary without a running NATS server.
type requestFunc func(context.Context, string, any, any) error

func (fn requestFunc) Request(ctx context.Context, subject string, request any, response any) error {
	return fn(ctx, subject, request, response)
}

// TestResolveRequiresVerifiedTokenAndUID covers file-service-svc/FR-001 and FR-002.
func TestResolveRequiresVerifiedTokenAndUID(t *testing.T) {
	for _, test := range []struct {
		name     string
		response authcontract.ValidateAccessTokenResponse
		uid      string
		wantErr  error
	}{
		{"valid", authcontract.ValidateAccessTokenResponse{Valid: true, Status: authcontract.StatusAuthenticated, Username: "alice"}, "1001", nil},
		{"invalid", authcontract.ValidateAccessTokenResponse{Valid: false, Status: authcontract.StatusAuthenticated, Username: "alice"}, "1001", ErrUnauthenticated},
		{"wrong status", authcontract.ValidateAccessTokenResponse{Valid: true, Status: authcontract.StatusDenied, Username: "alice"}, "1001", ErrUnauthenticated},
		{"missing username", authcontract.ValidateAccessTokenResponse{Valid: true, Status: authcontract.StatusAuthenticated}, "1001", ErrUnauthenticated},
		{"root UID", authcontract.ValidateAccessTokenResponse{Valid: true, Status: authcontract.StatusAuthenticated, Username: "alice"}, "0", ErrSubjectUnavailable},
		{"malformed UID", authcontract.ValidateAccessTokenResponse{Valid: true, Status: authcontract.StatusAuthenticated, Username: "alice"}, "not-a-uid", ErrSubjectUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			authorizer, err := NewAuthorizer(requestFunc(func(_ context.Context, subject string, request any, response any) error {
				calls++
				switch subject {
				case authcontract.ValidateAccessTokenRPCSubject:
					if request.(authcontract.ValidateAccessTokenRequest).AccessToken != "token" {
						t.Fatal("Auth received a different token")
					}
					*response.(*authcontract.ValidateAccessTokenResponse) = test.response
				case rbaccontract.GetSubjectRolesRPCSubject:
					if request.(rbaccontract.GetSubjectRolesRequest).Username != "alice" {
						t.Fatal("RBAC did not receive the verified username")
					}
					response.(*rbaccontract.GetSubjectRolesResponse).UID = test.uid
				default:
					t.Fatalf("unexpected RPC subject %q", subject)
				}
				return nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			subject, err := authorizer.Resolve(t.Context(), "token")
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Resolve error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr == nil && subject != (Subject{Username: "alice", UID: "1001"}) {
				t.Fatalf("Resolve subject = %+v", subject)
			}
			if test.wantErr == ErrUnauthenticated && calls != 1 {
				t.Fatalf("unauthenticated token caused %d RPCs", calls)
			}
		})
	}
}

// TestResolveStopsOnDependencyFailures covers file-service-svc/RR-001.
func TestResolveStopsOnDependencyFailures(t *testing.T) {
	for _, failedSubject := range []string{
		authcontract.ValidateAccessTokenRPCSubject,
		rbaccontract.GetSubjectRolesRPCSubject,
	} {
		t.Run(failedSubject, func(t *testing.T) {
			calls := 0
			authorizer, err := NewAuthorizer(requestFunc(func(_ context.Context, subject string, _ any, response any) error {
				calls++
				if subject == failedSubject {
					return errors.New("dependency unavailable")
				}
				if subject == authcontract.ValidateAccessTokenRPCSubject {
					*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{
						Valid: true, Status: authcontract.StatusAuthenticated, Username: "alice",
					}
				}
				return nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := authorizer.Resolve(t.Context(), "token"); err == nil {
				t.Fatal("dependency failure resolved a subject")
			}
			if failedSubject == authcontract.ValidateAccessTokenRPCSubject && calls != 1 {
				t.Fatalf("Auth failure caused %d RPC calls", calls)
			}
		})
	}
}

// TestDecisionsFailClosed covers file-service-svc/FR-003, FR-004, and RR-001.
func TestDecisionsFailClosed(t *testing.T) {
	for _, test := range []struct {
		name    string
		allowed bool
		rpcErr  error
		wantErr error
	}{
		{"allow", true, nil, nil},
		{"deny", false, nil, ErrDenied},
		{"RPC failure", true, errors.New("connection lost"), nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			authorizer, err := NewAuthorizer(requestFunc(func(_ context.Context, rpcSubject string, request any, response any) error {
				if rpcSubject != rbaccontract.CanReadPathRPCSubject {
					t.Fatalf("unexpected RPC subject %q", rpcSubject)
				}
				if request.(rbaccontract.CheckPathRequest).Path != "/home/alice/file" {
					t.Fatal("RBAC received a different path")
				}
				response.(*rbaccontract.DecisionResponse).Allowed = test.allowed
				return test.rpcErr
			}))
			if err != nil {
				t.Fatal(err)
			}
			err = authorizer.Read(t.Context(), Subject{Username: "alice", UID: "1001"}, "/home/alice/file")
			if test.rpcErr != nil && err == nil {
				t.Fatal("RPC failure allowed a read")
			}
			if test.rpcErr == nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("Read error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

// TestOperationRejectsInvalidInputs covers file-service-svc/FR-004 and FR-009.
func TestOperationRejectsInvalidInputs(t *testing.T) {
	authorizer, err := NewAuthorizer(requestFunc(func(_ context.Context, _ string, _ any, _ any) error {
		t.Fatal("invalid input reached RBAC")
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	subject := Subject{Username: "alice", UID: "1001"}
	if err := authorizer.Operation(t.Context(), subject, "/home/../etc/passwd", rbaccontract.PathOperationDelete); !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("unclean path error = %v", err)
	}
	if err := authorizer.Operation(t.Context(), subject, "/home/alice/file", "unsupported"); err == nil {
		t.Fatal("unsupported operation was accepted")
	}
}
