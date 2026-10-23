package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// TestRunShowsHelp verifies that callers can discover the operation contract.
// Requirements: file-writer/FR-005.
func TestRunShowsHelp(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "no arguments"},
		{name: "long help", args: []string{"--help"}},
		{name: "short help", args: []string{"-h"}},
		{name: "command help", args: []string{"write", "--help"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			if err := run(tc.args, bytes.NewReader(nil), &output); err != nil {
				t.Fatalf("show help: %v", err)
			}
			if !strings.Contains(output.String(), "Usage:\n") || !strings.Contains(output.String(), "checksum") {
				t.Fatalf("help is missing operation usage: %q", output.String())
			}
		})
	}
}

// TestInvocationRejectsInvalidFlags verifies malformed CLI requests stop before filesystem work.
// Requirements: file-writer/FR-005.
func TestInvocationRejectsInvalidFlags(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "unknown command", args: []string{"unknown", "--target", "/tmp/a.upload"}},
		{name: "unknown flag", args: []string{"create", "--target", "/tmp/a.upload", "--bogus"}},
		{name: "relative target", args: []string{"create", "--target", "a.upload"}},
		{name: "unclean target", args: []string{"create", "--target", "/tmp/../tmp/a.upload"}},
		{name: "wrong suffix", args: []string{"create", "--target", "/tmp/a"}},
		{name: "missing identity", args: []string{"stat", "--target", "/tmp/a.upload"}},
		{name: "missing write length", args: []string{"write", "--target", "/tmp/a.upload", "--expected-device", "1", "--expected-inode", "1", "--expected-uid", "1", "--offset", "0"}},
		{name: "missing truncate offset", args: []string{"truncate", "--target", "/tmp/a.upload", "--expected-device", "1", "--expected-inode", "1", "--expected-uid", "1"}},
		{name: "missing destination mode", args: []string{"publish", "--target", "/tmp/a.upload", "--expected-device", "1", "--expected-inode", "1", "--expected-uid", "1"}},
		{name: "partial destination identity", args: []string{"publish", "--target", "/tmp/a.upload", "--expected-device", "1", "--expected-inode", "1", "--expected-uid", "1", "--destination-device", "1"}},
		{name: "invalid publication mode", args: []string{"publish", "--target", "/tmp/a.upload", "--expected-device", "1", "--expected-inode", "1", "--expected-uid", "1", "--destination-absent", "--mode", "0777"}},
		{name: "mode on create", args: []string{"create", "--target", "/tmp/a.upload", "--mode", "0664"}},
		{name: "disallowed flag", args: []string{"create", "--target", "/tmp/a.upload", "--offset", "0"}},
		{name: "extra argument", args: []string{"create", "--target", "/tmp/a.upload", "extra"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := parseInvocation(tc.args); err == nil {
				t.Fatalf("accepted invalid invocation: %v", tc.args)
			}
		})
	}
	if got := (&operationError{Message: "publication failed"}).Error(); got != "publication failed" {
		t.Fatalf("operation error = %q", got)
	}
}

// TestCreateReturnsIdentity verifies creation reports the effective owner's new inode.
// Requirements: file-writer/FR-001.
func TestCreateReturnsIdentity(t *testing.T) {
	t.Parallel()
	stage := filepath.Join(t.TempDir(), "target.upload")
	created := mustRun(t, []string{"create", "--target", stage}, nil)
	if created.UID != uint32(os.Geteuid()) || created.Inode == 0 {
		t.Fatalf("created stage identity = %+v", created)
	}
}

// TestWriteReportsNextOffset verifies a successful write reports its durable offset.
// Requirements: file-writer/FR-002.
func TestWriteReportsNextOffset(t *testing.T) {
	t.Parallel()
	stage, identity := mustStageFixture(t)
	written := mustRun(t, append([]string{"write", "--target", stage, "--offset", "0", "--length", "5"}, identity...), []byte("ABCDE"))
	if written.BytesWritten != 5 || written.NextOffset != 5 {
		t.Fatalf("write result = %+v", written)
	}
}

// TestStatReportsSize verifies the identified stage's observed length.
// Requirements: file-writer/FR-004.
func TestStatReportsSize(t *testing.T) {
	t.Parallel()
	stage, identity := mustStageFixture(t)
	mustRun(t, append([]string{"write", "--target", stage, "--offset", "0", "--length", "5"}, identity...), []byte("ABCDE"))
	observed := mustRun(t, append([]string{"stat", "--target", stage}, identity...), nil)
	if observed.Size != 5 {
		t.Fatalf("stage size = %d, want 5", observed.Size)
	}
}

