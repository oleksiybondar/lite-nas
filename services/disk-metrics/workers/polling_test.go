package workers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"lite-nas/shared/metrics"
)

// TestCollectDeviceSnapshotsCoversRepresentativeDeviceShapes verifies the
// contract-required SATA, NVMe, USB, and missing-metadata cases.
func TestCollectDeviceSnapshotsCoversRepresentativeDeviceShapes(t *testing.T) {
	t.Parallel()

	fixture := newDiskFixture(t)
	fixture.addSATADiskWithPartitions()
	fixture.addNVMEDiskWithPartition()
	fixture.addUnmountedUSBDevice()
	ioCounters := fixture.diskIOCounters()

	collection, err := collectDeviceSnapshots(fixture.sysBlockPath, ioCounters)
	if err != nil {
		t.Fatalf("collectDeviceSnapshots() error = %v", err)
	}

	devicesByNode := indexDevicesByNode(collection.devices)
	assertDeviceKindAndConnection(t, devicesByNode, "sda", "disk", "SATA")
	assertDeviceKindAndConnection(t, devicesByNode, "sda1", "partition", "SATA")
	assertDeviceKindAndConnection(t, devicesByNode, "sda2", "partition", "SATA")
	assertDeviceParent(t, devicesByNode, "sda1", "sda")
	assertDeviceKindAndConnection(t, devicesByNode, "nvme0n1", "disk", "NVME")
	assertDeviceParent(t, devicesByNode, "nvme0n1p1", "nvme0n1")
	assertDeviceKindAndConnection(t, devicesByNode, "sdb", "disk", "USB")
	assertDeviceMetadataMissing(t, devicesByNode["sdb"])

	if got := devicesByNode["sda"].Partitions; len(got) != 2 {
		t.Fatalf("sda.Partitions length = %d, want 2", len(got))
	}
	if devicesByNode["sda"].IO == nil || devicesByNode["sda"].IO.ReadsCompleted != 1 {
		t.Fatalf("sda.IO = %#v, want parsed diskstats counters", devicesByNode["sda"].IO)
	}
}

// TestBuildMountSnapshotsSeparatesMountKinds verifies tmpfs, NFS/CIFS, and
// local-block mount handling plus filesystem summaries.
func TestBuildMountSnapshotsSeparatesMountKinds(t *testing.T) {
	t.Parallel()

	entries := []mountEntry{
		{Major: 8, Minor: 2, MountPoint: "/", Source: "/dev/sda2", Filesystem: "ext4"},
		{MountPoint: "/run", Source: "tmpfs", Filesystem: "tmpfs"},
		{MountPoint: "/mnt/share", Source: "server:/share", Filesystem: "nfs4", ReadOnly: true},
	}

	mounts, filesystems := buildMountSnapshots(entries, fixtureMountUsage)
	assertLocalBlockMount(t, mounts["/"])
	assertMemoryMount(t, mounts["/run"])
	assertRemoteReadonlyMount(t, mounts["/mnt/share"])
	assertFilesystemSummary(t, filesystems, "ext4", "local_block")
	assertFilesystemSummary(t, filesystems, "tmpfs", "memory")
	assertFilesystemSummary(t, filesystems, "nfs4", "network")
}

// TestParseDiskStatsSupportsLegacyAndExtendedLines verifies diskstats parsing
// with and without discard and flush fields.
func TestParseDiskStatsSupportsLegacyAndExtendedLines(t *testing.T) {
	t.Parallel()

	content := "" +
		"8 0 sda 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17\n" +
		"8 1 sda1 21 22 23 24 25 26 27 28 29 30 31\n"

	stats, err := parseDiskStats(content)
	if err != nil {
		t.Fatalf("parseDiskStats() error = %v", err)
	}

	if stats["sda"].FlushesCompleted == nil || *stats["sda"].FlushesCompleted != 16 {
		t.Fatalf("sda flush counters = %#v, want parsed extended values", stats["sda"])
	}
	if stats["sda1"].DiscardsCompleted != nil || stats["sda1"].FlushesCompleted != nil {
		t.Fatalf("sda1 optional counters = %#v, want nil legacy values", stats["sda1"])
	}
}

