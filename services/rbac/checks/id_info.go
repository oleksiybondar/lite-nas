package checks

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	getfaclparser "lite-nas/shared/parsers/acl/getfacl"
	idparser "lite-nas/shared/parsers/id/identity"
)

var evalSymlinks = filepath.EvalSymlinks

// ResolveIdentityByUID resolves identity information for one UID.
func ResolveIdentityByUID(ctx context.Context, runner Runner, uid string) (idparser.Identity, error) {
	idOutput, err := runner.Run(ctx, "id", uid)
	if err != nil {
		return idparser.Identity{}, fmt.Errorf("id failed: %w", err)
	}

	identity, err := idparser.Parse(string(idOutput))
	if err != nil {
		return idparser.Identity{}, fmt.Errorf("id parse failed: %w", err)
	}

	return identity, nil
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

func hasGroup(identity idparser.Identity, name string) bool {
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

func evaluateAccess(identity idparser.Identity, acl getfaclparser.Document, operation string) bool {
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

func ownerPermission(identity idparser.Identity, acl getfaclparser.Document) (getfaclparser.Permission, bool) {
	if identity.Username != acl.Owner {
		return getfaclparser.Permission{}, false
	}
	return acl.User, true
}

func matchedNamedUserPermission(identity idparser.Identity, acl getfaclparser.Document) (getfaclparser.Permission, bool) {
	permission, ok := acl.NamedUsers[identity.Username]
	if !ok {
		return getfaclparser.Permission{}, false
	}
	return applyMask(permission, acl.Mask), true
}

func groupPermissionAllows(identity idparser.Identity, acl getfaclparser.Document, operation string) bool {
	maskedGroupPermission := applyMask(matchedGroupPermission(identity, acl), acl.Mask)
	return isPermissionAllowed(maskedGroupPermission, operation)
}

func matchedGroupPermission(identity idparser.Identity, acl getfaclparser.Document) getfaclparser.Permission {
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
