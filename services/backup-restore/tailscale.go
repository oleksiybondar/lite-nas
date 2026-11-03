package backuprestore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const tailscaleArchivePath = "/tailscale.json"

// TailscaleState stores the local preferences reported as replayable set flags.
type TailscaleState struct {
	SchemaVersion int      `json:"schemaVersion"`
	Flags         []string `json:"flags"`
}

// TailscalePlugin runs only the documented preference export and apply commands.
type TailscalePlugin struct {
	run func(...string) ([]byte, error)
}

// NewTailscalePlugin binds the installed CLI to a bounded, shell-free runner.
func NewTailscalePlugin() *TailscalePlugin {
	return &TailscalePlugin{run: runTailscaleCommand}
}

// runTailscaleCommand executes the installed Tailscale CLI without a shell.
func runTailscaleCommand(arguments ...string) ([]byte, error) {
	if len(arguments) == 0 {
		return nil, fmt.Errorf("missing Tailscale command")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "/usr/bin/tailscale", arguments...)
	command.Env = []string{"PATH=/usr/bin:/bin", "LC_ALL=C"}
	var output limitedUFWOutput
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("tailscale %s failed: %w: %s", arguments[0], err, strings.TrimSpace(output.String()))
	}
	if output.exceeded {
		return nil, fmt.Errorf("tailscale output exceeds limit")
	}
	return output.Bytes(), nil
}

// Export captures preferences that the installed CLI says can be replayed by set.
func (plugin *TailscalePlugin) Export() ([]byte, error) {
	if plugin == nil || plugin.run == nil {
		return nil, fmt.Errorf("Tailscale plugin unavailable")
	}
	output, err := plugin.run("get", "--set-flags")
	if err != nil {
		return nil, err
	}
	state := TailscaleState{SchemaVersion: 1, Flags: strings.Fields(string(output))}
	if err := validateTailscaleState(state); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// Apply replays a validated snapshot through tailscale set.
func (plugin *TailscalePlugin) Apply(data []byte) error {
	if plugin == nil || plugin.run == nil {
		return fmt.Errorf("Tailscale plugin unavailable")
	}
	state, err := decodeTailscaleState(data)
	if err != nil {
		return err
	}
	currentOutput, err := plugin.run("get", "--set-flags")
	if err != nil {
		return err
	}
	current := TailscaleState{SchemaVersion: 1, Flags: strings.Fields(string(currentOutput))}
	if err := validateTailscaleState(current); err != nil {
		return err
	}
	supported := make(map[string]bool, len(current.Flags))
	for _, flag := range current.Flags {
		name, _, _ := strings.Cut(flag[2:], "=")
		supported[name] = true
	}
	for _, flag := range state.Flags {
		name, _, _ := strings.Cut(flag[2:], "=")
		if !supported[name] {
			return fmt.Errorf("Tailscale preference %q is unsupported by the installed CLI", name)
		}
	}
	_, err = plugin.run(append([]string{"set"}, state.Flags...)...)
	return err
}

// decodeTailscaleState rejects unknown JSON fields and non-replayable flags.
func decodeTailscaleState(data []byte) (TailscaleState, error) {
	var state TailscaleState
	if len(data) > 1<<20 {
		return state, fmt.Errorf("Tailscale JSON exceeds limit")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return state, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return state, fmt.Errorf("trailing Tailscale JSON")
	}
	return state, validateTailscaleState(state)
}

// validateTailscaleState bounds and parses the CLI's single-operand flag format.
func validateTailscaleState(state TailscaleState) error {
	if state.SchemaVersion != 1 || state.Flags == nil || len(state.Flags) > 64 {
		return fmt.Errorf("invalid Tailscale preference document")
	}
	seen := make(map[string]bool)
	for _, flag := range state.Flags {
		if !strings.HasPrefix(flag, "--") || len(flag) > 4096 {
			return fmt.Errorf("invalid Tailscale preference flag")
		}
		name, value, hasValue := strings.Cut(flag[2:], "=")
		if !validTailscaleFlagName(name) || name == "accept-risk" || name == "authkey" || seen[name] || strings.ContainsAny(value, "\x00\r\n\t ' \"\\") {
			return fmt.Errorf("unsupported Tailscale preference flag %q", name)
		}
		if hasValue && strings.HasPrefix(value, "--") {
			return fmt.Errorf("option-shaped Tailscale preference value")
		}
		if err := validateTailscaleFlagValue(name, value, hasValue); err != nil {
			return err
		}
		seen[name] = true
	}
	return nil
}

// validTailscaleFlagName rejects options and operands outside set's flag grammar.
func validTailscaleFlagName(name string) bool {
	if name == "" || len(name) > 64 || name[0] < 'a' || name[0] > 'z' {
		return false
	}
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}
	return true
}

// validateTailscaleFlagValue applies the operand rules of documented set flags.
func validateTailscaleFlagValue(name, value string, hasValue bool) error {
	stringsFlags := map[string]bool{"advertise-routes": true, "exit-node": true, "hostname": true, "netfilter-mode": true, "nickname": true, "operator": true, "relay-server-port": true, "relay-server-static-endpoints": true}
	boolFlags := map[string]bool{"accept-dns": true, "accept-routes": true, "advertise-connector": true, "advertise-exit-node": true, "auto-update": true, "exit-node-allow-lan-access": true, "report-posture": true, "shields-up": true, "snat-subnet-routes": true, "ssh": true, "stateful-filtering": true, "update-check": true, "webclient": true}
	if boolFlags[name] {
		if !hasValue || value == "true" || value == "false" {
			return nil
		}
		return fmt.Errorf("invalid boolean Tailscale preference %s", name)
	}
	if !stringsFlags[name] {
		// Newly added vendor flags are permitted only when the installed CLI
		// itself emits that flag in get --set-flags, checked before replay.
		return nil
	}
	if !hasValue {
		return fmt.Errorf("Tailscale preference %s requires a value", name)
	}
	switch name {
	case "advertise-routes":
		if value == "" {
			return nil
		}
		for _, route := range strings.Split(value, ",") {
			if _, err := netip.ParsePrefix(route); err != nil {
				return fmt.Errorf("invalid advertised Tailscale route")
			}
		}
	case "netfilter-mode":
		if value != "on" && value != "nodivert" && value != "off" {
			return fmt.Errorf("invalid Tailscale netfilter mode")
		}
	case "relay-server-port":
		if value == "" {
			return nil
		}
		port, err := strconv.Atoi(value)
		if err != nil || port < 0 || port > 65535 {
			return fmt.Errorf("invalid Tailscale relay server port")
		}
	case "relay-server-static-endpoints":
		if value == "" {
			return nil
		}
		for _, endpoint := range strings.Split(value, ",") {
			parsed, err := netip.ParseAddrPort(endpoint)
			if err != nil || !parsed.IsValid() || parsed.Port() == 0 {
				return fmt.Errorf("invalid Tailscale relay server endpoint")
			}
		}
	default:
		for _, char := range value {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || strings.ContainsRune("-._:@", char)) {
				return fmt.Errorf("invalid Tailscale preference %s", name)
			}
		}
	}
	return nil
}

// tailscaleInstalled reports whether the fixed CLI binary is available.
func tailscaleInstalled() bool {
	info, err := os.Stat("/usr/bin/tailscale")
	return err == nil && info.Mode().IsRegular() && info.Mode()&0o111 != 0
}
