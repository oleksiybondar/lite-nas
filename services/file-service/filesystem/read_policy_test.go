package filesystem

import (
	"strings"
	"testing"
)

// TestReadScopeEnforcesStorageBoundary covers file-service-svc/FR-003 and FR-008.
func TestReadScopeEnforcesStorageBoundary(t *testing.T) {
	storage := t.TempDir()
	policy, err := NewPolicy([]string{storage}, nil)
	if err != nil {
		t.Fatal(err)
	}
	operator, err := policy.ReadScope([]string{"lite-nas-operator"})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/etc", "/var", "/sys", "/proc", "/dev", "/usr"} {
		if operator.AllowsDirectory(path) || operator.AllowsFile(path+"/passwd") {
			t.Errorf("operator can read system path %q", path)
		}
	}
	if !operator.AllowsDirectory("/") || !operator.AllowsFile(storage+"/test.txt") {
		t.Fatal("operator cannot navigate to configured storage")
	}
	admin, err := policy.ReadScope([]string{"sudo"})
	if err != nil {
		t.Fatal(err)
	}
	if !admin.AllowsDirectory("/etc") || !admin.AllowsFile("/etc/passwd") {
		t.Fatal("administrator was denied system reads")
	}
}

// TestMountedZFSRootsParsesOnlyZFSMounts covers file-service-svc/FR-008.
func TestMountedZFSRootsParsesOnlyZFSMounts(t *testing.T) {
	mountInfo := strings.NewReader("1 0 0:1 / /etc rw - ext4 dev rw\n" +
		"2 0 0:2 / /testpool rw - zfs testpool rw\n" +
		"3 0 0:3 / /tmppool rw - zfs tmppool rw\n")
	roots, err := mountedZFSRoots(mountInfo)
	if err != nil || len(roots) != 2 || roots[0] != "/testpool" || roots[1] != "/tmppool" {
		t.Fatalf("ZFS roots = %q, error = %v", roots, err)
	}
}
