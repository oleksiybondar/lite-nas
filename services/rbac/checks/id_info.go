package checks

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	getfaclparser "lite-nas/shared/parsers/acl/getfacl"
)

var evalSymlinks = filepath.EvalSymlinks

// ResolveIdentityByUID resolves identity information for one UID.
func ResolveIdentityByUID(ctx context.Context, runner Runner, uid uint32) (Identity, error) {
	uidText := strconv.FormatUint(uint64(uid), 10)

	uidOutput, err := runner.Run(ctx, "id", "-u", uidText)
	if err != nil {
		return Identity{}, fmt.Errorf("id -u failed: %w", err)
	}

	gidOutput, err := runner.Run(ctx, "id", "-g", uidText)
	if err != nil {
		return Identity{}, fmt.Errorf("id -g failed: %w", err)
	}

	usernameOutput, err := runner.Run(ctx, "id", "-nu", uidText)
	if err != nil {
		return Identity{}, fmt.Errorf("id -nu failed: %w", err)
	}

	primaryGroupOutput, err := runner.Run(ctx, "id", "-ng", uidText)
	if err != nil {
		return Identity{}, fmt.Errorf("id -ng failed: %w", err)
	}

	groupsOutput, err := runner.Run(ctx, "id", "-Gn", uidText)
	if err != nil {
		return Identity{}, fmt.Errorf("id -Gn failed: %w", err)
	}

	return parseIdentity(
		string(uidOutput),
		string(gidOutput),
		string(usernameOutput),
		string(primaryGroupOutput),
		string(groupsOutput),
	)
}

func readACLForPath(ctx context.Context, runner Runner, path string) (getfaclparser.Document, error) {
	resolvedPath, err := evalSymlinks(path)
	if err != nil {
		return getfaclparser.Document{}, fmt.Errorf("resolve symlink target failed: %w", err)
	}

	output, err := runner.Run(ctx, "getfacl", "-p", resolvedPath)
	if err != nil {
		return getfaclparser.Document{}, fmt.Errorf("getfacl failed: %w", err)
	}

	document, err := getfaclparser.Parse(string(output))
	if err != nil {
		return getfaclparser.Document{}, fmt.Errorf("getfacl parse failed: %w", err)
	}

	return document, nil
}

func hasGroup(identity Identity, name string) bool {
	if identity.PrimaryGroup == name {
		return true
	}

	for _, group := range identity.Groups {
		if strings.TrimSpace(group) == name {
			return true
		}
	}

	return false
}

func applyMask(permission getfaclparser.Permission, mask *getfaclparser.Permission) getfaclparser.Permission {
	if mask == nil {
		return permission
	}

	return getfaclparser.Permission{
		Read:    permission.Read && mask.Read,
		Write:   permission.Write && mask.Write,
		Execute: permission.Execute && mask.Execute,
	}
}

func unionPermission(left getfaclparser.Permission, right getfaclparser.Permission) getfaclparser.Permission {
	return getfaclparser.Permission{
		Read:    left.Read || right.Read,
		Write:   left.Write || right.Write,
		Execute: left.Execute || right.Execute,
	}
}

func isPermissionAllowed(permission getfaclparser.Permission, operation string) bool {
	switch operation {
	case "read":
		return permission.Read
	case "write":
		return permission.Write
	case "execute":
		return permission.Execute
	default:
		return false
	}
}

func evaluateAccess(identity Identity, acl getfaclparser.Document, operation string) bool {
	if ownerPermission, ok := ownerPermission(identity, acl); ok {
		return isPermissionAllowed(ownerPermission, operation)
	}

	if namedUserPermission, ok := matchedNamedUserPermission(identity, acl); ok {
		return isPermissionAllowed(namedUserPermission, operation)
	}

	if groupPermissionAllows(identity, acl, operation) {
		return true
	}

	return isPermissionAllowed(acl.Other, operation)
}

func ownerPermission(identity Identity, acl getfaclparser.Document) (getfaclparser.Permission, bool) {
	if identity.Username != acl.Owner {
		return getfaclparser.Permission{}, false
	}
	return acl.User, true
}

func matchedNamedUserPermission(identity Identity, acl getfaclparser.Document) (getfaclparser.Permission, bool) {
	permission, ok := acl.NamedUsers[identity.Username]
	if !ok {
		return getfaclparser.Permission{}, false
	}
	return applyMask(permission, acl.Mask), true
}

func groupPermissionAllows(identity Identity, acl getfaclparser.Document, operation string) bool {
	maskedGroupPermission := applyMask(matchedGroupPermission(identity, acl), acl.Mask)
	return isPermissionAllowed(maskedGroupPermission, operation)
}

func matchedGroupPermission(identity Identity, acl getfaclparser.Document) getfaclparser.Permission {
	groupClassPermission := getfaclparser.Permission{}
	if hasGroup(identity, acl.Group) {
		groupClassPermission = unionPermission(groupClassPermission, acl.GroupObject)
	}
	for groupName, groupPermission := range acl.NamedGroups {
		if !hasGroup(identity, groupName) {
			continue
		}
		groupClassPermission = unionPermission(groupClassPermission, groupPermission)
	}
	return groupClassPermission
}
