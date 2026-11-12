package checks

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// ValidateIdentityOperand accepts one username or decimal UID as an id operand.
func ValidateIdentityOperand(value string) error {
	if value == "" || len(value) > 128 || strings.HasPrefix(value, "-") ||
		strings.IndexFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return errors.New("invalid identity operand")
	}
	return nil
}

// ValidateSudoArguments accepts a non-root UID and one absolute executable identity.
func ValidateSudoArguments(uid string, command string) error {
	if err := validateNumericUID(uid); err != nil {
		return err
	}
	return validateCommandPath(command)
}

// validateNumericUID accepts a non-root decimal UID used by RBAC host checks.
func validateNumericUID(value string) error {
	uid, err := strconv.ParseUint(value, 10, 32)
	if err != nil || uid == 0 || strconv.FormatUint(uid, 10) != value {
		return errors.New("UID must be a non-root decimal number")
	}
	return nil
}

// validateCommandPath accepts a single clean absolute executable identity.
func validateCommandPath(path string) error {
	if path == "" || len(path) > 4096 || !filepath.IsAbs(path) || filepath.Clean(path) != path ||
		strings.IndexByte(path, 0) >= 0 ||
		strings.IndexFunc(path, unicode.IsControl) >= 0 {
		return errors.New("command must be one clean absolute path")
	}
	return nil
}

// validateACLPath accepts a clean absolute pathname as a getfacl operand.
func validateACLPath(path string) error {
	if path == "" || len(path) > 4096 || !filepath.IsAbs(path) || filepath.Clean(path) != path ||
		strings.IndexByte(path, 0) >= 0 {
		return errors.New("ACL path must be clean and absolute")
	}
	return nil
}
