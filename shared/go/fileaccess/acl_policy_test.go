package fileaccess

import "testing"

// TestReadACLPolicySelectsAppFolderAndHonorsGlobalEnforcement covers app scoped read decisions.
func TestReadACLPolicySelectsAppFolderAndHonorsGlobalEnforcement(t *testing.T) {
	policy := ReadACLPolicy{
		ApplicationFolders: map[string][]string{"moments": {"/home/alice/gallery"}},
		EnforcePaths:       []string{"/home/alice/gallery/private"},
	}
	if got := policy.ReadDecisionPath("moments", "/home/alice/gallery/photo.jpg"); got != "/home/alice/gallery" {
		t.Fatalf("whitelisted decision path = %q, want folder", got)
	}
	if got := policy.ReadDecisionPath("admin-panel", "/home/alice/gallery/photo.jpg"); got != "/home/alice/gallery/photo.jpg" {
		t.Fatalf("other app decision path = %q, want exact file", got)
	}
	if got := policy.ReadDecisionPath("moments", "/home/alice/gallery/private/photo.jpg"); got != "/home/alice/gallery/private/photo.jpg" {
		t.Fatalf("enforced decision path = %q, want exact file", got)
	}
}

// TestReadACLPolicyParentCoverageStopsAtEnforcedEntries covers optimized directory listing checks.
func TestReadACLPolicyParentCoverageStopsAtEnforcedEntries(t *testing.T) {
	policy := ReadACLPolicy{
		ApplicationFolders: map[string][]string{"moments": {"/home/alice/gallery"}},
		EnforcePaths:       []string{"/home/alice/gallery/private"},
	}
	if !policy.ParentFolderCoversEntry("moments", "/home/alice/gallery", "/home/alice/gallery/photo.jpg") {
		t.Fatal("whitelisted parent did not cover an ordinary child")
	}
	if policy.ParentFolderCoversEntry("moments", "/home/alice/gallery", "/home/alice/gallery/private/secret.jpg") {
		t.Fatal("whitelisted parent covered a globally enforced child")
	}
}
