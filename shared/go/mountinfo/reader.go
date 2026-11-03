// Package mountinfo parses Linux mountinfo for both mount notifications and
// authoritative mount checks before antivirus scans.
package mountinfo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const maxMountInfoLine = 16 * 1024

// Record is one decoded Linux mountinfo entry from the current mount
// namespace. The kernel ID identifies one mount only within the current boot.
type Record struct {
	ID     uint64
	Point  string
	Source string
	FSType string
}

// Parse rejects malformed or oversized records instead of emitting
// partial snapshots that would look like mass unmounts.
func Parse(reader io.Reader) (map[uint64]Record, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), maxMountInfoLine)
	mounts := make(map[uint64]Record)
	for scanner.Scan() {
		mount, err := parseMountInfoLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		if _, exists := mounts[mount.ID]; exists {
			return nil, fmt.Errorf("duplicate mount ID %d", mount.ID)
		}
		mounts[mount.ID] = mount
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read mountinfo: %w", err)
	}
	return mounts, nil
}

// parseMountInfoLine reads the documented fields around the optional-field
// separator and decodes the kernel's octal path escapes.
func parseMountInfoLine(line string) (Record, error) {
	fields := strings.Fields(line)
	separator := -1
	for index, field := range fields {
		if field == "-" {
			separator = index
			break
		}
	}
	if separator < 6 || len(fields) < separator+3 {
		return Record{}, fmt.Errorf("invalid mountinfo record")
	}
	id, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil || id == 0 {
		return Record{}, fmt.Errorf("invalid mount ID")
	}
	point, err := decodeMountField(fields[4])
	if err != nil || !strings.HasPrefix(point, "/") {
		return Record{}, fmt.Errorf("invalid mount point")
	}
	source, err := decodeMountField(fields[separator+2])
	if err != nil {
		return Record{}, fmt.Errorf("invalid mount source: %w", err)
	}
	fstype := fields[separator+1]
	if len(point) > 4096 || len(source) > 4096 || len(fstype) > 128 || fstype == "" {
		return Record{}, fmt.Errorf("mount metadata exceeds contract limits")
	}
	return Record{ID: id, Point: point, Source: source, FSType: fstype}, nil
}

// decodeMountField accepts only the four escape sequences Linux mountinfo
// uses for spaces, tabs, newlines and backslashes.
func decodeMountField(field string) (string, error) {
	var out strings.Builder
	for index := 0; index < len(field); index++ {
		if field[index] != '\\' {
			out.WriteByte(field[index])
			continue
		}
		if index+3 >= len(field) {
			return "", fmt.Errorf("truncated mount escape")
		}
		switch field[index : index+4] {
		case `\040`:
			out.WriteByte(' ')
		case `\011`:
			out.WriteByte('\t')
		case `\012`:
			out.WriteByte('\n')
		case `\134`:
			out.WriteByte('\\')
		default:
			return "", fmt.Errorf("unsupported mount escape")
		}
		index += 3
	}
	return out.String(), nil
}

// ReadCurrent reads the caller's mount namespace. Consumers must re-read
// immediately before sensitive actions because mounts can change afterward.
func ReadCurrent() (map[uint64]Record, error) {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return Parse(file)
}
