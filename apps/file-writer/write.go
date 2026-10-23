package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sys/unix"
)

// result is the bounded machine-readable response for one completed operation.
type result struct {
	Operation    string `json:"operation"`
	BytesWritten int64  `json:"bytes_written"`
	NextOffset   int64  `json:"next_offset"`
	Device       uint64 `json:"device"`
	Inode        uint64 `json:"inode"`
	UID          uint32 `json:"uid"`
	Size         int64  `json:"size"`
	SHA256       string `json:"sha256,omitempty"`
}

// operationError reports whether cleanup succeeded after a failed publication.
type operationError struct {
	Code    string
	Message string
	Cleanup string
}

// Error returns the publication failure for ordinary Go error handling.
func (failure *operationError) Error() string {
	return failure.Message
}

// run validates the CLI boundary and emits one JSON result for a successful operation.
func run(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 || slices.Contains(args, "--help") || slices.Contains(args, "-h") {
		_, err := io.WriteString(stdout, helpText)
		return err
	}
	inv, err := parseInvocation(args)
	if err != nil {
		return err
	}
	if os.Geteuid() == 0 {
		return errors.New("writer must run as a non-root user")
	}
	var response result
	switch inv.Command {
	case "create":
		response, err = createStage(inv)
	case "write":
		response, err = writeStage(inv, stdin)
	case "stat":
		response, err = statStage(inv)
	case "checksum":
		response, err = checksumStage(inv)
	case "truncate":
		response, err = truncateStage(inv)
	case "publish":
		response, err = publishStage(inv)
	case "cleanup":
		response, err = cleanupStage(inv)
	}
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(response)
}

// openParent opens the stage's parent without following any ancestor symlink.
func openParent(target string) (int, string, error) {
	how := &unix.OpenHow{
		Flags:   unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC,
		Resolve: unix.RESOLVE_NO_SYMLINKS,
	}
	fd, err := unix.Openat2(unix.AT_FDCWD, filepath.Dir(target), how)
	if err != nil {
		return -1, "", fmt.Errorf("open parent directory: %w", err)
	}
	return fd, filepath.Base(target), nil
}

// stageIdentity builds the stable identity fields returned to File Service.
func stageIdentity(command string, stat *unix.Stat_t) result {
	return result{
		Operation: command,
		Device:    uint64(stat.Dev),
		Inode:     stat.Ino,
		UID:       stat.Uid,
		Size:      stat.Size,
	}
}

// createStage exclusively creates and syncs a destination-adjacent stage as the effective user.
func createStage(inv invocation) (result, error) {
	parent, name, err := openParent(inv.Target)
	if err != nil {
		return result{}, err
	}
	defer unix.Close(parent)
	fd, err := unix.Openat(parent, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return result{}, fmt.Errorf("create stage: %w", err)
	}
	stage := os.NewFile(uintptr(fd), inv.Target)
	defer stage.Close()
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return result{}, fmt.Errorf("stat new stage: %w", err)
	}
	if err := stage.Sync(); err != nil {
		return result{}, fmt.Errorf("sync new stage: %w", err)
	}
	if err := unix.Fsync(parent); err != nil {
		return result{}, fmt.Errorf("sync stage directory: %w", err)
	}
	return stageIdentity(inv.Command, &stat), nil
}

// openStage opens a regular stage through its pinned parent and compares the supplied identity.
func openStage(inv invocation, flags int) (int, string, *os.File, unix.Stat_t, error) {
	parent, name, err := openParent(inv.Target)
	if err != nil {
		return -1, "", nil, unix.Stat_t{}, err
	}
	fd, err := unix.Openat(parent, name, flags|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		unix.Close(parent)
		return -1, "", nil, unix.Stat_t{}, fmt.Errorf("open stage: %w", err)
	}
	stage := os.NewFile(uintptr(fd), inv.Target)
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		stage.Close()
		unix.Close(parent)
		return -1, "", nil, unix.Stat_t{}, fmt.Errorf("stat stage: %w", err)
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Nlink != 1 ||
		uint64(stat.Dev) != inv.ExpectedDevice || stat.Ino != inv.ExpectedInode ||
		int64(stat.Uid) != inv.ExpectedUID || stat.Uid != uint32(os.Geteuid()) {
		stage.Close()
		unix.Close(parent)
		return -1, "", nil, unix.Stat_t{}, errors.New("stage type, identity, or owner differs from expectation")
	}
	return parent, name, stage, stat, nil
}

// checkStageEntry rejects a stage pathname replaced after its descriptor was opened.
func checkStageEntry(parent int, name string, stat unix.Stat_t) error {
	var current unix.Stat_t
	if err := unix.Fstatat(parent, name, &current, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return fmt.Errorf("recheck stage entry: %w", err)
	}
	if current.Mode&unix.S_IFMT != unix.S_IFREG ||
		current.Dev != stat.Dev || current.Ino != stat.Ino || current.Uid != stat.Uid {
		return errors.New("stage path identity changed")
	}
	return nil
}

