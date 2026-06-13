package workers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"lite-nas/shared/metrics"
)

const diskSectorSizeBytes = 512

// listBaseBlockNodes lists base block-device nodes from /sys/block in stable
// order.
func listBaseBlockNodes(sysBlockPath string) ([]string, error) {
	entries, err := os.ReadDir(sysBlockPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", sysBlockPath, err)
	}

	nodes := make([]string, 0, len(entries))
	for _, entry := range entries {
		entryType := entry.Type()
		if entry.IsDir() || entryType&os.ModeSymlink != 0 {
			nodes = append(nodes, entry.Name())
		}
	}

	sort.Strings(nodes)
	return nodes, nil
}

// listPartitionNodes lists child partition nodes for one base block device.
func listPartitionNodes(basePath string) ([]string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", basePath, err)
	}

	partitions := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		partitionPath := joinPath(basePath, entry.Name())
		if !fileExists(joinPath(partitionPath, "partition")) {
			continue
		}

		partitions = append(partitions, entry.Name())
	}

	sort.Strings(partitions)
	return partitions, nil
}

// readBlockDeviceSnapshot reads one block-device-like entry from sysfs.
func readBlockDeviceSnapshot(
	devicePath string,
	node string,
	parent *string,
	ioCounters metrics.DiskDeviceIOCounters,
) (metrics.DiskDeviceSnapshot, error) {
	major, minor, err := readDeviceNumbers(joinPath(devicePath, "dev"))
	if err != nil {
		return metrics.DiskDeviceSnapshot{}, err
	}

	sizeSectors, err := readRequiredUint(joinPath(devicePath, "size"))
	if err != nil {
		return metrics.DiskDeviceSnapshot{}, err
	}

	kind := classifyDeviceKind(devicePath, node)
	device := metrics.DiskDeviceSnapshot{
		Node:           node,
		Kind:           kind,
		Parent:         resolveParentNode(devicePath, kind, parent),
		Major:          major,
		Minor:          minor,
		SizeBytes:      sizeSectors * diskSectorSizeBytes,
		Name:           buildDeviceName(devicePath),
		Vendor:         readOptionalStringCandidate(joinPath(devicePath, "device", "vendor")),
		Model:          readOptionalStringCandidate(joinPath(devicePath, "device", "model")),
		Description:    buildDeviceDescription(devicePath),
		Serial:         readOptionalStringCandidates(joinPath(devicePath, "device", "serial"), joinPath(devicePath, "serial")),
		WWN:            readOptionalStringCandidates(joinPath(devicePath, "device", "wwid"), joinPath(devicePath, "wwid")),
		ConnectionType: classifyConnectionType(devicePath, node, kind),
		Rotational:     readOptionalBool(joinPath(devicePath, "queue", "rotational")),
		Removable:      readOptionalBool(joinPath(devicePath, "removable")),
	}

	if ioCounters != (metrics.DiskDeviceIOCounters{}) {
		ioCopy := ioCounters
		device.IO = &ioCopy
	}

	return device, nil
}

// resolveParentNode resolves the parent node for partitions and derived
// devices.
func resolveParentNode(devicePath string, kind string, explicitParent *string) *string {
	if explicitParent != nil {
		return explicitParent
	}
	if !requiresSlaveParent(kind) {
		return nil
	}

	return firstSlaveParent(devicePath)
}

// classifyDeviceKind classifies one sysfs block-device-like entry.
func classifyDeviceKind(devicePath string, node string) string {
	switch {
	case isPartitionDevice(devicePath):
		return "partition"
	case specialDeviceKind(node) != "":
		return specialDeviceKind(node)
	case hasDirectDiskSignature(devicePath, node):
		return "disk"
	default:
		return "unknown"
	}
}

// classifyConnectionType classifies one device transport or family.
func classifyConnectionType(devicePath string, node string, kind string) string {
	if connectionType, ok := classifyVirtualConnection(kind); ok {
		return connectionType
	}

	if strings.HasPrefix(node, "nvme") {
		return "NVME"
	}
	if strings.HasPrefix(node, "mmcblk") {
		return "MMC"
	}

	return classifyPhysicalConnection(devicePath)
}

// classifyFilesystemCategory classifies one observed filesystem type into the
// service-level summary category.
func classifyFilesystemCategory(filesystem string) string {
	if category, ok := filesystemCategories[filesystem]; ok {
		return category
	}

	if filesystem == "fuse" || strings.HasPrefix(filesystem, "fuse.") {
		return "fuse"
	}

	return "unknown"
}

// isRemoteFilesystem reports whether one observed filesystem is network-backed.
func isRemoteFilesystem(filesystem string) bool {
	switch filesystem {
	case "cifs", "smb3", "nfs", "nfs4":
		return true
	default:
		return false
	}
}

// isMemoryFilesystem reports whether one observed filesystem is memory-backed.
func isMemoryFilesystem(filesystem string) bool {
	switch filesystem {
	case "tmpfs", "ramfs":
		return true
	default:
		return false
	}
}

// buildDeviceName derives one optional human-friendly name from sysfs model
// strings.
func buildDeviceName(devicePath string) *string {
	model := readOptionalStringCandidate(joinPath(devicePath, "device", "model"))
	if model != nil {
		return model
	}

	return nil
}

// buildDeviceDescription derives one optional human-friendly description from
// vendor and model strings.
func buildDeviceDescription(devicePath string) *string {
	vendor := readOptionalStringCandidate(joinPath(devicePath, "device", "vendor"))
	model := readOptionalStringCandidate(joinPath(devicePath, "device", "model"))
	if vendor == nil && model == nil {
		return nil
	}

	parts := make([]string, 0, 2)
	if vendor != nil {
		parts = append(parts, *vendor)
	}
	if model != nil {
		parts = append(parts, *model)
	}

	description := strings.Join(parts, " ")
	return &description
}

