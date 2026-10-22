package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestPolicyRejectsProtectedDescendants covers file-service-svc/FR-004 and FR-009.
func TestPolicyRejectsProtectedDescendants(t *testing.T) {
	root := t.TempDir()
	protected := filepath.Join(root, "private")
	if err := os.Mkdir(protected, 0o700); err != nil {
		t.Fatal(err)
	}
	policy, err := NewPolicy([]string{root}, []string{protected})
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{protected, filepath.Join(protected, "child")} {
		if err := policy.AllowMutation(target); !errors.Is(err, ErrProtectedPath) {
			t.Errorf("AllowMutation(%q) = %v, want protected", target, err)
		}
	}
	if err := policy.AllowMutation(filepath.Join(root, "private-backup", "file")); err != nil {
		t.Fatalf("sibling prefix rejected: %v", err)
	}
}

// TestPolicyRejectsOutsideWriteRoots covers file-service-svc/FR-004 and FR-009.
func TestPolicyRejectsOutsideWriteRoots(t *testing.T) {
	root := t.TempDir()
	policy, err := NewPolicy([]string{root}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := policy.AllowMutation(filepath.Join(filepath.Dir(root), "other")); !errors.Is(err, ErrOutsideWriteRoots) {
		t.Fatalf("outside target error = %v", err)
	}
	if err := policy.AllowMutation(root + "/../other"); err == nil {
		t.Fatal("unclean target was accepted")
	}
}
