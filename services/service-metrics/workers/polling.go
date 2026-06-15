package workers

import (
	"bufio"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"lite-nas/shared/metrics"
)

const (
	serviceManagerName = "systemd"
	serviceUnitType    = "service"
)

// PollingWorker periodically reads host service-manager sources and emits
// service snapshots into an output channel.
type PollingWorker struct {
	unitRoots      []string
	systemSliceDir string
	ticks          <-chan struct{}
	output         chan<- metrics.ServiceMetricsSnapshot
	errors         chan<- error
}

// NewPollingWorker creates a PollingWorker with the dependencies required for
// periodic service snapshot collection.
func NewPollingWorker(
	unitRoots []string,
	systemSliceDir string,
	ticks <-chan struct{},
	output chan<- metrics.ServiceMetricsSnapshot,
	errors chan<- error,
) PollingWorker {
	return PollingWorker{
		unitRoots:      append([]string(nil), unitRoots...),
		systemSliceDir: systemSliceDir,
		ticks:          ticks,
		output:         output,
		errors:         errors,
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

// poll reads all service sources required for one snapshot cycle.
func (w PollingWorker) poll() (metrics.ServiceMetricsSnapshot, error) {
	units, err := collectUnitMetadata(w.unitRoots)
	if err != nil {
		return metrics.ServiceMetricsSnapshot{}, err
	}

	enabledStates, err := collectEnabledStates(w.unitRoots)
	if err != nil {
		return metrics.ServiceMetricsSnapshot{}, err
	}

	cgroups, err := collectServiceCGroupSnapshots(w.systemSliceDir)
	if err != nil {
		return metrics.ServiceMetricsSnapshot{}, err
	}

	return metrics.ServiceMetricsSnapshot{
		Timestamp: time.Now(),
		Services:  composeServiceSnapshots(units, enabledStates, cgroups),
	}, nil
}

// unitMetadata stores factual unit-file information used during snapshot composition.
type unitMetadata struct {
	Name            string
	Description     *string
	FragmentPath    *string
	Installable     bool
	RemainAfterExit bool
	Masked          bool
}

// cgroupSnapshot stores factual cgroup information for one active service unit.
type cgroupSnapshot struct {
	Name      string
	CGroup    *string
	Slice     *string
	PIDs      []uint32
	Resources *metrics.ServiceUnitResources
}

// cpuUsageUsecReader accumulates usage_usec while scanning one cpu.stat file.
type cpuUsageUsecReader struct {
	value uint64
	found bool
}

// collectUnitMetadata discovers top-level service unit files from the provided
// roots in precedence order.
func collectUnitMetadata(unitRoots []string) (map[string]unitMetadata, error) {
	units := make(map[string]unitMetadata)
	for _, root := range unitRoots {
		if err := collectUnitMetadataFromRoot(root, units); err != nil {
			return nil, err
		}
	}

	return units, nil
}

// collectUnitMetadataFromRoot reads one unit root into the cumulative metadata map.
func collectUnitMetadataFromRoot(root string, units map[string]unitMetadata) error {
	entries, err := readDirIfExists(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := recordUnitMetadataEntry(root, entry, units); err != nil {
			return err
		}
	}

	return nil
}

// readDirIfExists reads one directory and treats missing directories as empty input.
func readDirIfExists(path string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return entries, err
}

// recordUnitMetadataEntry reads one service unit entry into the metadata map.
func recordUnitMetadataEntry(root string, entry fs.DirEntry, units map[string]unitMetadata) error {
	if !isServiceUnitEntry(entry) {
		return nil
	}

	name := entry.Name()
	if _, exists := units[name]; exists {
		return nil
	}

	metadata, err := readUnitMetadata(filepath.Join(root, name), name)
	if err != nil {
		return err
	}
	units[name] = metadata
	return nil
}

// isServiceUnitEntry reports whether one directory entry is a top-level service unit file.
func isServiceUnitEntry(entry fs.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(entry.Name(), ".service")
}

// readUnitMetadata reads one unit file and extracts factual metadata.
func readUnitMetadata(path string, name string) (unitMetadata, error) {
	metadata, resolvedPath, err := loadUnitMetadataPaths(path, name)
	if err != nil {
		return unitMetadata{}, err
	}
	if metadata.Masked {
		return metadata, nil
	}

	currentSection := ""
	err = scanFileLines(resolvedPath, func(line string) error {
		currentSection = applyUnitMetadataLine(&metadata, currentSection, line)
		return nil
	})
	if err != nil {
		return unitMetadata{}, err
	}

	return metadata, nil
}

// loadUnitMetadataPaths resolves one unit path and initializes path-derived metadata.
func loadUnitMetadataPaths(path string, name string) (unitMetadata, string, error) {
	fragmentPath, resolvedPath, masked, err := resolveUnitPaths(path)
	if err != nil {
		return unitMetadata{}, "", err
	}

	metadata := unitMetadata{
		Name:         name,
		FragmentPath: stringPointer(fragmentPath),
		Masked:       masked,
	}
	return metadata, resolvedPath, nil
}

// resolveUnitPaths resolves symlink-backed unit paths and identifies masked units.
func resolveUnitPaths(path string) (fragmentPath string, resolvedPath string, masked bool, err error) {
	fragmentPath = path
	resolvedPath = path

	info, err := os.Lstat(path)
	if err != nil {
		return "", "", false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fragmentPath, resolvedPath, false, nil
	}

	resolvedPath, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", "", false, err
	}
	fragmentPath = resolvedPath
	return fragmentPath, resolvedPath, resolvedPath == "/dev/null", nil
}

// applyUnitMetadataLine folds one parsed unit-file line into the metadata accumulator.
func applyUnitMetadataLine(metadata *unitMetadata, currentSection string, rawLine string) string {
	line := strings.TrimSpace(rawLine)
	if isIgnoredUnitLine(line) {
		return currentSection
	}
	if nextSection, ok := parseUnitSection(line); ok {
		return nextSection
	}

	key, value, ok := splitINIKeyValue(line)
	if !ok {
		return currentSection
	}

	applyUnitMetadataField(metadata, currentSection, key, value)
	return currentSection
}

// isIgnoredUnitLine reports whether one unit-file line should be skipped.
func isIgnoredUnitLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";")
}

