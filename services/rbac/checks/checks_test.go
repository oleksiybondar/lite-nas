package checks

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"testing"

	getfaclparser "lite-nas/shared/parsers/acl/getfacl"
)

func TestCanReadUsesOwnerPermission(t *testing.T) {
	t.Parallel()

	runCanAccessTest(t, func(ctx context.Context, runner Runner) (bool, error) {
		return CanRead(ctx, runner, 1002, "/input/path")
	}, append(identityCalls("1002", "testuser", "testgroup", "testgroup wheel"),
		scriptedCall{
			name:   "getfacl",
			args:   []string{"-p", "/resolved/path"},
			output: "# file: /resolved/path\n# owner: testuser\n# group: lite-nas\nuser::rwx\ngroup::--x\nother::---\n",
		},
	), true)
}

func TestCanWriteUsesGroupPermissionMaskedByMask(t *testing.T) {
	t.Parallel()

	runner := newScriptedRunner(append(identityCalls("1002", "testuser", "ops", "ops lite-nas"),
		scriptedCall{
			name:   "getfacl",
			args:   []string{"-p", "/resolved/path"},
			output: "# owner: root\n# group: lite-nas\nuser::rwx\ngroup::rwx\nmask::r-x\nother::---\n",
		},
	))

	withResolvedPath(t, "/resolved/path", func() {
		allowed, err := CanWrite(t.Context(), runner, 1002, "/input/path")
		if err != nil {
			t.Fatalf("CanWrite() error = %v", err)
		}
		if allowed {
			t.Fatalf("CanWrite() = true, want false because mask removes write")
		}
	})
}

func TestCanExecUsesOtherPermissionWhenNoOwnerOrGroupMatch(t *testing.T) {
	t.Parallel()

	runCanAccessTest(t, func(ctx context.Context, runner Runner) (bool, error) {
		return CanExec(ctx, runner, 1002, "/input/path")
	}, append(identityCalls("1002", "testuser", "testgroup", "testgroup"),
		scriptedCall{
			name:   "getfacl",
			args:   []string{"-p", "/resolved/path"},
			output: "# owner: root\n# group: lite-nas\nuser::rwx\ngroup::--x\nother::--x\n",
		},
	), true)
}

func TestCanSudoExecReturnsTrueOnSuccess(t *testing.T) {
	t.Parallel()

	runner := newScriptedRunner(append(identityCalls("1002", "testuser", "testgroup", "testgroup wheel"),
		scriptedCall{name: "sudo", args: []string{"-n", "-l", "-U", "testuser", "/usr/bin/zfs"}},
	))

	assertCanSudoExecAllowed(t, runner, true)
}

func TestCanSudoExecReturnsFalseOnExitError(t *testing.T) {
	t.Parallel()

	runner := newScriptedRunner(append(identityCalls("1002", "testuser", "testgroup", "testgroup wheel"),
		scriptedCall{name: "sudo", args: []string{"-n", "-l", "-U", "testuser", "/usr/bin/zfs"}, err: &exec.ExitError{}},
	))

	assertCanSudoExecAllowed(t, runner, false)
}

func TestCanSudoExecReturnsErrorOnRunnerFailure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("runner unavailable")
	runner := newScriptedRunner(append(identityCalls("1002", "testuser", "testgroup", "testgroup wheel"),
		scriptedCall{name: "sudo", args: []string{"-n", "-l", "-U", "testuser", "/usr/bin/zfs"}, err: wantErr},
	))

	assertCanSudoExecFails(t, runner)
}

func TestSharedParseACLDataParsesNamedEntriesAndMask(t *testing.T) {
	t.Parallel()

	acl, err := getfaclparser.Parse(
		"# owner: root\n" +
			"# group: lite-nas\n" +
			"user::rwx\n" +
			"user:testuser:r--\n" +
			"group::r-x\n" +
			"group:admins:rwx\n" +
			"mask::r-x\n" +
			"other::---\n",
	)
	if err != nil {
		t.Fatalf("ParseACLData() error = %v", err)
	}

	assertACLHasOwnerGroup(t, acl, "root", "lite-nas")
	assertMaskRXOnly(t, acl)
	assertNamedEntryExists(t, acl.NamedUsers, "testuser", "named user")
	assertNamedEntryExists(t, acl.NamedGroups, "admins", "named group")
}

