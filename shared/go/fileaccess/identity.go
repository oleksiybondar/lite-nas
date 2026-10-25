package fileaccess

import (
	"errors"
	"os/user"
	"strconv"
)

// ErrUnixIdentityMismatch means the verified username and RBAC UID disagree.
var ErrUnixIdentityMismatch = errors.New("verified username and RBAC UID do not identify one Unix user")

// VerifyUnixIdentity binds an Auth username and RBAC UID to one host account.
// Only the returned host username may be passed as a sudo user operand.
func VerifyUnixIdentity(subject Subject) (string, error) {
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
