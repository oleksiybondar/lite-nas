package authtokentest

import (
	"testing"
	"time"

	"lite-nas/shared/authtoken"
)

// MustIssueAccessToken returns a signed test token and its verifier for the supplied roles.
func MustIssueAccessToken(t *testing.T, roles []string) (authtoken.Verifier, string) {
	t.Helper()
	public, private := MustGenerateEd25519Key(t)
	issuer, err := authtoken.NewIssuer(authtoken.IssuerOptions{Issuer: "test", Audience: "test", AccessLifetime: time.Minute}, private)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{Issuer: "test", Audience: "test"}, public)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := issuer.Issue(authtoken.Principal{Subject: "1", Login: "tester", Roles: roles})
	if err != nil {
		t.Fatal(err)
	}
	return verifier, token
}