// TestParseMountInfoAndFallbackProcMounts verifies mountinfo parsing and the
// proc-mounts fallback path.
func TestParseMountInfo(t *testing.T) {
	t.Parallel()

	mountInfo := "39 28 8:2 / / rw,relatime - ext4 /dev/sda2 rw\n40 28 0:29 / /run rw,nosuid,nodev - tmpfs tmpfs rw,size=1m\n"
	entries, err := parseMountInfo(mountInfo)
	if err != nil {
		t.Fatalf("parseMountInfo() error = %v", err)
	}
	if len(entries) != 2 || entries[0].MountPoint != "/" || entries[1].Filesystem != "tmpfs" {
		t.Fatalf("parseMountInfo() = %#v, want parsed mounts", entries)
	}
}

func TestParseProcMountsFallback(t *testing.T) {
	t.Parallel()

	procMounts := "/dev/sda2 / ext4 rw 0 0\nserver:/share /mnt/share nfs4 ro 0 0\n"
	fallbackEntries, err := parseProcMounts(procMounts)
	if err != nil {
		t.Fatalf("parseProcMounts() error = %v", err)
	}
	if len(fallbackEntries) != 2 || !fallbackEntries[1].ReadOnly {
		t.Fatalf("parseProcMounts() = %#v, want readonly fallback entry", fallbackEntries)
	}
}

// TestPollingWorkerPollBuildsSnapshot verifies one integrated polling cycle
// from fixture-backed sysfs, diskstats, and mountinfo sources.
func TestPollingWorkerPollBuildsSnapshot(t *testing.T) {
	t.Parallel()

	fixture := newDiskFixture(t)
	fixture.addSATADiskWithPartitions()
	fixture.addNVMEDiskWithPartition()
	fixture.addUnmountedUSBDevice()
	fixture.writeDiskStats()
	fixture.writeMountInfo()

	worker := newPollingWorkerWithDependencies(
		fixture.sysBlockPath,
		fixture.procDiskStatsPath,
		fixture.procMountInfoPath,
		fixture.procMountsPath,
		nil,
		nil,
		nil,
		fixture.statFS,
	)

	snapshot, err := worker.poll()
	if err != nil {
		t.Fatalf("poll() error = %v", err)
	}

	assertIntegratedSnapshot(t, snapshot)
	devicesByNode := indexDevicesByNode(snapshot.Devices)
	if len(devicesByNode["sda2"].Mounts) != 1 || devicesByNode["sda2"].Mounts[0] != "/" {
		t.Fatalf("sda2.Mounts = %#v, want root mount link", devicesByNode["sda2"].Mounts)
	}
}

func fixtureMountUsage(path string) (mountUsage, error) {
	switch path {
	case "/":
		return mountUsage{TotalBytes: 1000, UsedBytes: 450, FreeBytes: 550, AvailableBytes: 520, UsedPercent: 45}, nil
	case "/run":
		return mountUsage{TotalBytes: 100, UsedBytes: 10, FreeBytes: 90, AvailableBytes: 90, UsedPercent: 10}, nil
	default:
		return mountUsage{}, errors.New("unavailable")
	}
}

func assertLocalBlockMount(t *testing.T, mount metrics.DiskMountSnapshot) {
	t.Helper()

	if mount.Device == nil || *mount.Device != "sda2" {
		t.Fatalf("mount.Device = %v, want sda2", mount.Device)
	}
	if mount.Remote {
		t.Fatal("mount.Remote = true, want false")
	}
}

func assertMemoryMount(t *testing.T, mount metrics.DiskMountSnapshot) {
	t.Helper()

	if !mount.MemoryBacked {
		t.Fatal("mount.MemoryBacked = false, want true")
	}
}