// parseUnitSection extracts one INI section header.
func parseUnitSection(line string) (string, bool) {
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return "", false
	}
	return strings.Trim(line, "[]"), true
}

// applyUnitMetadataField handles one parsed unit-file key-value pair.
func applyUnitMetadataField(metadata *unitMetadata, section string, key string, value string) {
	switch section {
	case "Unit":
		applyUnitSectionField(metadata, key, value)
	case "Install":
		metadata.Installable = true
	case "Service":
		applyServiceSectionField(metadata, key, value)
	}
}

// applyUnitSectionField handles one [Unit] section key-value pair.
func applyUnitSectionField(metadata *unitMetadata, key string, value string) {
	if key == "Description" && metadata.Description == nil {
		metadata.Description = stringPointer(value)
	}
}

// applyServiceSectionField handles one [Service] section key-value pair.
func applyServiceSectionField(metadata *unitMetadata, key string, value string) {
	if key == "RemainAfterExit" {
		metadata.RemainAfterExit = strings.EqualFold(value, "yes") || strings.EqualFold(value, "true")
	}
}

// collectEnabledStates resolves unit-file enablement state from systemd-style
// wants and requires symlink directories.
func collectEnabledStates(unitRoots []string) (map[string]string, error) {
	states := make(map[string]string)
	for _, root := range unitRoots {
		if err := collectEnabledStatesFromRoot(root, states); err != nil {
			return nil, err
		}
	}

	return states, nil
}

// collectEnabledStatesFromRoot walks one unit root and merges enablement states.
func collectEnabledStatesFromRoot(root string, states map[string]string) error {
	return walkDirIfExists(root, func(path string, d fs.DirEntry) error {
		return collectEnabledStateEntry(root, path, d, states)
	})
}

