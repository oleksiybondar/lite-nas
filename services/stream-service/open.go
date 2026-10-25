package main

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// openAuthorizedFile opens a canonical path without following a replaced symlink.
// The inode must still match the object observed before the RBAC decision.
func openAuthorizedFile(path string, before os.FileInfo) (*os.File, error) {
	root, err := unix.Open("/", unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(root)
	fd, err := unix.Openat2(root, strings.TrimPrefix(path, "/"), &unix.OpenHow{
		Flags:   unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW,
		Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS,
	})
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		file.Close()
		return nil, errors.New("authorized file changed before opening")
	}
	return file, nil
}
