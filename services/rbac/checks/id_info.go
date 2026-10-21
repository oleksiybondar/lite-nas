package checks

import (
	"context"
	"fmt"
	"strings"

	getfaclparser "lite-nas/shared/parsers/acl/getfacl"
	idparser "lite-nas/shared/parsers/id/identity"
)

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

// readACLForResolvedPath reads the access ACL of an already resolved inode.
func readACLForResolvedPath(ctx context.Context, runner Runner, resolvedPath string) (getfaclparser.Document, error) {
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

	if permission, matched := groupPermission(identity, acl); matched {
		return isPermissionAllowed(permission, operation)
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

// groupPermission reports whether any owning or named group entry matches.
// A matching group with insufficient rights must not fall through to other.
func groupPermission(identity idparser.Identity, acl getfaclparser.Document) (getfaclparser.Permission, bool) {
	matched := hasGroup(identity, acl.Group)
	for name := range acl.NamedGroups {
		matched = matched || hasGroup(identity, name)
	}
	maskedGroupPermission := applyMask(matchedGroupPermission(identity, acl), acl.Mask)
	return maskedGroupPermission, matched
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