// walkDirIfExists walks one directory tree and treats a missing root as empty input.
func walkDirIfExists(root string, visit func(path string, d fs.DirEntry) error) error {
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		return visit(path, d)
	})
	if errors.Is(walkErr, fs.ErrNotExist) {
		return nil
	}
	return walkErr
}

// collectEnabledStateEntry folds one walked unit-root entry into the enablement map.
func collectEnabledStateEntry(root string, path string, d fs.DirEntry, states map[string]string) error {
	if !isEnabledStateCandidate(d) {
		return nil
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	if isTopLevelRelativePath(rel) {
		return collectTopLevelEnabledState(path, d.Name(), states)
	}

	markEnabledState(rel, d.Name(), states)
	return nil
}

// isEnabledStateCandidate reports whether one walked entry may affect unit enablement state.
func isEnabledStateCandidate(d fs.DirEntry) bool {
	return !d.IsDir() && strings.HasSuffix(d.Name(), ".service")
}

// isTopLevelRelativePath reports whether one relative path stays at the unit-root top level.
func isTopLevelRelativePath(rel string) bool {
	return !strings.ContainsRune(rel, filepath.Separator)
}

// markEnabledState records enabled service state when the path passes through an enablement directory.
func markEnabledState(rel string, unitName string, states map[string]string) {
	if containsEnablementDirectory(rel) && states[unitName] != "masked" {
		states[unitName] = "enabled"
	}
}

// collectTopLevelEnabledState records masked top-level units before deeper enablement paths.
func collectTopLevelEnabledState(path string, unitName string, states map[string]string) error {
	masked, err := isMaskedUnitPath(path)
	if err != nil {
		return err
	}
	if masked {
		states[unitName] = "masked"
	}
	return nil
}

// isMaskedUnitPath reports whether one top-level unit path resolves to /dev/null.
func isMaskedUnitPath(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false, err
	}
	return resolvedPath == "/dev/null", nil
}

// containsEnablementDirectory reports whether a relative unit path passes
// through one of systemd's enablement directories.
func containsEnablementDirectory(rel string) bool {
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if strings.HasSuffix(part, ".wants") || strings.HasSuffix(part, ".requires") || strings.HasSuffix(part, ".upholds") {
			return true
		}
	}

	return false
}

// collectServiceCGroupSnapshots discovers service unit cgroups under the system slice.
func collectServiceCGroupSnapshots(systemSliceDir string) (map[string]cgroupSnapshot, error) {
	cgroups := make(map[string]cgroupSnapshot)
	err := walkDirIfExists(systemSliceDir, func(path string, d fs.DirEntry) error {
		return collectServiceCGroupSnapshotEntry(systemSliceDir, path, d, cgroups)
	})
	if err != nil {
		return nil, err
	}

	return cgroups, nil
}

// collectServiceCGroupSnapshotEntry records one service cgroup discovered by the walk.
func collectServiceCGroupSnapshotEntry(
	systemSliceDir string,
	path string,
	d fs.DirEntry,
	cgroups map[string]cgroupSnapshot,
) error {
	if !d.IsDir() || !strings.HasSuffix(d.Name(), ".service") {
		return nil
	}

	snapshot, err := readCGroupSnapshot(systemSliceDir, path)
	if err != nil {
		return err
	}
	cgroups[snapshot.Name] = snapshot
	return nil
}

// readCGroupSnapshot reads one service cgroup directory into snapshot metadata.
func readCGroupSnapshot(systemSliceDir string, path string) (cgroupSnapshot, error) {
	parentRoot := filepath.Dir(systemSliceDir)
	rel, err := filepath.Rel(parentRoot, path)
	if err != nil {
		return cgroupSnapshot{}, err
	}

	pids, err := readPIDList(filepath.Join(path, "cgroup.procs"))
	if err != nil {
		return cgroupSnapshot{}, err
	}

	resources, _, err := readUnitResources(path)
	if err != nil {
		return cgroupSnapshot{}, err
	}

	cgroupPath := "/" + filepath.ToSlash(rel)
	return cgroupSnapshot{
		Name:      filepath.Base(path),
		CGroup:    stringPointer(cgroupPath),
		Slice:     resolveOwningSlice(rel),
		PIDs:      pids,
		Resources: resources,
	}, nil
}

