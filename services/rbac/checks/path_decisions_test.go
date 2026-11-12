package checks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const directoryACL = "# owner: root\n# group: root\nuser::rwx\ngroup::--x\nother::--x\n"

// pathFixtureRunner supplies controlled identity and ACL data for real temporary paths.
type pathFixtureRunner struct {
	uid     string
	leaf    string
	leafACL string
	acls    map[string]string
	fail    string
}

// Run implements the host-command boundary without depending on host ACLs.
func (runner pathFixtureRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	uid := runner.uid
	if uid == "" {
		uid = "1002"
	}
	if name == "id" && len(args) == 2 && args[0] == "--" && args[1] == uid {
		return []byte("uid=" + uid + "(testuser) gid=1002(testgroup) groups=1002(testgroup)\n"), nil
	}
	if name != "getfacl" || len(args) != 3 || args[0] != "-p" || args[1] != "--" {
		return nil, errors.New("unexpected command")
	}
	if args[2] == runner.fail {
		return nil, errors.New("ACL unavailable")
	}
	if acl, ok := runner.acls[args[2]]; ok {
		return []byte(acl), nil
	}
	if args[2] == runner.leaf {
		return []byte(runner.leafACL), nil
	}
	return []byte(directoryACL), nil
}

// mustTestFile creates a regular inode for a path decision fixture.
func mustTestFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestPathReadDeniesUnsearchableAncestor covers rbac-service/FR-002 traversal.
func TestPathReadDeniesUnsearchableAncestor(t *testing.T) {
	path := mustTestFile(t)
	parent := filepath.Dir(path)
	runner := pathFixtureRunner{leaf: path, leafACL: "# owner: testuser\n# group: testgroup\nuser::rwx\ngroup::---\nother::---\n", acls: map[string]string{
		parent: "# owner: root\n# group: root\nuser::rwx\ngroup::---\nother::---\n",
	}}
	allowed, err := CanRead(t.Context(), runner, "1002", path)
	if err != nil || allowed {
		t.Fatalf("CanRead() = %v, %v; want deny", allowed, err)
	}
}

// TestCreateAllowsMissingTarget covers rbac-service/FR-002 creation under a writable parent.
func TestCreateAllowsMissingTarget(t *testing.T) {
	parent := t.TempDir()
	path := filepath.Join(parent, "new")
	runner := pathFixtureRunner{acls: map[string]string{
		parent: "# owner: testuser\n# group: testgroup\nuser::rwx\ngroup::---\nother::---\n",
	}}
	allowed, err := CheckPathOperation(t.Context(), runner, "1002", path, "create")
	if err != nil || !allowed {
		t.Fatalf("create = %v, %v; want allow", allowed, err)
	}
	allowed, err = CanWrite(t.Context(), runner, "1002", path)
	if err == nil || allowed {
		t.Fatalf("write missing = %v, %v; want deny", allowed, err)
	}
}

// TestExistingFileWriteStillAllowsOwner covers rbac-service/FR-002 legacy write RPC.
func TestExistingFileWriteStillAllowsOwner(t *testing.T) {
	path := mustTestFile(t)
	runner := pathFixtureRunner{leaf: path, leafACL: "# owner: testuser\n# group: testgroup\nuser::rw-\ngroup::---\nother::---\n"}
	allowed, err := CanWrite(t.Context(), runner, "1002", path)
	if err != nil || !allowed {
		t.Fatalf("existing-file write = %v, %v; want allow", allowed, err)
	}
}

// TestReplaceAndDeleteUseDirectoryRights covers rbac-service/FR-002 entry mutation.
func TestReplaceAndDeleteUseDirectoryRights(t *testing.T) {
	path := mustTestFile(t)
	runner := pathFixtureRunner{leaf: path, leafACL: "# owner: root\n# group: root\nuser::r--\ngroup::---\nother::---\n", acls: map[string]string{
		filepath.Dir(path): "# owner: testuser\n# group: testgroup\nuser::rwx\ngroup::---\nother::---\n",
	}}
	for _, operation := range []string{"replace", "delete"} {
		allowed, err := CheckPathOperation(t.Context(), runner, "1002", path, operation)
		if err != nil || !allowed {
			t.Fatalf("%s = %v, %v; want allow", operation, allowed, err)
		}
	}
}

// TestMutationRejectsSymlinkAndDirectory covers rbac-service/FR-002 unsupported targets.
func TestMutationRejectsSymlinkAndDirectory(t *testing.T) {
	path := mustTestFile(t)
	link := filepath.Join(filepath.Dir(path), "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	runner := pathFixtureRunner{acls: map[string]string{
		filepath.Dir(path): "# owner: testuser\n# group: testgroup\nuser::rwx\ngroup::---\nother::---\n",
	}}
	for _, target := range []string{link, filepath.Dir(path)} {
		for _, operation := range []string{"replace", "delete"} {
			allowed, err := CheckPathOperation(t.Context(), runner, "1002", target, operation)
			if err == nil && allowed {
				t.Fatalf("%s %s unexpectedly allowed", operation, target)
			}
		}
	}
}

