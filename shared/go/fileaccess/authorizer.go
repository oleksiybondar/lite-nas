// Package fileaccess provides the Auth and RBAC RPC checks shared by file-backed services.
package fileaccess

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	authcontract "lite-nas/shared/contracts/auth"
	rbaccontract "lite-nas/shared/contracts/rbac"
)

var (
	// ErrUnauthenticated means Auth did not verify a usable access token.
	ErrUnauthenticated = errors.New("access token is not authenticated")
	// ErrSubjectUnavailable means the verified username had no trustworthy RBAC UID.
	ErrSubjectUnavailable = errors.New("RBAC subject is unavailable")
	// ErrDenied means RBAC did not allow the requested path operation.
	ErrDenied = errors.New("path operation denied")
	// ErrInvalidPath means the caller supplied an ambiguous or unsupported path.
	ErrInvalidPath = errors.New("path must be clean and absolute")
)

// Requester is the subset of the shared messaging client needed for Auth and RBAC RPCs.
type Requester interface {
	Request(context.Context, string, any, any) error
}

// Subject binds an Auth-verified username to the UID and groups returned by RBAC.
type Subject struct {
	Username string
	UID      string
	Groups   []string
}

// Authorizer sends every token and path decision to the owning services.
type Authorizer struct {
	client Requester
}

// NewAuthorizer uses an existing NATS requester without taking ownership of its lifecycle.
func NewAuthorizer(client Requester) (*Authorizer, error) {
	if client == nil {
		return nil, errors.New("messaging requester is required")
	}
	return &Authorizer{client: client}, nil
}

// Resolve validates one access token and resolves its verified username to a non-root UID.
// Each protected HTTP request must call Resolve independently.
func (a *Authorizer) Resolve(ctx context.Context, token string) (Subject, error) {
	if len(token) == 0 || len(token) > 8192 || strings.TrimSpace(token) != token {
		return Subject{}, ErrUnauthenticated
	}
	var validated authcontract.ValidateAccessTokenResponse
	if err := a.client.Request(ctx, authcontract.ValidateAccessTokenRPCSubject,
		authcontract.ValidateAccessTokenRequest{AccessToken: token}, &validated); err != nil {
		return Subject{}, fmt.Errorf("Auth validation failed: %w", err)
	}
	if !validated.Valid || validated.Status != authcontract.StatusAuthenticated ||
		validated.Username == "" || len(validated.Username) > 128 {
		return Subject{}, ErrUnauthenticated
	}
	return a.resolveUsername(ctx, validated.Username)
}

// ResolveCaptionGrant validates one exact file path before resolving user ACLs.
func (a *Authorizer) ResolveCaptionGrant(ctx context.Context, grant, path string) (Subject, error) {
	if len(grant) == 0 || len(grant) > 128 || ValidateHostPath(path) != nil {
		return Subject{}, ErrUnauthenticated
	}
	var validated authcontract.CaptionGrantValidateResponse
	if err := a.client.Request(ctx, authcontract.CaptionGrantValidateRPCSubject,
		authcontract.CaptionGrantValidateRequest{Grant: grant, Path: path}, &validated); err != nil {
		return Subject{}, fmt.Errorf("caption grant validation failed: %w", err)
	}
	if !validated.Valid || validated.Username == "" || len(validated.Username) > 128 {
		return Subject{}, ErrUnauthenticated
	}
	return a.resolveUsername(ctx, validated.Username)
}

// resolveUsername maps an Auth-verified login to its current non-root host identity.
func (a *Authorizer) resolveUsername(ctx context.Context, username string) (Subject, error) {
	var roles rbaccontract.GetSubjectRolesResponse
	if err := a.client.Request(ctx, rbaccontract.GetSubjectRolesRPCSubject,
		rbaccontract.GetSubjectRolesRequest{Username: username}, &roles); err != nil {
		return Subject{}, fmt.Errorf("RBAC subject resolution failed: %w", err)
	}
	uid, err := strconv.ParseUint(roles.UID, 10, 32)
	if err != nil || uid == 0 || strconv.FormatUint(uid, 10) != roles.UID {
		return Subject{}, ErrSubjectUnavailable
	}
	return Subject{Username: username, UID: roles.UID, Groups: append([]string(nil), roles.Groups...)}, nil
}

// Read asks RBAC to authorize reading an existing resolved file or directory.
func (a *Authorizer) Read(ctx context.Context, subject Subject, path string) error {
	if err := validateDecisionInput(subject, path); err != nil {
		return err
	}
	var result rbaccontract.DecisionResponse
	if err := a.client.Request(ctx, rbaccontract.CanReadPathRPCSubject,
		rbaccontract.CheckPathRequest{UID: subject.UID, Path: path}, &result); err != nil {
		return fmt.Errorf("RBAC read decision failed: %w", err)
	}
	if !result.Allowed {
		return ErrDenied
	}
	return nil
}

// Operation asks RBAC to authorize a directory-entry change or traversal.
func (a *Authorizer) Operation(ctx context.Context, subject Subject, path string, operation rbaccontract.PathOperation) error {
	if err := validateDecisionInput(subject, path); err != nil {
		return err
	}
	switch operation {
	case rbaccontract.PathOperationCreate, rbaccontract.PathOperationReplace,
		rbaccontract.PathOperationDelete, rbaccontract.PathOperationTraverse:
	default:
		return errors.New("unsupported RBAC path operation")
	}
	var result rbaccontract.DecisionResponse
	if err := a.client.Request(ctx, rbaccontract.CheckPathOperationRPCSubject,
		rbaccontract.CheckPathOperationRequest{UID: subject.UID, Path: path, Operation: operation}, &result); err != nil {
		return fmt.Errorf("RBAC path operation decision failed: %w", err)
	}
	if !result.Allowed {
		return ErrDenied
	}
	return nil
}

// validateDecisionInput rejects values that cannot be sent to the RBAC path contracts.
func validateDecisionInput(subject Subject, path string) error {
	uid, err := strconv.ParseUint(subject.UID, 10, 32)
	if subject.Username == "" || err != nil || uid == 0 || strconv.FormatUint(uid, 10) != subject.UID {
		return ErrSubjectUnavailable
	}
	if len(path) == 0 || len(path) > 4096 || !filepath.IsAbs(path) || filepath.Clean(path) != path || strings.IndexByte(path, 0) >= 0 {
		return ErrInvalidPath
	}
	return nil
}