func assertRemoteReadonlyMount(t *testing.T, mount metrics.DiskMountSnapshot) {
	t.Helper()

	if !mount.Remote || !mount.ReadOnly {
		t.Fatalf("mount = %#v, want remote readonly mount", mount)
	}
	if mount.Device != nil {
		t.Fatalf("mount.Device = %v, want nil", mount.Device)
	}
}

func assertFilesystemSummary(
	t *testing.T,
	filesystems map[string]metrics.DiskFilesystemSummary,
	filesystem string,
	wantType string,
) {
	t.Helper()

	summary := filesystems[filesystem]
	if summary.Type != wantType || summary.MountCount != 1 {
		t.Fatalf("filesystems[%s] = %#v, want %s count 1", filesystem, summary, wantType)
	}
}

func assertIntegratedSnapshot(t *testing.T, snapshot metrics.DiskMetricsSnapshot) {
	t.Helper()

	if len(snapshot.Devices) == 0 {
		t.Fatal("snapshot.Devices length = 0, want discovered devices")
	}
	if _, ok := snapshot.Mounts["/"]; !ok {
		t.Fatal("snapshot.Mounts missing root mount")
	}
	if snapshot.Filesystems["tmpfs"].MountCount != 1 {
		t.Fatalf("snapshot.Filesystems[tmpfs] = %#v, want count 1", snapshot.Filesystems["tmpfs"])
	}
}

// TestPollingWorkerControlFlow verifies tick waiting, context shutdown, and
// asynchronous snapshot publication behavior.
func TestPollingWorkerControlFlow(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticks := make(chan struct{}, 1)
	output := make(chan metrics.DiskMetricsSnapshot, 1)
	errorsCh := make(chan error, 1)
	fixture := newDiskFixture(t)
	fixture.addSATADiskWithPartitions()
	fixture.writeDiskStats()
	fixture.writeMountInfo()
	worker := newPollingWorkerWithDependencies(
		fixture.sysBlockPath,
		fixture.procDiskStatsPath,
		fixture.procMountInfoPath,
		fixture.procMountsPath,
		ticks,
		output,
		errorsCh,
		fixture.statFS,
	)

	ticks <- struct{}{}
	if !worker.waitNextPoll(ctx) {
		t.Fatal("waitNextPoll() = false, want true")
	}

	worker.Start(ctx)
	ticks <- struct{}{}

	select {
	case <-output:
	case <-errorsCh:
		t.Fatal("received poll error, want snapshot")
	}
}

// diskFixture stores one deterministic filesystem fixture for worker tests.
type diskFixture struct {
	t                 *testing.T
	rootDir           string
	sysBlockPath      string
	procDiskStatsPath string
	procMountInfoPath string
	procMountsPath    string
	statFS            statFSFunc
}

// newDiskFixture allocates one worker fixture rooted in a temporary directory.
func newDiskFixture(t *testing.T) diskFixture {
	t.Helper()

	rootDir := t.TempDir()
	fixture := diskFixture{
		t:                 t,
		rootDir:           rootDir,
		sysBlockPath:      filepath.Join(rootDir, "sys", "block"),
		procDiskStatsPath: filepath.Join(rootDir, "proc", "diskstats"),
		procMountInfoPath: filepath.Join(rootDir, "proc", "self", "mountinfo"),
		procMountsPath:    filepath.Join(rootDir, "proc", "mounts"),
	}

	mustMkdirAll(t, fixture.sysBlockPath)
	fixture.statFS = func(path string) (mountUsage, error) {
		switch path {
		case "/":
			return mountUsage{TotalBytes: 1000, UsedBytes: 450, FreeBytes: 550, AvailableBytes: 520, UsedPercent: 45.05}, nil
		case "/run":
			return mountUsage{TotalBytes: 100, UsedBytes: 15, FreeBytes: 85, AvailableBytes: 85, UsedPercent: 15}, nil
		default:
			return mountUsage{}, errors.New("unavailable")
		}
	}

	return fixture
}