// TestChecksumReportsSHA256 verifies the staged bytes are read by the writer.
// Requirements: file-writer/FR-004.
func TestChecksumReportsSHA256(t *testing.T) {
	t.Parallel()
	stage, identity := mustStageFixture(t)
	mustRun(t, append([]string{"write", "--target", stage, "--offset", "0", "--length", "5"}, identity...), []byte("ABCDE"))
	checksum := mustRun(t, append([]string{"checksum", "--target", stage}, identity...), nil)
	wantHash := sha256.Sum256([]byte("ABCDE"))
	if checksum.SHA256 != hex.EncodeToString(wantHash[:]) {
		t.Fatalf("checksum = %q, want %x", checksum.SHA256, wantHash)
	}
}

// TestTruncateRecoversCommittedOffset verifies uncommitted bytes are discarded.
// Requirements: file-writer/FR-004.
func TestTruncateRecoversCommittedOffset(t *testing.T) {
	t.Parallel()
	stage, identity := mustStageFixture(t)
	mustRun(t, append([]string{"write", "--target", stage, "--offset", "0", "--length", "5"}, identity...), []byte("ABCDE"))
	mustRun(t, append([]string{"truncate", "--target", stage, "--offset", "3"}, identity...), nil)
	if got := string(mustReadFile(t, stage)); got != "ABC" {
		t.Fatalf("truncated stage = %q, want ABC", got)
	}
}

// TestCleanupRemovesIdentifiedStage verifies cancellation removes the stage entry.
// Requirements: file-writer/FR-007.
func TestCleanupRemovesIdentifiedStage(t *testing.T) {
	t.Parallel()
	stage, identity := mustStageFixture(t)
	mustRun(t, append([]string{"cleanup", "--target", stage}, identity...), nil)
	if _, err := os.Lstat(stage); !os.IsNotExist(err) {
		t.Fatalf("stage remains after cleanup: %v", err)
	}
}

// TestWriteReplacesPartialTail verifies a retry does not keep uncommitted bytes.
// Requirements: file-writer/FR-002.
func TestWriteReplacesPartialTail(t *testing.T) {
	t.Parallel()
	stage := filepath.Join(t.TempDir(), "target.upload")
	created := mustRun(t, []string{"create", "--target", stage}, nil)
	identity := identityArgs(created)
	mustRun(t, append([]string{"write", "--target", stage, "--offset", "0", "--length", "5"}, identity...), []byte("ABCxy"))
	mustRun(t, append([]string{"write", "--target", stage, "--offset", "3", "--length", "4"}, identity...), []byte("DEFG"))
	if got := string(mustReadFile(t, stage)); got != "ABCDEFG" {
		t.Fatalf("retried stage = %q, want ABCDEFG", got)
	}
}

// TestWriteRecoversShortInput verifies incomplete streams leave the committed prefix.
// Requirements: file-writer/FR-002.
func TestWriteRecoversShortInput(t *testing.T) {
	t.Parallel()
	stage := filepath.Join(t.TempDir(), "target.upload")
	identity := identityArgs(mustRun(t, []string{"create", "--target", stage}, nil))
	mustRun(t, append([]string{"write", "--target", stage, "--offset", "0", "--length", "3"}, identity...), []byte("ABC"))
	var output bytes.Buffer
	err := run(append([]string{"write", "--target", stage, "--offset", "3", "--length", "2"}, identity...), bytes.NewReader([]byte("D")), &output)
	if err == nil {
		t.Fatal("expected incomplete input error")
	}
	if got := string(mustReadFile(t, stage)); got != "ABC" {
		t.Fatalf("recovered stage = %q, want ABC", got)
	}
}

// TestIdentityMismatchRejectsReplacement verifies a path cannot silently bind a new inode.
// Requirements: file-writer/FR-003.
func TestIdentityMismatchRejectsReplacement(t *testing.T) {
	t.Parallel()
	stage := filepath.Join(t.TempDir(), "target.upload")
	identity := identityArgs(mustRun(t, []string{"create", "--target", stage}, nil))
	if err := os.Rename(stage, stage+".old"); err != nil {
		t.Fatalf("move original stage: %v", err)
	}
	if err := os.WriteFile(stage, []byte("other"), 0o600); err != nil {
		t.Fatalf("replace stage: %v", err)
	}
	var output bytes.Buffer
	if err := run(append([]string{"cleanup", "--target", stage}, identity...), nil, &output); err == nil {
		t.Fatal("expected identity mismatch")
	}
	if got := string(mustReadFile(t, stage)); got != "other" {
		t.Fatalf("replacement stage changed: %q", got)
	}
}