// readPIDList reads one cgroup.procs file into a deduplicated PID slice.
func readPIDList(path string) ([]uint32, error) {
	seen := make(map[uint32]struct{})
	pids := make([]uint32, 0)

	err := scanFileLines(path, func(line string) error {
		return parsePIDLine(line, seen, &pids)
	})
	if err != nil {
		return nil, err
	}

	return pids, nil
}

// parsePIDLine parses and deduplicates one cgroup.procs line.
func parsePIDLine(line string, seen map[uint32]struct{}, pids *[]uint32) error {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}

	pid, err := strconv.ParseUint(trimmed, 10, 32)
	if err != nil {
		return err
	}
	pid32 := uint32(pid)
	if _, exists := seen[pid32]; exists {
		return nil
	}

	seen[pid32] = struct{}{}
	*pids = append(*pids, pid32)
	return nil
}

// readUnitResources reads cgroup-scoped memory and CPU values when available.
func readUnitResources(path string) (*metrics.ServiceUnitResources, bool, error) {
	memoryBytes, hasMemory, err := readOptionalUint64File(filepath.Join(path, "memory.current"))
	if err != nil {
		return nil, false, err
	}

	cpuUsageUsec, hasCPUUsage, err := readOptionalCPUUsageUsec(filepath.Join(path, "cpu.stat"))
	if err != nil {
		return nil, false, err
	}

	if !hasMemory && !hasCPUUsage {
		return nil, false, nil
	}

	return &metrics.ServiceUnitResources{
		MemoryBytes:  optionalUint64Pointer(memoryBytes, hasMemory),
		CPUUsageUsec: optionalUint64Pointer(cpuUsageUsec, hasCPUUsage),
	}, true, nil
}

// readOptionalUint64File loads one optional uint64 file and tolerates missing paths.
func readOptionalUint64File(path string) (uint64, bool, error) {
	value, err := readUint64File(path)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return value, true, nil
}

// readOptionalCPUUsageUsec loads usage_usec from an optional cpu.stat file.
func readOptionalCPUUsageUsec(path string) (uint64, bool, error) {
	value, err := readCPUUsageUsec(path)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return value, true, nil
}

// optionalUint64Pointer converts one optional uint64 value into a pointer when present.
func optionalUint64Pointer(value uint64, ok bool) *uint64 {
	if !ok {
		return nil
	}
	return &value
}

// readUint64File reads one integer value from a single-line file.
func readUint64File(path string) (uint64, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is configured by the application
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}

// readCPUUsageUsec parses usage_usec from one cgroup cpu.stat file.
func readCPUUsageUsec(path string) (uint64, error) {
	reader := cpuUsageUsecReader{}
	err := scanFileLines(path, reader.consume)
	if err != nil {
		return 0, err
	}
	if !reader.found {
		return 0, fs.ErrNotExist
	}

	return reader.value, nil
}

// consume folds one cpu.stat line into the usage_usec reader.
func (r *cpuUsageUsecReader) consume(line string) error {
	key, value, ok := splitWhitespacePair(line)
	if !ok || key != "usage_usec" {
		return nil
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return err
	}
	if !r.found {
		r.value = parsed
		r.found = true
	}
	return nil
}

// scanFileLines streams one text file line-by-line into the provided handler.
func scanFileLines(path string, handle func(string) error) (err error) {
	file, err := os.Open(path) // #nosec G304 -- path is configured by the application
	if err != nil {
		return err
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if handleErr := handle(scanner.Text()); handleErr != nil {
			return handleErr
		}
	}

	return scanner.Err()
}