// addSATADiskWithPartitions adds one SATA disk fixture with two partitions.
func (f diskFixture) addSATADiskWithPartitions() {
	devicePath := filepath.Join(f.sysBlockPath, "sda")
	f.addBlockDevice(devicePath, "sda", "8:0", "2000", map[string]string{
		"device/vendor":    "ATA\n",
		"device/model":     "WDC Blue\n",
		"device/serial":    "SATA-SERIAL\n",
		"device/wwid":      "wwn-0x50014ee2\n",
		"queue/rotational": "1\n",
		"removable":        "0\n",
	}, filepath.Join(f.rootDir, "devices", "pci0000:00", "0000:00:17.0", "ata1", "host0", "target0:0:0", "0:0:0:0", "block", "sda"))
	f.addPartition(devicePath, "sda1", "8:1", "500")
	f.addPartition(devicePath, "sda2", "8:2", "1500")
}

// addNVMEDiskWithPartition adds one NVMe disk fixture with one partition.
func (f diskFixture) addNVMEDiskWithPartition() {
	devicePath := filepath.Join(f.sysBlockPath, "nvme0n1")
	f.addBlockDevice(devicePath, "nvme0n1", "259:0", "4000", map[string]string{
		"device/model":     "Fast NVMe\n",
		"device/serial":    "NVME-SERIAL\n",
		"queue/rotational": "0\n",
		"removable":        "0\n",
	}, filepath.Join(f.rootDir, "devices", "pci0000:00", "0000:00:1d.0", "nvme", "nvme0", "block", "nvme0n1"))
	f.addPartition(devicePath, "nvme0n1p1", "259:1", "3500")
}

// addUnmountedUSBDevice adds one USB-backed block device without vendor/model
// files to cover optional metadata handling.
func (f diskFixture) addUnmountedUSBDevice() {
	devicePath := filepath.Join(f.sysBlockPath, "sdb")
	f.addBlockDevice(devicePath, "sdb", "8:16", "1000", map[string]string{
		"queue/rotational": "0\n",
		"removable":        "1\n",
	}, filepath.Join(f.rootDir, "devices", "pci0000:00", "0000:00:14.0", "usb1", "1-1", "block", "sdb"))
}

// addBlockDevice writes one base block device fixture and symlinks it from
// /sys/block.
func (f diskFixture) addBlockDevice(sysBlockDevicePath string, node string, dev string, size string, files map[string]string, realPath string) {
	mustMkdirAll(f.t, realPath)
	mustMkdirAll(f.t, filepath.Join(realPath, "device"))
	mustWriteFile(f.t, filepath.Join(realPath, "dev"), dev+"\n")
	mustWriteFile(f.t, filepath.Join(realPath, "size"), size+"\n")

	for relativePath, content := range files {
		mustWriteFile(f.t, filepath.Join(realPath, relativePath), content)
	}

	mustMkdirAll(f.t, filepath.Dir(sysBlockDevicePath))
	mustSymlink(f.t, realPath, sysBlockDevicePath)
	mustMkdirAll(f.t, filepath.Join(realPath, "slaves"))
	_ = node
}

// addPartition writes one child partition fixture under a base block device.
func (f diskFixture) addPartition(basePath string, node string, dev string, size string) {
	realBasePath, err := filepath.EvalSymlinks(basePath)
	if err != nil {
		f.t.Fatalf("EvalSymlinks(%q) error = %v", basePath, err)
	}

	partitionPath := filepath.Join(realBasePath, node)
	mustMkdirAll(f.t, partitionPath)
	mustWriteFile(f.t, filepath.Join(partitionPath, "partition"), "1\n")
	mustWriteFile(f.t, filepath.Join(partitionPath, "dev"), dev+"\n")
	mustWriteFile(f.t, filepath.Join(partitionPath, "size"), size+"\n")
}

