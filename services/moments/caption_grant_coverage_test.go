package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	authcontract "lite-nas/shared/contracts/auth"
)

// TestCaptionGrantIssuesForExactManifest checks the Auth Service request scope.
func TestCaptionGrantIssuesForExactManifest(t *testing.T) {
	path := "/photos/lizard.jpg"
	wantHash := sha256.Sum256([]byte(path))
	manager := &scanManager{grants: subjectLookupFunc(func(_ context.Context, subject string, request any, response any) error {
		if subject != authcontract.CaptionGrantIssueRPCSubject || request.(authcontract.CaptionGrantIssueRequest).PathHashes[0] != hex.EncodeToString(wantHash[:]) {
			return errors.New("incorrect grant scope")
		}
		*response.(*authcontract.CaptionGrantIssueResponse) = authcontract.CaptionGrantIssueResponse{
			Grant: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ", ExpiresAt: time.Now().Add(time.Hour),
		}
		return nil
	})}
	grant, err := manager.issueCaptionGrant(t.Context(), scanJob{ID: "job"}, "user-token", []string{path})
	if err != nil || len(grant) != 43 {
		t.Fatalf("issued grant = %q, error = %v", grant, err)
	}
}

// TestCaptionGrantRejectsInvalidAuthResponse checks untrusted RPC output.
func TestCaptionGrantRejectsInvalidAuthResponse(t *testing.T) {
	manager := &scanManager{grants: subjectLookupFunc(func(_ context.Context, _ string, _ any, response any) error {
		*response.(*authcontract.CaptionGrantIssueResponse) = authcontract.CaptionGrantIssueResponse{Grant: "short", ExpiresAt: time.Now().Add(time.Hour)}
		return nil
	})}
	_, err := manager.issueCaptionGrant(t.Context(), scanJob{ID: "job"}, "user-token", []string{"/photos/lizard.jpg"})
	if err == nil {
		t.Fatal("invalid grant was accepted")
	}
}

// TestCaptionGrantPropagatesAuthFailure checks RPC transport errors.
func TestCaptionGrantPropagatesAuthFailure(t *testing.T) {
	failure := errors.New("auth unavailable")
	manager := &scanManager{grants: subjectLookupFunc(func(context.Context, string, any, any) error { return failure })}
	_, err := manager.issueCaptionGrant(t.Context(), scanJob{ID: "job"}, "user-token", []string{"/photos/lizard.jpg"})
	if !errors.Is(err, failure) {
		t.Fatalf("grant error = %v, want %v", err, failure)
	}
}
