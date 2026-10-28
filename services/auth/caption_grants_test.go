package main

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"testing"
	"time"

	"lite-nas/services/auth/pamauth"
	"lite-nas/shared/authtoken"
	authcontract "lite-nas/shared/contracts/auth"
)

// mustCaptionGrantStore prepares protected durable state for one test.
func mustCaptionGrantStore(t *testing.T, now func() time.Time) *captionGrantStore {
	t.Helper()
	store, err := openCaptionGrantStore(filepath.Join(t.TempDir(), "caption-grants.json"), now)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

// mustIssueCaptionGrant creates a grant for one known file path.
func mustIssueCaptionGrant(t *testing.T, store *captionGrantStore) string {
	t.Helper()
	path := sha256.Sum256([]byte("/photos/lizard.jpg"))
	grant, _, err := store.issue("job", "alice", []string{hex.EncodeToString(path[:])})
	if err != nil {
		t.Fatal(err)
	}
	return grant
}

// TestCaptionGrantSurvivesAuthRestart checks durable scoped validation.
func TestCaptionGrantSurvivesAuthRestart(t *testing.T) {
	store := mustCaptionGrantStore(t, nil)
	grant := mustIssueCaptionGrant(t, store)
	reopened, err := openCaptionGrantStore(store.path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.validate(grant, "/photos/lizard.jpg"); got != "alice" {
		t.Fatalf("grant username = %q", got)
	}
}

// TestCaptionGrantRejectsUnlistedPath keeps a grant inside its gallery manifest.
func TestCaptionGrantRejectsUnlistedPath(t *testing.T) {
	store := mustCaptionGrantStore(t, nil)
	grant := mustIssueCaptionGrant(t, store)
	if got := store.validate(grant, "/photos/other.jpg"); got != "" {
		t.Fatalf("unlisted path resolved %q", got)
	}
}

// TestCaptionGrantExpiresAtSevenDays bounds background access.
func TestCaptionGrantExpiresAtSevenDays(t *testing.T) {
	now := time.Now()
	store := mustCaptionGrantStore(t, func() time.Time { return now })
	grant := mustIssueCaptionGrant(t, store)
	store.now = func() time.Time { return now.Add(captionGrantLifetime) }
	if got := store.validate(grant, "/photos/lizard.jpg"); got != "" {
		t.Fatalf("expired grant resolved %q", got)
	}
}

// TestCaptionGrantRevocationStopsReads covers terminal job cleanup.
func TestCaptionGrantRevocationStopsReads(t *testing.T) {
	store := mustCaptionGrantStore(t, nil)
	grant := mustIssueCaptionGrant(t, store)
	if _, err := store.revoke(grant); err != nil {
		t.Fatal(err)
	}
	if got := store.validate(grant, "/photos/lizard.jpg"); got != "" {
		t.Fatalf("revoked grant resolved %q", got)
	}
}

// TestCaptionGrantRPCRequiresUserToken checks the NATS input boundary.
func TestCaptionGrantRPCRequiresUserToken(t *testing.T) {
	runtimeDeps := authRuntimeFixture(t, pamauth.Result{})
	runtimeDeps.CaptionGrants = mustCaptionGrantStore(t, nil)
	path := sha256.Sum256([]byte("/photos/lizard.jpg"))
	response, err := handleCaptionGrantIssueRPC(runtimeDeps, rpcEnvelope(t, authcontract.CaptionGrantIssueRequest{
		UserAccessToken: "not-a-token", JobID: "00000000-0000-4000-8000-000000000001", PathHashes: []string{hex.EncodeToString(path[:])},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if response.Grant != "" {
		t.Fatal("unverified user received a grant")
	}
}

// TestCaptionGrantRPCIssuesForVerifiedUser checks Auth-owned grant creation.
func TestCaptionGrantRPCIssuesForVerifiedUser(t *testing.T) {
	runtimeDeps := authRuntimeFixture(t, pamauth.Result{})
	runtimeDeps.CaptionGrants = mustCaptionGrantStore(t, nil)
	userToken, _, err := runtimeDeps.Tokens.Issuer.Issue(authtoken.Principal{Subject: "alice", Login: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	path := sha256.Sum256([]byte("/photos/lizard.jpg"))
	response, err := handleCaptionGrantIssueRPC(runtimeDeps, rpcEnvelope(t, authcontract.CaptionGrantIssueRequest{
		UserAccessToken: userToken, JobID: "00000000-0000-4000-8000-000000000001", PathHashes: []string{hex.EncodeToString(path[:])},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got := runtimeDeps.CaptionGrants.validate(response.Grant, "/photos/lizard.jpg"); got != "alice" {
		t.Fatalf("verified grant user = %q", got)
	}
}
