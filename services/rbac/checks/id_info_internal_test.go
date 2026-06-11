package checks

import (
	"testing"

	getfaclparser "lite-nas/shared/parsers/acl/getfacl"
	idparser "lite-nas/shared/parsers/id/identity"
)

func TestEvaluateAccessOwnerBranch(t *testing.T) {
	t.Parallel()

	identity := idparser.Identity{Username: "owner", PrimaryGroup: "staff", Groups: []string{"staff"}}
	acl := getfaclparser.Document{
		Owner:       "owner",
		Group:       "staff",
		User:        getfaclparser.Permission{Read: true, Write: false, Execute: true},
		GroupObject: getfaclparser.Permission{Read: false, Write: false, Execute: false},
		Other:       getfaclparser.Permission{Read: false, Write: false, Execute: false},
		NamedUsers:  map[string]getfaclparser.Permission{},
		NamedGroups: map[string]getfaclparser.Permission{},
	}

	if !evaluateAccess(identity, acl, "read") {
		t.Fatalf("evaluateAccess() expected owner read to be allowed")
	}
	if evaluateAccess(identity, acl, "write") {
		t.Fatalf("evaluateAccess() expected owner write to be denied")
	}
}

func TestEvaluateAccessNamedUserAndMask(t *testing.T) {
	t.Parallel()

	identity := idparser.Identity{Username: "alice", PrimaryGroup: "staff", Groups: []string{"staff"}}
	acl := getfaclparser.Document{
		Owner:       "owner",
		Group:       "staff",
		User:        getfaclparser.Permission{Read: true, Write: true, Execute: true},
		GroupObject: getfaclparser.Permission{Read: false, Write: false, Execute: false},
		Other:       getfaclparser.Permission{Read: false, Write: false, Execute: false},
		Mask:        &getfaclparser.Permission{Read: true, Write: false, Execute: false},
		NamedUsers: map[string]getfaclparser.Permission{
			"alice": {Read: true, Write: true, Execute: true},
		},
		NamedGroups: map[string]getfaclparser.Permission{},
	}

	if !evaluateAccess(identity, acl, "read") {
		t.Fatalf("evaluateAccess() expected masked named user read to be allowed")
	}
	if evaluateAccess(identity, acl, "write") {
		t.Fatalf("evaluateAccess() expected masked named user write to be denied")
	}
}

func TestEvaluateAccessGroupAndOtherBranches(t *testing.T) {
	t.Parallel()

	identity := idparser.Identity{Username: "bob", PrimaryGroup: "dev", Groups: []string{"dev", "ops"}}
	acl := getfaclparser.Document{
		Owner:       "owner",
		Group:       "dev",
		User:        getfaclparser.Permission{Read: true, Write: true, Execute: true},
		GroupObject: getfaclparser.Permission{Read: true, Write: false, Execute: false},
		Other:       getfaclparser.Permission{Read: false, Write: false, Execute: true},
		NamedUsers:  map[string]getfaclparser.Permission{},
		NamedGroups: map[string]getfaclparser.Permission{"ops": {Read: false, Write: true, Execute: false}},
	}

	if !evaluateAccess(identity, acl, "write") {
		t.Fatalf("evaluateAccess() expected named group write to be allowed")
	}
	if !evaluateAccess(identity, acl, "execute") {
		t.Fatalf("evaluateAccess() expected other execute fallback to be allowed")
	}

	nonMember := idparser.Identity{Username: "charlie", PrimaryGroup: "none", Groups: []string{"none"}}
	if !evaluateAccess(nonMember, acl, "execute") {
		t.Fatalf("evaluateAccess() expected other execute to be allowed")
	}
}

func TestHasGroup(t *testing.T) {
	t.Parallel()

	identity := idparser.Identity{PrimaryGroup: "staff", Groups: []string{"wheel"}}
	if !hasGroup(identity, "staff") || !hasGroup(identity, "wheel") || hasGroup(identity, "nogroup") {
		t.Fatalf("hasGroup() returned unexpected result")
	}
}

func TestApplyMask(t *testing.T) {
	t.Parallel()

	permission := getfaclparser.Permission{Read: true, Write: true, Execute: false}
	mask := getfaclparser.Permission{Read: true, Write: false, Execute: true}
	masked := applyMask(permission, &mask)
	if !masked.Read || masked.Write || masked.Execute {
		t.Fatalf("applyMask() returned unexpected masked permission: %#v", masked)
	}
}

func TestUnionPermission(t *testing.T) {
	t.Parallel()

	union := unionPermission(
		getfaclparser.Permission{Read: true, Write: false, Execute: false},
		getfaclparser.Permission{Read: false, Write: true, Execute: false},
	)
	if !union.Read || !union.Write || union.Execute {
		t.Fatalf("unionPermission() returned unexpected permission: %#v", union)
	}
}

func TestIsPermissionAllowed(t *testing.T) {
	t.Parallel()

	if !isPermissionAllowed(getfaclparser.Permission{Execute: true}, "execute") {
		t.Fatalf("isPermissionAllowed() expected execute true")
	}
	if isPermissionAllowed(getfaclparser.Permission{Execute: true}, "unknown") {
		t.Fatalf("isPermissionAllowed() expected unknown operation false")
	}
}

func TestIsInteractiveShell(t *testing.T) {
	t.Parallel()

	if isInteractiveShell("/sbin/nologin") {
		t.Fatalf("isInteractiveShell() expected nologin false")
	}
	if isInteractiveShell("/bin/false") {
		t.Fatalf("isInteractiveShell() expected false-shell false")
	}
	if !isInteractiveShell("/bin/bash") {
		t.Fatalf("isInteractiveShell() expected bash true")
	}
}