// TestStickyDirectoryDeniesOtherOwner covers rbac-service/FR-002 sticky-directory rules.
func TestStickyDirectoryDeniesOtherOwner(t *testing.T) {
	parent := t.TempDir()
	if err := os.Chmod(parent, 0o777|os.ModeSticky); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "file")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := pathFixtureRunner{uid: "424242", acls: map[string]string{
		parent: "# owner: testuser\n# group: testgroup\nuser::rwx\ngroup::rwx\nother::rwx\n",
	}}
	allowed, err := CheckPathOperation(t.Context(), runner, "424242", path, "delete")
	if err != nil || allowed {
		t.Fatalf("sticky delete = %v, %v; want deny", allowed, err)
	}
}

// TestStickyDirectoryAllowsTargetOwner covers rbac-service/FR-002 owner exception.
func TestStickyDirectoryAllowsTargetOwner(t *testing.T) {
	parent := t.TempDir()
	if err := os.Chmod(parent, 0o777|os.ModeSticky); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "file")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	uid := os.Getuid()
	if uid == 0 {
		uid = 1002
		if err := os.Chown(path, uid, -1); err != nil {
			t.Skipf("cannot provision sticky-file owner: %v", err)
		}
	}
	runner := pathFixtureRunner{uid: strconv.Itoa(uid), acls: map[string]string{
		parent: "# owner: root\n# group: root\nuser::rwx\ngroup::rwx\nother::rwx\n",
	}}
	allowed, err := CheckPathOperation(t.Context(), runner, strconv.Itoa(uid), path, "delete")
	if err != nil || !allowed {
		t.Fatalf("sticky delete by file owner = %v, %v; want allow", allowed, err)
	}
}

// TestTraverseChecksTargetDirectory covers rbac-service/FR-002 explicit traversal.
func TestTraverseChecksTargetDirectory(t *testing.T) {
	path := t.TempDir()
	runner := pathFixtureRunner{acls: map[string]string{
		path: "# owner: testuser\n# group: testgroup\nuser::r-x\ngroup::---\nother::---\n",
	}}
	allowed, err := CheckPathOperation(t.Context(), runner, "1002", path, "traverse")
	if err != nil || !allowed {
		t.Fatalf("traverse = %v, %v; want allow", allowed, err)
	}
}

// TestPathDecisionFailsClosedOnACLFailure covers rbac-service/RR-001.
func TestPathDecisionFailsClosedOnACLFailure(t *testing.T) {
	path := mustTestFile(t)
	runner := pathFixtureRunner{fail: filepath.Dir(path)}
	allowed, err := CheckPathOperation(t.Context(), runner, "1002", path, "delete")
	if err == nil || allowed {
		t.Fatalf("delete ACL failure = %v, %v; want deny and error", allowed, err)
	}
}

// TestSymlinkTargetAncestorMustBeSearchable covers rbac-service/FR-002 resolved traversal.
func TestSymlinkTargetAncestorMustBeSearchable(t *testing.T) {
	root := t.TempDir()
	targetDirectory := filepath.Join(root, "target")
	if err := os.Mkdir(targetDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(targetDirectory, "file")
	if err := os.WriteFile(target, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(targetDirectory, link); err != nil {
		t.Fatal(err)
	}
	runner := pathFixtureRunner{leaf: target, leafACL: "# owner: testuser\n# group: testgroup\nuser::rwx\ngroup::---\nother::---\n", acls: map[string]string{
		targetDirectory: "# owner: root\n# group: root\nuser::rwx\ngroup::---\nother::---\n",
	}}
	allowed, err := CanRead(t.Context(), runner, "1002", filepath.Join(link, "file"))
	if err != nil || allowed {
		t.Fatalf("read through unsearchable target = %v, %v; want deny", allowed, err)
	}
}

// TestReadFollowsSafeFileSymlink covers rbac-service/FR-002 final symlink resolution.
func TestReadFollowsSafeFileSymlink(t *testing.T) {
	path := mustTestFile(t)
	link := filepath.Join(filepath.Dir(path), "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	runner := pathFixtureRunner{leaf: path, leafACL: "# owner: testuser\n# group: testgroup\nuser::r--\ngroup::---\nother::---\n"}
	allowed, err := CanRead(t.Context(), runner, "1002", link)
	if err != nil || !allowed {
		t.Fatalf("symlink read = %v, %v; want allow", allowed, err)
	}
}

// TestMissingParentDeniesCreate covers rbac-service/RR-001 required ancestor failure.
func TestMissingParentDeniesCreate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "file")
	allowed, err := CheckPathOperation(t.Context(), pathFixtureRunner{}, "1002", path, "create")
	if err == nil || allowed {
		t.Fatalf("create with missing parent = %v, %v; want deny and error", allowed, err)
	}
}

// TestInvalidPathAndUIDDeny covers rbac-service/RR-001 malformed identity and path inputs.
func TestInvalidPathAndUIDDeny(t *testing.T) {
	path := mustTestFile(t)
	for _, request := range []struct{ uid, path string }{
		{"0", path},
		{"bad", path},
		{"1002", "relative"},
		{"1002", filepath.Dir(path) + "/../" + filepath.Base(path)},
	} {
		allowed, err := CanRead(t.Context(), pathFixtureRunner{}, request.uid, request.path)
		if err == nil || allowed {
			t.Fatalf("CanRead(%q, %q) = %v, %v; want deny and error", request.uid, request.path, allowed, err)
		}
	}
}
