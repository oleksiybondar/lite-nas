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

// TestPolicyRequiresConfiguredWriteRoots prevents a hidden fallback root.
func TestPolicyRequiresConfiguredWriteRoots(t *testing.T) {
	if _, err := NewPolicy(nil, nil); err == nil {
		t.Fatal("NewPolicy accepted missing write roots")
	}
}

// TestPolicyUsesOnlyConfiguredProtection covers removal of code-owned roots.
func TestPolicyUsesOnlyConfiguredProtection(t *testing.T) {
	policy, err := NewPolicy([]string{"/"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := policy.AllowMutation("/etc/lite-nas-test"); err != nil {
		t.Fatalf("unconfigured root was denied: %v", err)
	}
}

// TestPolicyProtectsFutureConfiguredRoot covers absent protected directories.
func TestPolicyProtectsFutureConfiguredRoot(t *testing.T) {
	root := t.TempDir()
	protected := filepath.Join(root, "future")
	policy, err := NewPolicy([]string{root}, []string{protected})
	if err != nil {
		t.Fatal(err)
	}
	if err := policy.AllowMutation(filepath.Join(protected, "file")); !errors.Is(err, ErrProtectedPath) {
		t.Fatalf("future protected root error = %v", err)
	}
}

// TestDirectoryModeUsesConfiguredMask covers file-service-svc/FR-010 presets.
func TestDirectoryModeUsesConfiguredMask(t *testing.T) {
	root := t.TempDir()
	policy, err := NewPolicy([]string{root}, nil, ReadRules{KnownPathDirectories: []string{root + "/*/.lite-nas"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		path    string
		private bool
		want    string
	}{
		{name: "known path", path: root + "/alice/.lite-nas", want: "0771"},
		{name: "other path", path: root + "/alice/photos", want: "0775"},
		{name: "private wins", path: root + "/alice/.lite-nas", private: true, want: "0770"},
		{name: "mask one segment", path: root + "/alice/sub/.lite-nas", want: "0775"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := policy.DirectoryMode(test.path, test.private); got != test.want {
				t.Fatalf("mode = %s, want %s", got, test.want)
			}
		})
	}
}

// TestReadRulesSeparateCallerClasses covers file-service-svc/FR-010 read policy.
func TestReadRulesSeparateCallerClasses(t *testing.T) {
	root := t.TempDir()
	policy, err := NewPolicy([]string{root}, nil, ReadRules{
		NonAdminDenied: []string{root + "/restricted"}, AllDenied: []string{root + "/sealed"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		groups  []string
		path    string
		allowed bool
	}{
		{name: "user restricted", groups: []string{"lite-nas-operator"}, path: root + "/restricted/file", allowed: false},
		{name: "admin restricted", groups: []string{"sudo"}, path: root + "/restricted/file", allowed: true},
		{name: "admin sealed", groups: []string{"sudo"}, path: root + "/sealed/file", allowed: false},
		{name: "user sibling", groups: []string{"lite-nas-operator"}, path: root + "/restricted-other/file", allowed: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			scope, err := policy.ReadScope(test.groups)
			if err != nil {
				t.Fatal(err)
			}
			if got := scope.AllowsFile(test.path); got != test.allowed {
				t.Fatalf("allowed = %v, want %v", got, test.allowed)
			}
		})
	}
}

// TestPathMaskValidationRejectsMalformedMasks covers config input boundaries.
func TestPathMaskValidationRejectsMalformedMasks(t *testing.T) {
	for _, mask := range []string{"relative/*", "/home/a*/.lite-nas", "/home/**/.lite-nas", "/home/?/.lite-nas", "/home/../x"} {
		if err := ValidatePathMask(mask); err == nil {
			t.Fatalf("accepted mask %q", mask)
		}
	}
}

// TestDeniedPathsCoverMutations covers file-service-svc/FR-009 caller restrictions.
func TestDeniedPathsCoverMutations(t *testing.T) {
	root := t.TempDir()
	policy, err := NewPolicy([]string{root}, nil, ReadRules{
		NonAdminDenied: []string{root + "/*/private"}, AllDenied: []string{root + "/sealed"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name       string
		path       string
		groups     []string
		wantDenied bool
	}{
		{name: "non admin masked", path: root + "/alice/private/file", groups: []string{"lite-nas-operator"}, wantDenied: true},
		{name: "admin permitted", path: root + "/alice/private/file", groups: []string{"sudo"}},
		{name: "admin sealed", path: root + "/sealed/file", groups: []string{"sudo"}, wantDenied: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			denied := errors.Is(policy.AllowSubject(test.path, test.groups), ErrDeniedPath)
			if denied != test.wantDenied {
				t.Fatalf("denied = %v, want %v", denied, test.wantDenied)
			}
		})
	}
}