func TestParseIdentityParsesAllFields(t *testing.T) {
	t.Parallel()

	identity, err := parseIdentity("1002\n", "1002\n", "testuser\n", "testgroup\n", "testgroup wheel\n")
	if err != nil {
		t.Fatalf("parseIdentity() error = %v", err)
	}
	assertIdentityCoreFields(t, identity, 1002, 1002, "testuser", "testgroup")
	assertIdentityGroups(t, identity, []string{"testgroup", "wheel"})
}

func withResolvedPath(t *testing.T, resolvedPath string, run func()) {
	t.Helper()

	original := evalSymlinks
	evalSymlinks = func(string) (string, error) {
		return resolvedPath, nil
	}
	defer func() {
		evalSymlinks = original
	}()

	run()
}

func identityCalls(uid string, username string, primaryGroup string, groups string) []scriptedCall {
	return []scriptedCall{
		{name: "id", args: []string{"-u", uid}, output: uid + "\n"},
		{name: "id", args: []string{"-g", uid}, output: uid + "\n"},
		{name: "id", args: []string{"-nu", uid}, output: username + "\n"},
		{name: "id", args: []string{"-ng", uid}, output: primaryGroup + "\n"},
		{name: "id", args: []string{"-Gn", uid}, output: groups + "\n"},
	}
}

func assertCanSudoExecFails(t *testing.T, runner Runner) {
	t.Helper()

	allowed, err := CanSudoExec(t.Context(), runner, 1002, "/usr/bin/zfs")
	if err == nil {
		t.Fatalf("CanSudoExec() error = nil, want non-nil")
	}
	if allowed {
		t.Fatalf("CanSudoExec() = true, want false")
	}
}

func assertCanSudoExecAllowed(t *testing.T, runner Runner, wantAllowed bool) {
	t.Helper()

	allowed, err := CanSudoExec(t.Context(), runner, 1002, "/usr/bin/zfs")
	if err != nil {
		t.Fatalf("CanSudoExec() error = %v", err)
	}
	if allowed != wantAllowed {
		t.Fatalf("CanSudoExec() = %v, want %v", allowed, wantAllowed)
	}
}

func runCanAccessTest(
	t *testing.T,
	check func(ctx context.Context, runner Runner) (bool, error),
	calls []scriptedCall,
	wantAllowed bool,
) {
	t.Helper()

	runner := newScriptedRunner(calls)
	withResolvedPath(t, "/resolved/path", func() {
		allowed, err := check(t.Context(), runner)
		if err != nil {
			t.Fatalf("check() error = %v", err)
		}
		if allowed != wantAllowed {
			t.Fatalf("check() = %v, want %v", allowed, wantAllowed)
		}
	})
}

func assertACLHasOwnerGroup(t *testing.T, acl getfaclparser.Document, owner string, group string) {
	t.Helper()
	if acl.Owner != owner || acl.Group != group {
		t.Fatalf("owner/group parse mismatch: %#v", acl)
	}
}

func assertMaskRXOnly(t *testing.T, acl getfaclparser.Document) {
	t.Helper()
	if acl.Mask == nil || !acl.Mask.Read || acl.Mask.Write || !acl.Mask.Execute {
		t.Fatalf("mask parse mismatch: %#v", acl.Mask)
	}
}

func assertNamedEntryExists(t *testing.T, entries map[string]getfaclparser.Permission, name string, label string) {
	t.Helper()
	if _, ok := entries[name]; !ok {
		t.Fatalf("%s entry missing", label)
	}
}

func assertIdentityCoreFields(t *testing.T, identity Identity, uid uint32, gid uint32, username string, primaryGroup string) {
	t.Helper()
	if identity.UID != uid || identity.GID != gid || identity.Username != username || identity.PrimaryGroup != primaryGroup {
		t.Fatalf("identity parse mismatch: %#v", identity)
	}
}

func assertIdentityGroups(t *testing.T, identity Identity, wantGroups []string) {
	t.Helper()
	if !reflect.DeepEqual(identity.Groups, wantGroups) {
		t.Fatalf("groups parse mismatch: %#v", identity.Groups)
	}
}

type scriptedCall struct {
	name   string
	args   []string
	output string
	err    error
}

type scriptedRunner struct {
	calls []scriptedCall
	index int
}

func newScriptedRunner(calls []scriptedCall) *scriptedRunner {
	return &scriptedRunner{calls: calls}
}

func (runner *scriptedRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	if runner.index >= len(runner.calls) {
		return nil, errors.New("unexpected command call")
	}

	call := runner.calls[runner.index]
	runner.index++

	if call.name != name {
		return nil, errors.New("unexpected command name")
	}
	if !reflect.DeepEqual(call.args, args) {
		return nil, errors.New("unexpected command args")
	}

	return []byte(call.output), call.err
}
