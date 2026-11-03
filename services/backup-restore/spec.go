package backuprestore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"

	"lite-nas/shared/systemd"
)

// Selection identifies a host path and whether the complete directory tree is owned.
type Selection struct {
	Path      string
	Directory bool
	Optional  bool
}

// Spec is one validated compiled-plugin configuration for an application.
type Spec struct {
	Application  string
	SourcePath   string
	Plugin       string
	ConfigPath   string
	Selections   []Selection
	RestartUnits []string
}

// LoadSpecs reads application drop-ins once at startup. RPC callers cannot add paths.
func LoadSpecs(directory string) (map[string]Spec, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	specs := make(map[string]Spec)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".conf" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".conf")
		if !validApplication(name) {
			return nil, fmt.Errorf("invalid application ID %q", name)
		}
		path := filepath.Join(directory, entry.Name())
		spec, err := LoadSpec(path, name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		specs[name] = spec
	}
	return specs, nil
}

// LoadSpec validates one compiled plugin and its declared host path selections.
func LoadSpec(path, application string) (Spec, error) {
	if !validApplication(application) {
		return Spec{}, fmt.Errorf("invalid application ID %q", application)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}
	configuration, err := ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: false}, data)
	if err != nil {
		return Spec{}, err
	}
	backup, err := configuration.GetSection("backup")
	if err != nil || !backup.HasKey("plugin") {
		return Spec{}, fmt.Errorf("unsupported backup plugin or settings")
	}
	for _, key := range backup.Keys() {
		if key.Name() != "plugin" && key.Name() != "restart_units" && key.Name() != "config_path" {
			return Spec{}, fmt.Errorf("unsupported backup plugin or settings")
		}
	}
	plugin := backup.Key("plugin").String()
	if plugin != "file" && plugin != "ufw" && plugin != "cloudflared" && plugin != "tailscale" {
		return Spec{}, fmt.Errorf("unsupported backup plugin or settings")
	}
	spec := Spec{Application: application, SourcePath: path, Plugin: plugin}
	if backup.HasKey("config_path") {
		if plugin != "cloudflared" {
			return Spec{}, fmt.Errorf("config_path is only supported by cloudflared")
		}
		spec.ConfigPath = backup.Key("config_path").String()
	}
	if backup.HasKey("restart_units") {
		seenUnits := make(map[string]bool)
		for _, raw := range strings.Split(backup.Key("restart_units").String(), ",") {
			unit := strings.TrimSpace(raw)
			if unit == "" && strings.TrimSpace(backup.Key("restart_units").String()) == "" {
				break
			}
			if !systemd.ValidUnitName(unit) || seenUnits[unit] {
				return Spec{}, fmt.Errorf("invalid or repeated restart unit %q", unit)
			}
			seenUnits[unit] = true
			spec.RestartUnits = append(spec.RestartUnits, unit)
		}
	}
	if plugin == "ufw" {
		if application != "ufw-rules" || len(configuration.Sections()) != 2 || len(spec.RestartUnits) != 0 {
			return Spec{}, fmt.Errorf("UFW plugin accepts no selections or restart units")
		}
		spec.Selections = []Selection{{Path: ufwArchivePath}}
		return spec, nil
	}
	if plugin == "tailscale" {
		if application != "tailscale" || len(configuration.Sections()) != 2 || len(spec.RestartUnits) != 0 {
			return Spec{}, fmt.Errorf("Tailscale plugin accepts no selections or restart units")
		}
		spec.Selections = []Selection{{Path: tailscaleArchivePath}}
		return spec, nil
	}
	seen := make(map[string]bool)
	for _, section := range configuration.Sections() {
		if section.Name() == ini.DefaultSection || section.Name() == "backup" {
			continue
		}
		name := section.Name()
		isFile := strings.HasPrefix(name, "file.") && len(name) > len("file.")
		isDirectory := strings.HasPrefix(name, "directory.") && len(name) > len("directory.")
		if (!isFile && !isDirectory) || !section.HasKey("path") || len(section.Keys()) > 2 {
			return Spec{}, fmt.Errorf("invalid selection section %q", name)
		}
		for _, key := range section.Keys() {
			if key.Name() != "path" && key.Name() != "optional" {
				return Spec{}, fmt.Errorf("invalid selection section %q", name)
			}
		}
		optional := false
		if section.HasKey("optional") {
			switch section.Key("optional").String() {
			case "true":
				optional = true
			case "false":
			default:
				return Spec{}, fmt.Errorf("invalid optional value in %q", name)
			}
		}
		selected := section.Key("path").String()
		if !filepath.IsAbs(selected) || selected == "/" || filepath.Clean(selected) != selected || seen[selected] {
			return Spec{}, fmt.Errorf("invalid or repeated path %q", selected)
		}
		seen[selected] = true
		spec.Selections = append(spec.Selections, Selection{Path: selected, Directory: isDirectory, Optional: optional})
	}
	if len(spec.Selections) == 0 {
		return Spec{}, fmt.Errorf("spec has no selected paths")
	}
	if plugin == "cloudflared" && (application != "cloudflared" || !filepath.IsAbs(spec.ConfigPath) || filepath.Clean(spec.ConfigPath) != spec.ConfigPath || !cloudflaredSelected(spec, spec.ConfigPath)) {
		return Spec{}, fmt.Errorf("cloudflared plugin requires a selected absolute config path")
	}
	for i, left := range spec.Selections {
		for _, right := range spec.Selections[i+1:] {
			if left.Directory && strings.HasPrefix(right.Path, left.Path+"/") || right.Directory && strings.HasPrefix(left.Path, right.Path+"/") {
				return Spec{}, fmt.Errorf("overlapping selections %q and %q", left.Path, right.Path)
			}
		}
	}
	return spec, nil
}

// validApplication accepts a single safe filename and NATS argument component.
func validApplication(name string) bool {
	if name == "" || len(name) > 64 || name[0] == '-' {
		return false
	}
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}
	return true
}
