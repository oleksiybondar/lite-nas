package workers

import (
	"context"
	"sort"
	"time"

	"lite-nas/shared/metrics"
)

// PollingWorker periodically reads host disk metric sources and emits disk
// snapshots into an output channel.
type PollingWorker struct {
	sysBlockPath      string
	procDiskStatsPath string
	procSelfMountInfo string
	procMountsPath    string
	ticks             <-chan struct{}
	output            chan<- metrics.DiskMetricsSnapshot
	errors            chan<- error
	statFS            statFSFunc
}

// NewPollingWorker creates a PollingWorker with the dependencies required for
// periodic disk snapshot collection.
func NewPollingWorker(
	sysBlockPath string,
	procDiskStatsPath string,
	procSelfMountInfo string,
	procMountsPath string,
	ticks <-chan struct{},
	output chan<- metrics.DiskMetricsSnapshot,
	errors chan<- error,
) PollingWorker {
	return newPollingWorkerWithDependencies(
		sysBlockPath,
		procDiskStatsPath,
		procSelfMountInfo,
		procMountsPath,
		ticks,
		output,
		errors,
		readMountUsage,
	)
}

// newPollingWorkerWithDependencies creates a PollingWorker with injected
// helpers for deterministic unit testing.
func newPollingWorkerWithDependencies(
	sysBlockPath string,
	procDiskStatsPath string,
	procSelfMountInfo string,
	procMountsPath string,
	ticks <-chan struct{},
	output chan<- metrics.DiskMetricsSnapshot,
	errors chan<- error,
	statFS statFSFunc,
) PollingWorker {
	return PollingWorker{
		sysBlockPath:      sysBlockPath,
		procDiskStatsPath: procDiskStatsPath,
		procSelfMountInfo: procSelfMountInfo,
		procMountsPath:    procMountsPath,
		ticks:             ticks,
		output:            output,
		errors:            errors,
		statFS:            statFS,
	}
}

// Start launches the polling worker in a separate goroutine.
func (w PollingWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run executes the polling loop until the provided context is canceled.
func (w PollingWorker) run(ctx context.Context) {
	for {
		if !w.waitNextPoll(ctx) {
			return
		}

		w.pollAndSend(ctx)
	}
}

// waitNextPoll blocks until the next poll tick arrives or the context is
// canceled.
func (w PollingWorker) waitNextPoll(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case _, ok := <-w.ticks:
		return ok
	}
}

// pollAndSend performs one polling cycle and forwards the resulting snapshot.
func (w PollingWorker) pollAndSend(ctx context.Context) {
	snapshot, err := w.poll()
	if err != nil {
		w.emitError(ctx, err)
		return
	}

	select {
	case <-ctx.Done():
		return
	case w.output <- snapshot:
	}
}

// emitError reports one polling error to the runtime when possible.
func (w PollingWorker) emitError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	select {
	case <-ctx.Done():
		return
	case w.errors <- err:
	default:
	}
}

// poll reads all disk sources required for one snapshot cycle.
func (w PollingWorker) poll() (metrics.DiskMetricsSnapshot, error) {
	ioCounters, err := parseDiskStatsFile(w.procDiskStatsPath)
	if err != nil {
		return metrics.DiskMetricsSnapshot{}, err
	}

	deviceCollection, err := collectDeviceSnapshots(w.sysBlockPath, ioCounters)
	if err != nil {
		return metrics.DiskMetricsSnapshot{}, err
	}

	mountEntries, err := collectMountEntries(w.procSelfMountInfo, w.procMountsPath)
	if err != nil {
		return metrics.DiskMetricsSnapshot{}, err
	}

	mounts, filesystems := buildMountSnapshots(mountEntries, w.statFS)
	linkMountsToDevices(deviceCollection, mountEntries)

	return metrics.DiskMetricsSnapshot{
		Timestamp:   time.Now(),
		Devices:     deviceCollection.devices,
		Filesystems: filesystems,
		Mounts:      mounts,
	}, nil
}

// deviceCollection stores device snapshots together with lookup indexes used
// during mount-to-device linking.
type deviceCollection struct {
	devices      []metrics.DiskDeviceSnapshot
	byDevNumbers map[string]int
}

// collectDeviceSnapshots reads sysfs block devices and returns snapshot DTOs
// plus helper indexes for later mount linking.
func collectDeviceSnapshots(
	sysBlockPath string,
	ioCounters map[string]metrics.DiskDeviceIOCounters,
) (deviceCollection, error) {
	baseNodes, err := listBaseBlockNodes(sysBlockPath)
	if err != nil {
		return deviceCollection{}, err
	}

	collection := deviceCollection{
		devices:      make([]metrics.DiskDeviceSnapshot, 0, len(baseNodes)),
		byDevNumbers: make(map[string]int),
	}

	for _, baseNode := range baseNodes {
		if err := appendBaseDeviceSnapshots(&collection, sysBlockPath, baseNode, ioCounters); err != nil {
			return deviceCollection{}, err
		}
	}

	sort.Slice(collection.devices, func(i int, j int) bool {
		return collection.devices[i].Node < collection.devices[j].Node
	})

	collection.byDevNumbers = rebuildDeviceIndex(collection.devices)
	return collection, nil
}