// resolveOwningSlice resolves the nearest parent slice from one relative cgroup path.
func resolveOwningSlice(rel string) *string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for i := len(parts) - 2; i >= 0; i-- {
		if strings.HasSuffix(parts[i], ".slice") {
			return stringPointer(parts[i])
		}
	}
	return nil
}

// composeServiceSnapshots normalizes discovered unit and cgroup data into one
// sorted snapshot collection.
func composeServiceSnapshots(
	units map[string]unitMetadata,
	enabledStates map[string]string,
	cgroups map[string]cgroupSnapshot,
) []metrics.ServiceUnitSnapshot {
	names := buildSortedUnitNames(units, cgroups)
	services := make([]metrics.ServiceUnitSnapshot, 0, len(names))
	for _, name := range names {
		unit := units[name]
		cgroup, hasCGroup := cgroups[name]
		services = append(services, buildServiceSnapshot(name, unit, enabledStates[name], cgroup, hasCGroup))
	}

	return services
}

// buildSortedUnitNames builds a sorted union of unit and cgroup names.
func buildSortedUnitNames(
	units map[string]unitMetadata,
	cgroups map[string]cgroupSnapshot,
) []string {
	seen := make(map[string]struct{}, len(units)+len(cgroups))
	for name := range units {
		seen[name] = struct{}{}
	}
	for name := range cgroups {
		seen[name] = struct{}{}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// buildServiceSnapshot builds one normalized service snapshot entry.
func buildServiceSnapshot(
	name string,
	unit unitMetadata,
	enabledState string,
	cgroup cgroupSnapshot,
	hasCGroup bool,
) metrics.ServiceUnitSnapshot {
	loadState := stringPointer("loaded")
	if unit.Masked {
		loadState = stringPointer("masked")
	}

	activeState, subState := resolveRuntimeStates(unit, cgroup, hasCGroup)
	resolvedEnabledState := resolveEnabledState(unit, enabledState)
	mainPID := firstPID(cgroup.PIDs)

	return metrics.ServiceUnitSnapshot{
		Name:         name,
		Manager:      serviceManagerName,
		UnitType:     serviceUnitType,
		Description:  unit.Description,
		LoadState:    loadState,
		ActiveState:  activeState,
		SubState:     subState,
		EnabledState: resolvedEnabledState,
		MainPID:      mainPID,
		PIDs:         append([]uint32(nil), cgroup.PIDs...),
		CGroup:       cgroup.CGroup,
		Slice:        cgroup.Slice,
		FragmentPath: unit.FragmentPath,
		Resources:    cgroup.Resources,
	}
}

// resolveRuntimeStates resolves the coarse and detailed runtime state strings.
func resolveRuntimeStates(unit unitMetadata, cgroup cgroupSnapshot, hasCGroup bool) (*string, *string) {
	if hasCGroup && len(cgroup.PIDs) > 0 {
		return stringPointer("active"), stringPointer("running")
	}
	if unit.RemainAfterExit {
		return stringPointer("active"), stringPointer("exited")
	}
	return stringPointer("inactive"), stringPointer("dead")
}

// resolveEnabledState resolves one service enablement string.
func resolveEnabledState(unit unitMetadata, explicitState string) *string {
	if explicitState != "" {
		return stringPointer(explicitState)
	}
	if unit.Masked {
		return stringPointer("masked")
	}
	if unit.Installable {
		return stringPointer("disabled")
	}
	if unit.Name == "" && unit.FragmentPath == nil {
		return nil
	}
	return stringPointer("static")
}

// firstPID returns the first PID from one PID slice when available.
func firstPID(pids []uint32) *uint32 {
	if len(pids) == 0 {
		return nil
	}
	return &pids[0]
}

// splitINIKeyValue splits one INI-style key-value line.
func splitINIKeyValue(line string) (string, string, bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

// splitWhitespacePair splits one whitespace-delimited key-value line.
func splitWhitespacePair(line string) (string, string, bool) {
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return "", "", false
	}
	return fields[0], fields[1], true
}

// stringPointer allocates a string pointer for one value.
func stringPointer(value string) *string {
	return &value
}