// readDeviceNumbers reads one sysfs dev file formatted as "<major>:<minor>".
func readDeviceNumbers(path string) (uint64, uint64, error) {
	content, err := readTrustedFile(path)
	if err != nil {
		return 0, 0, fmt.Errorf("read %s: %w", path, err)
	}

	return parseDevNumbers(strings.TrimSpace(string(content)))
}

// readRequiredUint reads one required unsigned integer file.
func readRequiredUint(path string) (uint64, error) {
	content, err := readTrustedFile(path)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", path, err)
	}

	value, err := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse uint from %s: %w", path, err)
	}

	return value, nil
}

// readOptionalBool reads one optional sysfs boolean marker represented as 0 or
// 1.
func readOptionalBool(path string) *bool {
	value, err := readRequiredUint(path)
	if err != nil {
		return nil
	}

	boolValue := value != 0
	return &boolValue
}

// readOptionalStringCandidate reads one optional string file and trims it.
func readOptionalStringCandidate(path string) *string {
	return readOptionalStringCandidates(path)
}

// readOptionalStringCandidates returns the first non-empty readable string from
// the provided candidate paths.
func readOptionalStringCandidates(paths ...string) *string {
	for _, path := range paths {
		content, err := readTrustedFile(path)
		if err != nil {
			continue
		}

		value := strings.TrimSpace(string(content))
		if value == "" {
			continue
		}

		return &value
	}

	return nil
}

// resolveRealPath resolves one sysfs path for transport heuristics and falls
// back to the original path on resolution failure.
func resolveRealPath(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}

	return resolved
}

// fileExists reports whether one path currently exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// joinPath joins path elements using filepath.Join.
func joinPath(parts ...string) string {
	return filepath.Join(parts...)
}

// formatDevNumbers formats one major/minor pair for internal lookups.
func formatDevNumbers(major uint64, minor uint64) string {
	return fmt.Sprintf("%d:%d", major, minor)
}

// uint64Pointer allocates one uint64 pointer.
func uint64Pointer(value uint64) *uint64 {
	return &value
}

// float64Pointer allocates one float64 pointer.
func float64Pointer(value float64) *float64 {
	return &value
}

var filesystemCategories = map[string]string{
	"ext4":       "local_block",
	"xfs":        "local_block",
	"btrfs":      "local_block",
	"vfat":       "local_block",
	"exfat":      "local_block",
	"ntfs":       "local_block",
	"zfs":        "local_block",
	"cifs":       "network",
	"smb3":       "network",
	"nfs":        "network",
	"nfs4":       "network",
	"tmpfs":      "memory",
	"ramfs":      "memory",
	"proc":       "pseudo",
	"sysfs":      "pseudo",
	"devtmpfs":   "pseudo",
	"devpts":     "pseudo",
	"cgroup":     "pseudo",
	"cgroup2":    "pseudo",
	"securityfs": "pseudo",
	"debugfs":    "pseudo",
	"tracefs":    "pseudo",
	"configfs":   "pseudo",
	"fusectl":    "pseudo",
	"mqueue":     "pseudo",
	"hugetlbfs":  "pseudo",
	"pstore":     "pseudo",
	"efivarfs":   "pseudo",
	"bpf":        "pseudo",
}

func requiresSlaveParent(kind string) bool {
	switch kind {
	case "dm", "md", "zvol", "unknown":
		return true
	default:
		return false
	}
}

func firstSlaveParent(devicePath string) *string {
	entries, err := os.ReadDir(joinPath(devicePath, "slaves"))
	if err != nil || len(entries) == 0 {
		return nil
	}

	sort.Slice(entries, func(i int, j int) bool { return entries[i].Name() < entries[j].Name() })
	parent := entries[0].Name()
	return &parent
}

func isPartitionDevice(devicePath string) bool {
	return fileExists(joinPath(devicePath, "partition"))
}

func hasDirectDiskSignature(devicePath string, node string) bool {
	return fileExists(joinPath(devicePath, "device")) ||
		strings.HasPrefix(node, "nvme") ||
		strings.HasPrefix(node, "mmcblk")
}

func classifyVirtualConnection(kind string) (string, bool) {
	switch kind {
	case "loop":
		return "LOOP", true
	case "dm":
		return "DM", true
	case "md":
		return "MD", true
	default:
		return "", false
	}
}

func classifyPhysicalConnection(devicePath string) string {
	realPath := strings.ToLower(resolveRealPath(devicePath))

	for _, marker := range connectionMarkers {
		if strings.Contains(realPath, marker.path) {
			return marker.connectionType
		}
	}

	return "UNKNOWN"
}

type connectionMarker struct {
	path           string
	connectionType string
}

var connectionMarkers = []connectionMarker{
	{path: "/usb", connectionType: "USB"},
	{path: "/virtio", connectionType: "VIRTIO"},
	{path: "/ata", connectionType: "SATA"},
	{path: "/sas", connectionType: "SAS"},
	{path: "/ide", connectionType: "IDE"},
	{path: "/scsi", connectionType: "SCSI"},
}

func specialDeviceKind(node string) string {
	switch {
	case strings.HasPrefix(node, "loop"):
		return "loop"
	case strings.HasPrefix(node, "dm-"):
		return "dm"
	case strings.HasPrefix(node, "md"):
		return "md"
	case strings.HasPrefix(node, "zd"):
		return "zvol"
	default:
		return ""
	}
}
