package fileaccess

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// MountedZFSRoots returns canonical mountpoints from the current Linux mount table.
// A malformed mount table fails closed rather than expanding the read scope.
func MountedZFSRoots() ([]string, error) {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ParseZFSRoots(file)
}

// ParseZFSRoots extracts ZFS mountpoints without invoking a host command.
func ParseZFSRoots(mountInfo io.Reader) ([]string, error) {
	var roots []string
	scanner := bufio.NewScanner(mountInfo)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " - ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed mountinfo record")
		}
		filesystem := strings.Fields(parts[1])
		if len(filesystem) == 0 {
			return nil, fmt.Errorf("malformed mountinfo filesystem")
		}
		if filesystem[0] != "zfs" {
			continue
		}
		mount := strings.Fields(parts[0])
		if len(mount) < 5 {
			return nil, fmt.Errorf("malformed ZFS mountinfo record")
		}
		path, err := strconv.Unquote("\"" + mount[4] + "\"")
		if err != nil || ValidateHostPath(path) != nil || path == "/" {
			return nil, fmt.Errorf("invalid ZFS mountpoint")
		}
		roots = append(roots, path)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan mountinfo: %w", err)
	}
	return roots, nil
}
