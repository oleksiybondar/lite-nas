package config_test

import (
	"testing"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
)

// TestLoadACLConfigParsesSharedRules covers the shared ACL section consumed by file and stream services.
func TestLoadACLConfigParsesSharedRules(t *testing.T) {
	cfgFile, err := ini.Load([]byte("[acl]\nmoments_folders=/home/alice/gallery\nadmin_panel_folders=/srv/shared\nenforce_paths=/home/alice/gallery/private\n"))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := config.LoadACLConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadACLConfig() error = %v", err)
	}
	if got := policy.ReadDecisionPath("moments", "/home/alice/gallery/photo.jpg"); got != "/home/alice/gallery" {
		t.Fatalf("moments decision path = %q, want configured folder", got)
	}
	if got := policy.ReadDecisionPath("admin-panel", "/srv/shared/item.jpg"); got != "/srv/shared" {
		t.Fatalf("admin-panel decision path = %q, want configured folder", got)
	}
	if got := policy.ReadDecisionPath("moments", "/home/alice/gallery/private/secret.jpg"); got != "/home/alice/gallery/private/secret.jpg" {
		t.Fatalf("enforced decision path = %q, want exact path", got)
	}
}

// TestLoadACLConfigRejectsUncleanPaths covers malformed shared ACL configuration.
func TestLoadACLConfigRejectsUncleanPaths(t *testing.T) {
	for _, source := range []string{
		"[acl]\nmoments_folders=relative/path\n",
		"[acl]\nenforce_paths=/home/../private\n",
		"[acl]\nmoments_folders=/home/alice/gallery,\n",
	} {
		t.Run(source, func(t *testing.T) {
			cfgFile, err := ini.Load([]byte(source))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := config.LoadACLConfig(cfgFile); err == nil {
				t.Fatal("LoadACLConfig() accepted an unclean ACL path")
			}
		})
	}
}