// TestHardLinkedStageIsRejected prevents an upload stage from aliasing another path.
// Requirements: file-writer/FR-003.
func TestHardLinkedStageIsRejected(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	stage := filepath.Join(directory, "target.upload")
	identity := identityArgs(mustRun(t, []string{"create", "--target", stage}, nil))
	if err := os.Link(stage, filepath.Join(directory, "alias")); err != nil {
		t.Fatalf("link stage: %v", err)
	}
	var output bytes.Buffer
	if err := run(append([]string{"stat", "--target", stage}, identity...), nil, &output); err == nil {
		t.Fatal("expected linked stage rejection")
	}
}

// TestCreateRejectsConflictAndSymlink verifies exclusive creation preserves existing entries.
// Requirements: file-writer/FR-001.
func TestCreateRejectsConflictAndSymlink(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	stage := filepath.Join(directory, "target.upload")
	mustRun(t, []string{"create", "--target", stage}, nil)
	for _, path := range []string{stage, filepath.Join(directory, "link.upload")} {
		if path != stage {
			if err := os.Symlink(stage, path); err != nil {
				t.Fatalf("create symlink: %v", err)
			}
		}
		var output bytes.Buffer
		if err := run([]string{"create", "--target", path}, nil, &output); err == nil {
			t.Fatalf("expected conflict for %s", path)
		}
	}
}

// TestPublishReplacesRegularDestination verifies the stage inode becomes the completed file.
// Requirements: file-writer/FR-006.
func TestPublishReplacesRegularDestination(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	stage := filepath.Join(directory, "target.upload")
	destination := filepath.Join(directory, "target")
	created := mustRun(t, []string{"create", "--target", stage}, nil)
	identity := identityArgs(created)
	mustRun(t, append([]string{"write", "--target", stage, "--offset", "0", "--length", "3"}, identity...), []byte("new"))
	if err := os.WriteFile(destination, []byte("old"), 0o600); err != nil {
		t.Fatalf("write old destination: %v", err)
	}
	old, err := os.Stat(destination)
	if err != nil {
		t.Fatalf("stat destination: %v", err)
	}
	oldIdentity := old.Sys().(*syscall.Stat_t)
	args := append([]string{"publish", "--target", stage, "--destination-device", fmt.Sprint(oldIdentity.Dev), "--destination-inode", fmt.Sprint(oldIdentity.Ino)}, identity...)
	mustRun(t, args, nil)
	if got := string(mustReadFile(t, destination)); got != "new" {
		t.Fatalf("published content = %q, want new", got)
	}
	published, err := os.Stat(destination)
	if err != nil {
		t.Fatalf("stat published destination: %v", err)
	}
	publishedIdentity := published.Sys().(*syscall.Stat_t)
	if publishedIdentity.Ino != created.Inode || publishedIdentity.Uid != created.UID {
		t.Fatalf("published identity = inode %d uid %d, want inode %d uid %d",
			publishedIdentity.Ino, publishedIdentity.Uid, created.Inode, created.UID)
	}
}

// TestWriteRejectsGap verifies an invalid offset cannot create a sparse stage.
// Requirements: file-writer/FR-002.
func TestWriteRejectsGap(t *testing.T) {
	t.Parallel()
	stage := filepath.Join(t.TempDir(), "target.upload")
	identity := identityArgs(mustRun(t, []string{"create", "--target", stage}, nil))
	var output bytes.Buffer
	err := run(append([]string{"write", "--target", stage, "--offset", "1", "--length", "1"}, identity...), bytes.NewReader([]byte("A")), &output)
	if err == nil {
		t.Fatal("expected offset gap rejection")
	}
	if got := mustReadFile(t, stage); len(got) != 0 {
		t.Fatalf("stage changed after rejected offset: %q", got)
	}
}

// TestPublishRejectsChangedDestination verifies publication leaves a changed destination untouched.
// Requirements: file-writer/FR-006.
func TestPublishRejectsChangedDestination(t *testing.T) {
	t.Parallel()
	destination, _, err := runPublishWithChangedDestination(t)
	if err == nil {
		t.Fatal("publish error = nil, want destination state mismatch")
	}
	if got := string(mustReadFile(t, destination)); got != "old" {
		t.Fatalf("destination changed: %q", got)
	}
}

