package filesystem

import "lite-nas/shared/fileaccess"

// ErrUnixIdentityMismatch means Auth and RBAC did not identify the same Unix account.
var ErrUnixIdentityMismatch = fileaccess.ErrUnixIdentityMismatch

// VerifyUnixIdentity binds the Auth username and RBAC UID before sudo invocation.
// The returned username is read from the host account database, not HTTP input.
func VerifyUnixIdentity(subject fileaccess.Subject) (string, error) {
	return fileaccess.VerifyUnixIdentity(subject)
}
