package filesystem

import (
	"errors"
	"os"
	"os/user"
	"testing"

	"lite-nas/shared/fileaccess"
)

// TestVerifyUnixIdentityBindsUsernameAndUID covers file-service-svc/SR-001.
func TestVerifyUnixIdentityBindsUsernameAndUID(t *testing.T) {
	account, err := user.Current()
	if err != nil || os.Getuid() == 0 {
		t.Skip("requires a non-root host test user")
	}
	username, err := VerifyUnixIdentity(fileaccess.Subject{Username: account.Username, UID: account.Uid})
	if err != nil || username != account.Username {
		t.Fatalf("valid identity = %q, error = %v", username, err)
	}
	if _, err := VerifyUnixIdentity(fileaccess.Subject{Username: account.Username, UID: "999999999"}); !errors.Is(err, ErrUnixIdentityMismatch) {
		t.Fatalf("mismatched UID error = %v", err)
	}
}