// diskIOCounters returns deterministic device I/O counters for fixture devices.
func (f diskFixture) diskIOCounters() map[string]metrics.DiskDeviceIOCounters {
	stats, err := parseDiskStats("" +
		"8 0 sda 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17\n" +
		"8 1 sda1 21 22 23 24 25 26 27 28 29 30 31\n" +
		"8 2 sda2 41 42 43 44 45 46 47 48 49 50 51\n" +
		"259 0 nvme0n1 61 62 63 64 65 66 67 68 69 70 71\n" +
		"259 1 nvme0n1p1 81 82 83 84 85 86 87 88 89 90 91\n" +
		"8 16 sdb 101 102 103 104 105 106 107 108 109 110 111\n")
	if err != nil {
		f.t.Fatalf("parseDiskStats() error = %v", err)
	}

	return stats
}

// writeDiskStats writes the fixture diskstats file.
func (f diskFixture) writeDiskStats() {
	stats := "" +
		"8 0 sda 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17\n" +
		"8 1 sda1 21 22 23 24 25 26 27 28 29 30 31\n" +
		"8 2 sda2 41 42 43 44 45 46 47 48 49 50 51\n" +
		"259 0 nvme0n1 61 62 63 64 65 66 67 68 69 70 71\n" +
		"259 1 nvme0n1p1 81 82 83 84 85 86 87 88 89 90 91\n" +
		"8 16 sdb 101 102 103 104 105 106 107 108 109 110 111\n"
	mustWriteFile(f.t, f.procDiskStatsPath, stats)
}

// writeMountInfo writes deterministic mountinfo and fallback proc mounts
// fixtures.
func (f diskFixture) writeMountInfo() {
	mountInfo := "" +
		"39 28 8:2 / / rw,relatime - ext4 /dev/sda2 rw\n" +
		"40 28 0:29 / /run rw,nosuid,nodev - tmpfs tmpfs rw,size=1m\n" +
		"41 28 0:44 / /mnt/share ro,relatime - nfs4 server:/share ro\n"
	mustWriteFile(f.t, f.procMountInfoPath, mountInfo)
	mustWriteFile(f.t, f.procMountsPath, "/dev/sda2 / ext4 rw 0 0\n")
}

// indexDevicesByNode returns one lookup map keyed by device node.
func indexDevicesByNode(devices []metrics.DiskDeviceSnapshot) map[string]metrics.DiskDeviceSnapshot {
	index := make(map[string]metrics.DiskDeviceSnapshot, len(devices))
	for _, device := range devices {
		index[device.Node] = device
	}

	return index
}

// assertDeviceKindAndConnection verifies the contract classification for one
// device node.
func assertDeviceKindAndConnection(t *testing.T, devices map[string]metrics.DiskDeviceSnapshot, node string, wantKind string, wantConnection string) {
	t.Helper()

	device, ok := devices[node]
	if !ok {
		t.Fatalf("device %q missing from snapshot", node)
	}
	if device.Kind != wantKind || device.ConnectionType != wantConnection {
		t.Fatalf("device %q = %#v, want kind=%s connection=%s", node, device, wantKind, wantConnection)
	}
}

// assertDeviceParent verifies the parent node for one child device.
func assertDeviceParent(t *testing.T, devices map[string]metrics.DiskDeviceSnapshot, node string, wantParent string) {
	t.Helper()

	device := devices[node]
	if device.Parent == nil || *device.Parent != wantParent {
		t.Fatalf("device %q parent = %v, want %s", node, device.Parent, wantParent)
	}
}

// assertDeviceMetadataMissing verifies that optional metadata is omitted rather
// than invented.
func assertDeviceMetadataMissing(t *testing.T, device metrics.DiskDeviceSnapshot) {
	t.Helper()

	if device.Vendor != nil || device.Model != nil || device.Serial != nil || device.WWN != nil {
		t.Fatalf("device optional metadata = %#v, want nil values", device)
	}
}

// mustMkdirAll creates one directory tree for test fixtures.
func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o750); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

// mustWriteFile writes one fixture file and creates its parent directories.
func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

// mustSymlink creates one fixture symlink.
func mustSymlink(t *testing.T, target string, path string) {
	t.Helper()

	if err := os.Symlink(target, path); err != nil {
		t.Fatalf("Symlink(%q, %q) error = %v", target, path, err)
	}
}
