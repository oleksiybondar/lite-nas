package backuprestore

import (
	"archive/tar"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// CloudflaredPlugin validates a locally managed tunnel through cloudflared.
type CloudflaredPlugin struct {
	run func(string) error
}

// NewCloudflaredPlugin binds the installed CLI to ingress validation.
func NewCloudflaredPlugin() *CloudflaredPlugin {
	return &CloudflaredPlugin{run: runCloudflaredValidation}
}

// runCloudflaredValidation asks cloudflared to validate one explicit config path.
func runCloudflaredValidation(path string) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || strings.HasPrefix(path, "-") {
		return fmt.Errorf("invalid cloudflared config path")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "/usr/bin/cloudflared", "tunnel", "--config", path, "ingress", "validate")
	command.Env = []string{"PATH=/usr/bin:/bin", "LC_ALL=C"}
	var output limitedUFWOutput
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return fmt.Errorf("cloudflared ingress validation failed: %w: %s", err, strings.TrimSpace(output.String()))
	}
	if output.exceeded {
		return fmt.Errorf("cloudflared output exceeds limit")
	}
	return nil
}

// Validate checks the local tunnel config and its archived credentials before capture or restore.
func (plugin *CloudflaredPlugin) Validate(config []byte, archivedFiles map[string]bool, configPath string) error {
	if plugin == nil || plugin.run == nil || len(config) == 0 || len(config) > 1<<20 {
		return fmt.Errorf("cloudflared plugin or config unavailable")
	}
	credentialPath, err := cloudflaredCredentialPath(config)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(credentialPath) || filepath.Clean(credentialPath) != credentialPath || !archivedFiles[credentialPath] {
		return fmt.Errorf("cloudflared credential file must be selected for backup")
	}
	return plugin.run(configPath)
}

// cloudflaredCredentialPath reads the two required local tunnel fields.
func cloudflaredCredentialPath(config []byte) (string, error) {
	var document struct {
		Tunnel          string `yaml:"tunnel"`
		CredentialsFile string `yaml:"credentials-file"`
	}
	if err := yaml.Unmarshal(config, &document); err != nil {
		return "", err
	}
	if document.Tunnel == "" || document.CredentialsFile == "" {
		return "", fmt.Errorf("cloudflared config must name a tunnel and credential file")
	}
	return document.CredentialsFile, nil
}

// validateLiveCloudflared validates a configured local tunnel before archiving it.
func validateLiveCloudflared(spec Spec, plugin *CloudflaredPlugin) error {
	configPath := spec.ConfigPath
	config, err := readBoundedUFWFile(configPath)
	if err != nil {
		return err
	}
	credentialPath, err := cloudflaredCredentialPath(config)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(credentialPath) || filepath.Clean(credentialPath) != credentialPath || !cloudflaredSelected(spec, credentialPath) {
		return fmt.Errorf("cloudflared credential file must be selected for backup")
	}
	info, err := os.Lstat(credentialPath)
	if err != nil {
		return err
	}
	return plugin.Validate(config, map[string]bool{credentialPath: info.Mode().IsRegular()}, configPath)
}

// cloudflaredSelected checks only content selections, excluding the archived spec.
func cloudflaredSelected(spec Spec, path string) bool {
	for _, selection := range spec.Selections {
		if path == selection.Path || selection.Directory && strings.HasPrefix(path, selection.Path+"/") {
			return true
		}
	}
	return false
}

// validateStagedCloudflared checks a complete archive before changing host files.
func validateStagedCloudflared(spec Spec, entries []stagedEntry, plugin *CloudflaredPlugin) error {
	configPath := spec.ConfigPath
	archivedFiles := make(map[string]bool)
	var stagedConfig string
	for _, entry := range entries {
		if entry.header.Typeflag == tar.TypeReg {
			archivedFiles[entry.path] = true
		}
		if entry.path == configPath {
			stagedConfig = entry.stage
		}
	}
	if stagedConfig == "" {
		return fmt.Errorf("cloudflared config missing from archive")
	}
	config, err := readBoundedUFWFile(stagedConfig)
	if err != nil {
		return err
	}
	return plugin.Validate(config, archivedFiles, stagedConfig)
}

// cloudflaredInstalled reports whether the fixed CLI binary is available.
func cloudflaredInstalled() bool {
	info, err := os.Stat("/usr/bin/cloudflared")
	return err == nil && info.Mode().IsRegular() && info.Mode()&0o111 != 0
}
