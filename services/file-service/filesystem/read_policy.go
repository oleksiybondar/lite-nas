package filesystem

import (
	"fmt"
	"io"
	"os"

	"lite-nas/shared/fileaccess"
	"lite-nas/shared/roleauth"
)

// ReadScope is one request's view of directories visible through File Service.
type ReadScope struct {
	administrator bool
	roots         []string
	denied        []string
}

// ReadScope resolves a subject's role and mounted ZFS storage for one request.
func (p Policy) ReadScope(groups []string) (ReadScope, error) {
	if roleauth.MatchesRequirement(groups, roleauth.RequirementAdministrator) {
		return ReadScope{administrator: true, denied: p.readDeniedAll}, nil
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
	roots := make([]string, 0, len(p.readRoots)+len(p.writeRoots)+len(zfsRoots))
	for _, root := range append(append(append([]string{}, p.readRoots...), p.writeRoots...), zfsRoots...) {
		if root == "/" {
			continue
		}
		roots = append(roots, root)
	}
	return ReadScope{roots: roots, denied: append(append([]string{}, p.readDeniedAll...), p.readDeniedNonAdmin...)}, nil
}

// AllowsFile permits regular-file reads only within an approved storage tree.
func (scope ReadScope) AllowsFile(path string) bool {
	return !deniedByMask(scope.denied, path) && (scope.administrator || fileaccess.WithinRoots(path, scope.roots))
}

// AllowsDirectory also permits ancestors needed to navigate to approved storage.
func (scope ReadScope) AllowsDirectory(path string) bool {
	if scope.AllowsFile(path) {
		return true
	}
	if deniedByMask(scope.denied, path) {
		return false
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
	return fileaccess.ParseZFSRoots(mountInfo)
}
