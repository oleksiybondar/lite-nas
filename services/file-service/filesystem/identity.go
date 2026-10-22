package filesystem

import (
	"errors"
	"os/user"
	"strconv"

	"lite-nas/shared/fileaccess"
)

// ErrUnixIdentityMismatch means Auth and RBAC did not identify the same Unix account.
var ErrUnixIdentityMismatch = errors.New("verified username and RBAC UID do not identify one Unix user")

// VerifyUnixIdentity binds the Auth username and RBAC UID before sudo invocation.
// The returned username is read from the host account database, not HTTP input.
func VerifyUnixIdentity(subject fileaccess.Subject) (string, error) {
	if subject.Username == "" || subject.UID == "" {
		return "", ErrUnixIdentityMismatch
	}
	account, err := user.Lookup(subject.Username)
	if err != nil {
		return "", ErrUnixIdentityMismatch
	}
	uid, err := strconv.ParseUint(account.Uid, 10, 32)
	if err != nil || uid == 0 || strconv.FormatUint(uid, 10) != subject.UID || account.Username != subject.Username {
		return "", ErrUnixIdentityMismatch
	}
	return account.Username, nil
}