// appendBaseDeviceSnapshots appends one base device plus any discovered partitions.
func appendBaseDeviceSnapshots(
	collection *deviceCollection,
	sysBlockPath string,
	baseNode string,
	ioCounters map[string]metrics.DiskDeviceIOCounters,
) error {
	basePath := joinPath(sysBlockPath, baseNode)
	baseSnapshot, err := readBlockDeviceSnapshot(basePath, baseNode, nil, ioCounters[baseNode])
	if err != nil {
		return err
	}

	partitions, err := listPartitionNodes(basePath)
	if err != nil {
		return err
	}

	baseSnapshot.Partitions = append(baseSnapshot.Partitions, partitions...)
	appendDeviceSnapshot(collection, baseSnapshot)

	return appendPartitionSnapshots(collection, basePath, baseSnapshot.Node, partitions, ioCounters)
}

// appendPartitionSnapshots appends child partition snapshots for one base device.
func appendPartitionSnapshots(
	collection *deviceCollection,
	basePath string,
	parentNode string,
	partitions []string,
	ioCounters map[string]metrics.DiskDeviceIOCounters,
) error {
	for _, partitionNode := range partitions {
		partitionPath := joinPath(basePath, partitionNode)
		partitionSnapshot, err := readBlockDeviceSnapshot(
			partitionPath,
			partitionNode,
			&parentNode,
			ioCounters[partitionNode],
		)
		if err != nil {
			return err
		}

		appendDeviceSnapshot(collection, partitionSnapshot)
	}

	return nil
}

func appendDeviceSnapshot(collection *deviceCollection, snapshot metrics.DiskDeviceSnapshot) {
	index := len(collection.devices)
	collection.devices = append(collection.devices, snapshot)
	collection.byDevNumbers[formatDevNumbers(snapshot.Major, snapshot.Minor)] = index
}

// rebuildDeviceIndex rebuilds the dev-number lookup after a device slice has
// been reordered.
func rebuildDeviceIndex(devices []metrics.DiskDeviceSnapshot) map[string]int {
	index := make(map[string]int, len(devices))
	for i := range devices {
		index[formatDevNumbers(devices[i].Major, devices[i].Minor)] = i
	}

	return index
}

// buildMountSnapshots converts parsed mount entries into snapshot DTOs and
// derives filesystem summaries from them.
func buildMountSnapshots(
	entries []mountEntry,
	statFS statFSFunc,
) (map[string]metrics.DiskMountSnapshot, map[string]metrics.DiskFilesystemSummary) {
	mounts := make(map[string]metrics.DiskMountSnapshot, len(entries))
	filesystems := make(map[string]metrics.DiskFilesystemSummary)

	for _, entry := range entries {
		snapshot := metrics.DiskMountSnapshot{
			Mountpoint:   entry.MountPoint,
			Source:       entry.Source,
			Filesystem:   entry.Filesystem,
			Device:       resolveSourceDeviceNode(entry.Source),
			ReadOnly:     entry.ReadOnly,
			Remote:       isRemoteFilesystem(entry.Filesystem),
			MemoryBacked: isMemoryFilesystem(entry.Filesystem),
		}

		if usage, err := statFS(entry.MountPoint); err == nil {
			snapshot.TotalBytes = uint64Pointer(usage.TotalBytes)
			snapshot.UsedBytes = uint64Pointer(usage.UsedBytes)
			snapshot.FreeBytes = uint64Pointer(usage.FreeBytes)
			snapshot.AvailableBytes = uint64Pointer(usage.AvailableBytes)
			snapshot.UsedPercent = float64Pointer(usage.UsedPercent)
		}

		mounts[entry.MountPoint] = snapshot

		filesystemSummary := filesystems[entry.Filesystem]
		filesystemSummary.Type = classifyFilesystemCategory(entry.Filesystem)
		filesystemSummary.MountCount++
		filesystems[entry.Filesystem] = filesystemSummary
	}

	return mounts, filesystems
}

// linkMountsToDevices adds mountpoint references to local device entries when
// the mount table exposes matching major/minor numbers.
func linkMountsToDevices(collection deviceCollection, entries []mountEntry) {
	for _, entry := range entries {
		index, ok := collection.byDevNumbers[formatDevNumbers(entry.Major, entry.Minor)]
		if !ok {
			continue
		}

		device := &collection.devices[index]
		device.Mounts = append(device.Mounts, entry.MountPoint)
	}

	for i := range collection.devices {
		sort.Strings(collection.devices[i].Mounts)
		sort.Strings(collection.devices[i].Partitions)
	}
}
