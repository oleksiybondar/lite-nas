package workers

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// mountEntry stores one parsed mount entry from mountinfo or proc mounts.
type mountEntry struct {
	Major      uint64
	Minor      uint64
	MountPoint string
	Source     string
	Filesystem string
	ReadOnly   bool
}

// mountUsage stores one mount usage calculation derived from statfs.
type mountUsage struct {
	TotalBytes     uint64
	UsedBytes      uint64
	FreeBytes      uint64
	AvailableBytes uint64
	UsedPercent    float64
}

// statFSFunc abstracts mount usage collection for deterministic tests.
type statFSFunc func(path string) (mountUsage, error)

// collectMountEntries reads mountinfo and falls back to /proc/mounts when
// mountinfo cannot be read.
func collectMountEntries(mountInfoPath string, procMountsPath string) ([]mountEntry, error) {
	entries, err := parseMountInfoFile(mountInfoPath)
	if err == nil {
		return entries, nil
	}

	return parseProcMountsFile(procMountsPath)
}

// parseMountInfoFile reads and parses Linux /proc/self/mountinfo.
func parseMountInfoFile(path string) ([]mountEntry, error) {
	content, err := readTrustedFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return parseMountInfo(string(content))
}

// parseMountInfo parses Linux mountinfo content.
func parseMountInfo(content string) ([]mountEntry, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	result := make([]mountEntry, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		entry, err := parseMountInfoLine(line)
		if err != nil {
			return nil, err
		}

		result = append(result, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan mountinfo: %w", err)
	}

	return result, nil
}

// parseMountInfoLine parses one mountinfo line.
func parseMountInfoLine(line string) (mountEntry, error) {
	sections := strings.SplitN(line, " - ", 2)
	if len(sections) != 2 {
		return mountEntry{}, fmt.Errorf("mountinfo line missing separator: %q", line)
	}

	left := strings.Fields(sections[0])
	right := strings.Fields(sections[1])
	if len(left) < 6 || len(right) < 3 {
		return mountEntry{}, fmt.Errorf("mountinfo line malformed: %q", line)
	}

	major, minor, err := parseDevNumbers(left[2])
	if err != nil {
		return mountEntry{}, err
	}

	return mountEntry{
		Major:      major,
		Minor:      minor,
		MountPoint: decodeMountField(left[4]),
		Source:     decodeMountField(right[1]),
		Filesystem: right[0],
		ReadOnly:   mountOptionsContain(left[5], right[2], "ro"),
	}, nil
}

// parseProcMountsFile reads and parses Linux /proc/mounts.
func parseProcMountsFile(path string) ([]mountEntry, error) {
	content, err := readTrustedFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return parseProcMounts(string(content))
}

// parseProcMounts parses Linux /proc/mounts content.
func parseProcMounts(content string) ([]mountEntry, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	result := make([]mountEntry, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			return nil, fmt.Errorf("proc mounts line malformed: %q", line)
		}

		result = append(result, mountEntry{
			MountPoint: decodeMountField(fields[1]),
			Source:     decodeMountField(fields[0]),
			Filesystem: fields[2],
			ReadOnly:   mountOptionsContain(fields[3], "", "ro"),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan proc mounts: %w", err)
	}

	return result, nil
}

// readMountUsage reads filesystem usage using statfs.
func readMountUsage(path string) (mountUsage, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return mountUsage{}, err
	}

	if stat.Bsize <= 0 {
		return mountUsage{}, fmt.Errorf("statfs %s returned invalid block size %d", path, stat.Bsize)
	}

	blockSize := uint64(stat.Bsize)
	total := stat.Blocks * blockSize
	free := stat.Bfree * blockSize
	available := stat.Bavail * blockSize
	used := total - free

	usedPercent := 0.0
	if total > 0 {
		usedPercent = (float64(used) / float64(total)) * 100
	}

	return mountUsage{
		TotalBytes:     total,
		UsedBytes:      used,
		FreeBytes:      free,
		AvailableBytes: available,
		UsedPercent:    usedPercent,
	}, nil
}

// resolveSourceDeviceNode resolves one mount source to a local kernel node when
// the source points directly at or resolves to /dev/<node>.
func resolveSourceDeviceNode(source string) *string {
	if !strings.HasPrefix(source, "/dev/") {
		return nil
	}

	resolved, err := filepath.EvalSymlinks(source)
	if err == nil && strings.HasPrefix(resolved, "/dev/") {
		node := filepath.Base(resolved)
		return &node
	}

	trimmed := strings.TrimPrefix(source, "/dev/")
	if !strings.Contains(trimmed, "/") {
		node := filepath.Base(source)
		return &node
	}

	return nil
}

// parseDevNumbers parses one "<major>:<minor>" string.
func parseDevNumbers(value string) (uint64, uint64, error) {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("parse dev numbers %q: missing separator", value)
	}

	major, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse major %q: %w", parts[0], err)
	}

	minor, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse minor %q: %w", parts[1], err)
	}

	return major, minor, nil
}

// mountOptionsContain reports whether either option list contains one key.
func mountOptionsContain(primary string, secondary string, key string) bool {
	return commaListContains(primary, key) || commaListContains(secondary, key)
}

// commaListContains reports whether one comma-separated option string contains
// the provided key.
func commaListContains(options string, key string) bool {
	for _, option := range strings.Split(options, ",") {
		if option == key {
			return true
		}
	}

	return false
}

// decodeMountField decodes octal-escaped mountinfo and proc-mounts fields.
func decodeMountField(value string) string {
	replacer := strings.NewReplacer(
		`\\`, `\`,
		`\040`, " ",
		`\011`, "\t",
		`\012`, "\n",
		`\134`, `\`,
	)

	return replacer.Replace(value)
}