// TestPublishFailureCleansStage verifies a failed publication removes the identified stage.
// Requirements: file-writer/FR-007.
func TestPublishFailureCleansStage(t *testing.T) {
	t.Parallel()
	_, stage, err := runPublishWithChangedDestination(t)
	var failure *operationError
	if !errors.As(err, &failure) || failure.Cleanup != "succeeded" {
		t.Fatalf("publish failure = %v, want successful cleanup", err)
	}
	if _, err := os.Lstat(stage); !os.IsNotExist(err) {
		t.Fatalf("stage remains after failed publish: %v", err)
	}
}

// runPublishWithChangedDestination creates an identified stage and a destination that
// appears after the stage was created, then returns the failed publication result.
func runPublishWithChangedDestination(t *testing.T) (string, string, error) {
	t.Helper()
	directory := t.TempDir()
	stage := filepath.Join(directory, "target.upload")
	destination := filepath.Join(directory, "target")
	identity := identityArgs(mustRun(t, []string{"create", "--target", stage}, nil))
	if err := os.WriteFile(destination, []byte("old"), 0o600); err != nil {
		t.Fatalf("write destination: %v", err)
	}
	args := append([]string{"publish", "--target", stage, "--destination-absent"}, identity...)
	var output bytes.Buffer
	err := run(args, nil, &output)
	return destination, stage, err
}

// TestPublishReportsCleanupFailure verifies a failed removal has its own response code.
// Requirements: file-writer/FR-007.
func TestPublishReportsCleanupFailure(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	stage := filepath.Join(directory, "target.upload")
	identity := identityArgs(mustRun(t, []string{"create", "--target", stage}, nil))
	if err := os.Chmod(directory, 0o500); err != nil {
		t.Fatalf("make directory read-only: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(directory, 0o700); err != nil {
			t.Errorf("restore directory permissions: %v", err)
		}
	})
	var output bytes.Buffer
	err := run(append([]string{"publish", "--target", stage, "--destination-absent"}, identity...), nil, &output)
	var failure *operationError
	if !errors.As(err, &failure) || failure.Cleanup != "failed" {
		t.Fatalf("publish failure = %v, want failed cleanup", err)
	}
}

// mustStageFixture creates an owned stage and returns its path and identity flags.
func mustStageFixture(t *testing.T) (string, []string) {
	t.Helper()
	stage := filepath.Join(t.TempDir(), "target.upload")
	return stage, identityArgs(mustRun(t, []string{"create", "--target", stage}, nil))
}

// TestPublishAppliesSelectedMode covers file-writer/FR-003 publication permissions.
func TestPublishAppliesSelectedMode(t *testing.T) {
	for _, test := range []struct {
		name string
		mode string
		want os.FileMode
	}{
		{name: "default shared", want: 0o664},
		{name: "shared", mode: "0664", want: 0o664},
		{name: "private", mode: "0660", want: 0o660},
	} {
		t.Run(test.name, func(t *testing.T) {
			stage, identity := mustStageFixture(t)
			args := []string{"publish", "--target", stage, "--destination-absent"}
			if test.mode != "" {
				args = append(args, "--mode", test.mode)
			}
			mustRun(t, append(args, identity...), nil)
			info, err := os.Stat(strings.TrimSuffix(stage, ".upload"))
			if err != nil {
				t.Fatal(err)
			}
			if got := info.Mode().Perm(); got != test.want {
				t.Fatalf("mode = %04o, want %04o", got, test.want)
			}
		})
	}
}

// identityArgs passes a created stage's filesystem identity into a later operation.
func identityArgs(created result) []string {
	return []string{
		"--expected-device", fmt.Sprint(created.Device),
		"--expected-inode", fmt.Sprint(created.Inode),
		"--expected-uid", fmt.Sprint(created.UID),
	}
}

// mustRun invokes a successful writer operation and decodes its JSON response.
func mustRun(t *testing.T, args []string, input []byte) result {
	t.Helper()
	var output bytes.Buffer
	if err := run(args, bytes.NewReader(input), &output); err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	var response result
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decode result %q: %v", output.String(), err)
	}
	return response
}

// mustReadFile reads a test file after an expected successful operation.
func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return content
}
