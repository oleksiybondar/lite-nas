package filesystem

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"lite-nas/shared/fileaccess"
	"lite-nas/shared/roleauth"
)

// systemReadRoots cannot become non-admin storage merely through a mount or write-root setting.
var systemReadRoots = []string{
	"/bin", "/boot", "/dev", "/etc", "/lib", "/lib64", "/proc",
	"/root", "/run", "/sbin", "/snap", "/sys", "/usr", "/var",
}

// ReadScope is one request's view of directories visible through File Service.
type ReadScope struct {
	administrator bool
	roots         []string
}

// ReadScope resolves a subject's role and mounted ZFS storage for one request.
func (p Policy) ReadScope(groups []string) (ReadScope, error) {
	if roleauth.MatchesRequirement(groups, roleauth.RequirementAdministrator) {
		return ReadScope{administrator: true}, nil
	}
	mountInfo, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ReadScope{}, fmt.Errorf("read mount information: %w", err)
	}
	defer mountInfo.Close()
	zfsRoots, err := mountedZFSRoots(mountInfo)
	if err != nil {
		return ReadScope{}, err
	}
	roots := make([]string, 0, 1+len(p.writeRoots)+len(zfsRoots))
	for _, root := range append(append([]string{"/home"}, p.writeRoots...), zfsRoots...) {
		if root == "/" || fileaccess.WithinRoots(root, systemReadRoots) {
			continue
		}
		roots = append(roots, root)
	}
	return ReadScope{roots: roots}, nil
}

// AllowsFile permits regular-file reads only within an approved storage tree.
func (scope ReadScope) AllowsFile(path string) bool {
	return scope.administrator || fileaccess.WithinRoots(path, scope.roots)
}

// AllowsDirectory also permits ancestors needed to navigate to approved storage.
func (scope ReadScope) AllowsDirectory(path string) bool {
	if scope.AllowsFile(path) {
		return true
	}
	for _, root := range scope.roots {
		if fileaccess.WithinRoots(root, []string{path}) {
			return true
		}
	}
	return false
}

// mountedZFSRoots reads mounted ZFS paths from Linux mountinfo without executing zfs.
func mountedZFSRoots(mountInfo io.Reader) ([]string, error) {
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
		if err != nil || fileaccess.ValidateHostPath(path) != nil {
			return nil, fmt.Errorf("invalid ZFS mountpoint")
		}
		roots = append(roots, path)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan mountinfo: %w", err)
	}
	return roots, nil
}