// writeStage replaces an uncommitted tail with exactly the declared stdin bytes.
func writeStage(inv invocation, stdin io.Reader) (result, error) {
	parent, name, stage, stat, err := openStage(inv, unix.O_RDWR)
	if err != nil {
		return result{}, err
	}
	defer unix.Close(parent)
	defer stage.Close()
	if err := unix.Flock(int(stage.Fd()), unix.LOCK_EX); err != nil {
		return result{}, fmt.Errorf("lock stage: %w", err)
	}
	if err := unix.Fstat(int(stage.Fd()), &stat); err != nil {
		return result{}, fmt.Errorf("stat locked stage: %w", err)
	}
	if err := checkStageEntry(parent, name, stat); err != nil {
		return result{}, err
	}
	if inv.Offset > stat.Size {
		return result{}, errors.New("offset exceeds current stage size")
	}
	if err := stage.Truncate(inv.Offset); err != nil {
		return result{}, fmt.Errorf("truncate uncommitted tail: %w", err)
	}
	if _, err := stage.Seek(inv.Offset, io.SeekStart); err != nil {
		return result{}, fmt.Errorf("seek stage: %w", err)
	}
	written, copyErr := io.CopyN(stage, stdin, inv.Length)
	if copyErr == nil {
		var n int64
		n, copyErr = io.CopyN(io.Discard, stdin, 1)
		if n != 0 {
			copyErr = errors.New("input exceeds declared length")
		} else if errors.Is(copyErr, io.EOF) {
			copyErr = nil
		}
	}
	if copyErr != nil {
		if err := stage.Truncate(inv.Offset); err != nil {
			return result{}, fmt.Errorf("copy input: %v; recover stage: %w", copyErr, err)
		}
		if err := stage.Sync(); err != nil {
			return result{}, fmt.Errorf("copy input: %v; sync recovery: %w", copyErr, err)
		}
		return result{}, fmt.Errorf("copy input: %w", copyErr)
	}
	if err := stage.Sync(); err != nil {
		return result{}, fmt.Errorf("sync stage: %w", err)
	}
	response := stageIdentity(inv.Command, &stat)
	response.BytesWritten = written
	response.NextOffset = inv.Offset + written
	response.Size = response.NextOffset
	return response, nil
}

// statStage reports the current size of the identified stage.
func statStage(inv invocation) (result, error) {
	parent, name, stage, stat, err := openStage(inv, unix.O_RDONLY)
	if err != nil {
		return result{}, err
	}
	defer unix.Close(parent)
	defer stage.Close()
	if err := unix.Flock(int(stage.Fd()), unix.LOCK_SH); err != nil {
		return result{}, fmt.Errorf("lock stage: %w", err)
	}
	if err := unix.Fstat(int(stage.Fd()), &stat); err != nil {
		return result{}, fmt.Errorf("stat stage: %w", err)
	}
	if err := checkStageEntry(parent, name, stat); err != nil {
		return result{}, err
	}
	return stageIdentity(inv.Command, &stat), nil
}

// checksumStage reads the identified stage under a shared lock and reports SHA-256.
func checksumStage(inv invocation) (result, error) {
	parent, name, stage, stat, err := openStage(inv, unix.O_RDONLY)
	if err != nil {
		return result{}, err
	}
	defer unix.Close(parent)
	defer stage.Close()
	if err := unix.Flock(int(stage.Fd()), unix.LOCK_SH); err != nil {
		return result{}, fmt.Errorf("lock stage: %w", err)
	}
	if err := unix.Fstat(int(stage.Fd()), &stat); err != nil {
		return result{}, fmt.Errorf("stat locked stage: %w", err)
	}
	if err := checkStageEntry(parent, name, stat); err != nil {
		return result{}, err
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, stage); err != nil {
		return result{}, fmt.Errorf("read stage: %w", err)
	}
	var after unix.Stat_t
	if err := unix.Fstat(int(stage.Fd()), &after); err != nil {
		return result{}, fmt.Errorf("stat stage after checksum: %w", err)
	}
	if stat.Size != after.Size || stat.Mtim != after.Mtim || stat.Ctim != after.Ctim {
		return result{}, errors.New("stage changed during checksum")
	}
	if err := checkStageEntry(parent, name, after); err != nil {
		return result{}, err
	}
	response := stageIdentity(inv.Command, &after)
	response.SHA256 = hex.EncodeToString(hash.Sum(nil))
	return response, nil
}

// truncateStage discards uncommitted bytes after the supplied offset and syncs recovery.
func truncateStage(inv invocation) (result, error) {
	parent, name, stage, stat, err := openStage(inv, unix.O_RDWR)
	if err != nil {
		return result{}, err
	}
	defer unix.Close(parent)
	defer stage.Close()
	if err := unix.Flock(int(stage.Fd()), unix.LOCK_EX); err != nil {
		return result{}, fmt.Errorf("lock stage: %w", err)
	}
	if err := unix.Fstat(int(stage.Fd()), &stat); err != nil {
		return result{}, fmt.Errorf("stat locked stage: %w", err)
	}
	if err := checkStageEntry(parent, name, stat); err != nil {
		return result{}, err
	}
	if inv.Offset > stat.Size {
		return result{}, errors.New("offset exceeds current stage size")
	}
	if err := stage.Truncate(inv.Offset); err != nil {
		return result{}, fmt.Errorf("truncate stage: %w", err)
	}
	if err := stage.Sync(); err != nil {
		return result{}, fmt.Errorf("sync stage: %w", err)
	}
	response := stageIdentity(inv.Command, &stat)
	response.NextOffset = inv.Offset
	response.Size = inv.Offset
	return response, nil
}

