package fileaccess

import (
	"errors"
	"os/user"
	"testing"
)

// mustCurrentAccount returns the Unix account running this test.
func mustCurrentAccount(t *testing.T) *user.User {
	t.Helper()
	account, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	return account
}

// TestVerifyUnixIdentityAcceptsCurrentAccount covers the Unix reader identity boundary.
func TestVerifyUnixIdentityAcceptsCurrentAccount(t *testing.T) {
	account := mustCurrentAccount(t)
	if account.Uid == "0" {
		t.Skip("root is intentionally rejected as a media reader")
	}
	got, err := VerifyUnixIdentity(Subject{Username: account.Username, UID: account.Uid})
	if err != nil {
		t.Fatal(err)
	}
	if got != account.Username {
		t.Fatalf("verified username = %q, want %q", got, account.Username)
	}
}

// TestVerifyUnixIdentityRejectsUnboundAccounts covers malformed and mismatched identities.
func TestVerifyUnixIdentityRejectsUnboundAccounts(t *testing.T) {
	account := mustCurrentAccount(t)
	for _, subject := range []Subject{
		{Username: account.Username},
		{UID: account.Uid},
		{Username: account.Username, UID: "4294967295"},
		{Username: "lite-nas-no-such-user", UID: account.Uid},
	} {
		t.Run(subject.Username+":"+subject.UID, func(t *testing.T) {
			_, err := VerifyUnixIdentity(subject)
			if !errors.Is(err, ErrUnixIdentityMismatch) {
				t.Fatalf("identity error = %v, want mismatch", err)
			}
		})
	}
}
