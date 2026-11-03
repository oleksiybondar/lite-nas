package workers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"lite-nas/shared/metrics"
)

func TestParseMountInfoFile(t *testing.T) {
	t.Parallel()

	mountInfoPath := filepath.Join(t.TempDir(), "mountinfo")
	mustWriteCoverageFile(t, mountInfoPath, "39 28 8:2 / / rw,relatime - ext4 /dev/sda2 rw\n")

	entries, err := parseMountInfoFile(mountInfoPath)
	if err != nil {
		t.Fatalf("parseMountInfoFile() error = %v", err)
	}
	if len(entries) != 1 || entries[0].MountPoint != "/" {
		t.Fatalf("parseMountInfoFile() = %#v, want root mount", entries)
	}
}

func TestParseProcMountsFile(t *testing.T) {
	t.Parallel()

	procMountsPath := filepath.Join(t.TempDir(), "mounts")
	mustWriteCoverageFile(t, procMountsPath, "/dev/sda2 / ext4 rw 0 0\n")

	entries, err := parseProcMountsFile(procMountsPath)
	if err != nil {
		t.Fatalf("parseProcMountsFile() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Filesystem != "ext4" {
		t.Fatalf("parseProcMountsFile() = %#v, want ext4 mount", entries)
	}
}

func TestReadMountUsage(t *testing.T) {
	t.Parallel()

	usage, err := readMountUsage(t.TempDir())
	if err != nil {
		t.Fatalf("readMountUsage() error = %v", err)
	}
	if usage.TotalBytes == 0 {
		t.Fatalf("readMountUsage() = %#v, want non-zero totals", usage)
	}
}

func TestResolveSourceDeviceNode(t *testing.T) {
	t.Parallel()

	directNode := resolveSourceDeviceNode("/dev/sdb1")
	if directNode == nil || *directNode != "sdb1" {
		t.Fatalf("resolveSourceDeviceNode(/dev/sdb1) = %v, want sdb1", directNode)
	}

	if unexpected := resolveSourceDeviceNode("/dev/disk/by-id/example"); unexpected != nil {
		t.Fatalf("resolveSourceDeviceNode() = %v, want nil for nested /dev path", unexpected)
	}
}

func TestResolveParentNodeFromSlaves(t *testing.T) {
	t.Parallel()

	devicePath := filepath.Join(t.TempDir(), "dm-0")
	mustMkdirAll(t, filepath.Join(devicePath, "slaves", "sda"))

	parent := resolveParentNode(devicePath, "dm", nil)
	if parent == nil || *parent != "sda" {
		t.Fatalf("resolveParentNode() = %v, want sda", parent)
	}
}

func TestResolveParentNodeSkipsDiskKind(t *testing.T) {
	t.Parallel()

	if noParent := resolveParentNode(t.TempDir(), "disk", nil); noParent != nil {
		t.Fatalf("resolveParentNode() = %v, want nil for disk kind", noParent)
	}
}

func TestClassifyDeviceKind(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	partitionPath := filepath.Join(rootDir, "sda1")
	mustWriteCoverageFile(t, filepath.Join(partitionPath, "partition"), "1\n")

	if got := classifyDeviceKind(partitionPath, "sda1"); got != "partition" {
		t.Fatalf("classifyDeviceKind() = %q, want partition", got)
	}
	if got := classifyDeviceKind(rootDir, "loop0"); got != "loop" {
		t.Fatalf("classifyDeviceKind(loop0) = %q, want loop", got)
	}
	if got := classifyDeviceKind(rootDir, "mystery0"); got != "unknown" {
		t.Fatalf("classifyDeviceKind(mystery0) = %q, want unknown", got)
	}
}

func TestClassifyVirtualConnectionAndSpecialKind(t *testing.T) {
	t.Parallel()

	if value, ok := classifyVirtualConnection("md"); !ok || value != "MD" {
		t.Fatalf("classifyVirtualConnection(md) = %q,%t, want MD,true", value, ok)
	}
	if value, ok := classifyVirtualConnection("disk"); ok || value != "" {
		t.Fatalf("classifyVirtualConnection(disk) = %q,%t, want empty,false", value, ok)
	}
	if got := specialDeviceKind("zd0"); got != "zvol" {
		t.Fatalf("specialDeviceKind(zd0) = %q, want zvol", got)
	}
}

func TestClassifyPhysicalConnectionAndFilesystemCategory(t *testing.T) {
	t.Parallel()

	usbPath := filepath.Join(t.TempDir(), "usb", "sdb")
	mustMkdirAll(t, filepath.Dir(usbPath))

	if got := classifyPhysicalConnection(usbPath); got != "USB" {
		t.Fatalf("classifyPhysicalConnection() = %q, want USB", got)
	}
	if got := classifyFilesystemCategory("fuse.overlayfs"); got != "fuse" {
		t.Fatalf("classifyFilesystemCategory() = %q, want fuse", got)
	}
}

func TestNewPollingWorkerUsesReadMountUsage(t *testing.T) {
	t.Parallel()

	worker := NewPollingWorker(
		"sys",
		"diskstats",
		"mountinfo",
		"mounts",
		make(chan struct{}),
		make(chan metrics.DiskMetricsSnapshot, 1),
		make(chan error, 1),
	)
	if worker.statFS == nil {
		t.Fatal("NewPollingWorker() statFS = nil, want readMountUsage")
	}
}

func TestEmitErrorForwardsError(t *testing.T) {
	t.Parallel()

	errorCh := make(chan error, 1)
	worker := newPollingWorkerWithDependencies("", "", "", "", nil, nil, errorCh, nil)
	worker.emitError(context.Background(), errors.New("boom"))

	select {
	case err := <-errorCh:
		if err == nil {
			t.Fatal("emitError() error = nil, want forwarded error")
		}
	default:
		t.Fatal("emitError() did not forward error")
	}
}

func TestPollAndSendReturnsOnCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker := newPollingWorkerWithDependencies(
		"missing-sys",
		"missing-diskstats",
		"missing-mountinfo",
		"missing-mounts",
		nil,
		make(chan metrics.DiskMetricsSnapshot, 1),
		make(chan error, 1),
		nil,
	)
	worker.pollAndSend(ctx)
}

func mustWriteCoverageFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