// cleanupStage removes only the identified stage entry and syncs its directory.
func cleanupStage(inv invocation) (result, error) {
	parent, name, stage, stat, err := openStage(inv, unix.O_RDONLY)
	if err != nil {
		return result{}, err
	}
	defer unix.Close(parent)
	defer stage.Close()
	if err := unix.Flock(int(stage.Fd()), unix.LOCK_EX); err != nil {
		return result{}, fmt.Errorf("lock stage: %w", err)
	}
	if err := checkStageEntry(parent, name, stat); err != nil {
		return result{}, err
	}
	if err := unix.Unlinkat(parent, name, 0); err != nil {
		return result{}, fmt.Errorf("remove stage: %w", err)
	}
	if err := unix.Fsync(parent); err != nil {
		return result{}, fmt.Errorf("sync stage directory: %w", err)
	}
	return stageIdentity(inv.Command, &stat), nil
}

// publishStage checks the destination then atomically renames the identified stage.
func publishStage(inv invocation) (result, error) {
	parent, name, stage, stat, err := openStage(inv, unix.O_RDONLY)
	if err != nil {
		return result{}, err
	}
	defer unix.Close(parent)
	defer stage.Close()
	if err := unix.Flock(int(stage.Fd()), unix.LOCK_EX); err != nil {
		return result{}, fmt.Errorf("lock stage: %w", err)
	}
	destination := strings.TrimSuffix(name, ".upload")
	if err := checkStageEntry(parent, name, stat); err != nil {
		return result{}, publishFailure(parent, name, stat, err)
	}
	var current unix.Stat_t
	destinationErr := unix.Fstatat(parent, destination, &current, unix.AT_SYMLINK_NOFOLLOW)
	if inv.DestinationAbsent {
		if !errors.Is(destinationErr, unix.ENOENT) {
			return result{}, publishFailure(parent, name, stat, errors.New("destination is not absent"))
		}
	} else {
		if destinationErr != nil || current.Mode&unix.S_IFMT != unix.S_IFREG ||
			uint64(current.Dev) != inv.DestinationDevice || current.Ino != inv.DestinationInode {
			return result{}, publishFailure(parent, name, stat, errors.New("destination identity changed"))
		}
	}
	mode := uint32(0o664)
	if inv.Mode == "0660" {
		mode = 0o660
	}
	if inv.DestinationAbsent {
		err = unix.Renameat2(parent, name, parent, destination, unix.RENAME_NOREPLACE)
	} else {
		// Linux cannot compare the destination inode atomically with replacement.
		err = unix.Renameat2(parent, name, parent, destination, 0)
	}
	if err != nil {
		return result{}, publishFailure(parent, name, stat, fmt.Errorf("rename stage: %w", err))
	}
	if err := unix.Fchmod(int(stage.Fd()), mode); err != nil {
		return result{}, &operationError{Code: "publish_mode_failed", Message: fmt.Sprintf("set published mode: %v", err), Cleanup: "not_possible"}
	}
	if err := stage.Sync(); err != nil {
		return result{}, &operationError{Code: "publish_sync_failed", Message: fmt.Sprintf("sync published file: %v", err), Cleanup: "not_possible"}
	}
	if err := unix.Fsync(parent); err != nil {
		return result{}, &operationError{Code: "publish_sync_failed", Message: fmt.Sprintf("sync published directory: %v", err), Cleanup: "not_possible"}
	}
	return stageIdentity(inv.Command, &stat), nil
}

// publishFailure attempts safe stage cleanup and records whether it succeeded.
func publishFailure(parent int, name string, stat unix.Stat_t, cause error) error {
	if err := checkStageEntry(parent, name, stat); err != nil {
		return &operationError{Code: "publish_failed_cleanup_skipped", Message: fmt.Sprintf("%v; %v", cause, err), Cleanup: "skipped"}
	}
	if err := unix.Unlinkat(parent, name, 0); err != nil {
		return &operationError{Code: "publish_failed_cleanup_failed", Message: fmt.Sprintf("%v; remove stage: %v", cause, err), Cleanup: "failed"}
	}
	if err := unix.Fsync(parent); err != nil {
		return &operationError{Code: "publish_failed_cleanup_failed", Message: fmt.Sprintf("%v; sync cleanup: %v", cause, err), Cleanup: "failed"}
	}
	return &operationError{Code: "publish_failed_cleaned", Message: cause.Error(), Cleanup: "succeeded"}
}
