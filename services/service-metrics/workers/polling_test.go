package workers

import (
	"path/filepath"
	"testing"

	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/fstest"
)

// Requirements: service-metrics-svc/FR-002, service-metrics-svc/FR-003, service-metrics-svc/FR-005, service-metrics-svc/FR-006
func TestPollingWorkerPollBuildsServiceSnapshotFromUnitsAndCgroups(t *testing.T) {
	t.Parallel()

	snapshot := mustPollFixtureSnapshot(t, []string{"etc", "usr"})
	assertServiceSnapshotCount(t, snapshot, 4)
	assertAlphaSnapshot(t, findServiceSnapshot(t, snapshot.Services, "alpha.service"))
	assertBetaSnapshot(t, findServiceSnapshot(t, snapshot.Services, "beta.service"))
	assertGammaSnapshot(t, findServiceSnapshot(t, snapshot.Services, "gamma.service"))
}

// Requirements: service-metrics-svc/FR-003, service-metrics-svc/FR-007
func TestPollingWorkerPollTreatsRemainAfterExitUnitAsActiveExited(t *testing.T) {
	t.Parallel()

	snapshot := mustPollFixtureSnapshot(t, []string{"etc"})
	delta := findServiceSnapshot(t, snapshot.Services, "delta.service")
	assertStringPointer(t, delta.ActiveState, "active")
	assertStringPointer(t, delta.SubState, "exited")
}

// mustPollFixtureSnapshot polls one standard fixture using the requested unit roots.
func mustPollFixtureSnapshot(t *testing.T, roots []string) metrics.ServiceMetricsSnapshot {
	t.Helper()

	fixture := newPollingFixture(t)
	worker := NewPollingWorker(resolveFixtureUnitRoots(fixture, roots), fixture.systemSlice, nil, nil, nil)

	snapshot, err := worker.poll()
	if err != nil {
		t.Fatalf("poll() error = %v", err)
	}

	return snapshot
}

// resolveFixtureUnitRoots maps fixture root selectors to absolute fixture paths.
func resolveFixtureUnitRoots(fixture pollingFixture, roots []string) []string {
	resolved := make([]string, 0, len(roots))
	for _, root := range roots {
		switch root {
		case "etc":
			resolved = append(resolved, fixture.etcSystemd)
		case "usr":
			resolved = append(resolved, fixture.usrSystemd)
		}
	}
	return resolved
}

// assertServiceSnapshotCount verifies the expected number of services in one snapshot.
func assertServiceSnapshotCount(t *testing.T, snapshot metrics.ServiceMetricsSnapshot, want int) {
	t.Helper()
	if len(snapshot.Services) != want {
		t.Fatalf("len(Services) = %d, want %d", len(snapshot.Services), want)
	}
}

// assertAlphaSnapshot verifies the active service fixture with runtime and resource data.
func assertAlphaSnapshot(t *testing.T, service metrics.ServiceUnitSnapshot) {
	t.Helper()

	assertStringPointer(t, service.ActiveState, "active")
	assertStringPointer(t, service.SubState, "running")
	assertStringPointer(t, service.EnabledState, "enabled")
	assertStringPointer(t, service.CGroup, "/system.slice/alpha.service")
	assertMainPID(t, service, 101)
	assertMemoryBytes(t, service, 4096)
	assertCPUUsageUsec(t, service, 777)
}

// assertBetaSnapshot verifies the static inactive service fixture.
func assertBetaSnapshot(t *testing.T, service metrics.ServiceUnitSnapshot) {
	t.Helper()
	assertStringPointer(t, service.EnabledState, "static")
	assertStringPointer(t, service.ActiveState, "inactive")
}

// assertGammaSnapshot verifies the masked service fixture.
func assertGammaSnapshot(t *testing.T, service metrics.ServiceUnitSnapshot) {
	t.Helper()
	assertStringPointer(t, service.EnabledState, "masked")
	assertStringPointer(t, service.LoadState, "masked")
}

// assertMainPID verifies the primary PID attached to one service snapshot.
func assertMainPID(t *testing.T, service metrics.ServiceUnitSnapshot, want uint32) {
	t.Helper()
	if service.MainPID == nil || *service.MainPID != want {
		t.Fatalf("MainPID = %v, want %d", service.MainPID, want)
	}
}

// assertMemoryBytes verifies the collected memory usage for one service snapshot.
func assertMemoryBytes(t *testing.T, service metrics.ServiceUnitSnapshot, want uint64) {
	t.Helper()
	if service.Resources == nil || service.Resources.MemoryBytes == nil || *service.Resources.MemoryBytes != want {
		t.Fatalf("memory bytes = %#v, want %d", service.Resources, want)
	}
}

// assertCPUUsageUsec verifies the collected CPU usage for one service snapshot.
func assertCPUUsageUsec(t *testing.T, service metrics.ServiceUnitSnapshot, want uint64) {
	t.Helper()
	if service.Resources == nil || service.Resources.CPUUsageUsec == nil || *service.Resources.CPUUsageUsec != want {
		t.Fatalf("cpu usage = %#v, want %d", service.Resources, want)
	}
}

// pollingFixture groups filesystem paths used by polling tests.
type pollingFixture struct {
	etcSystemd  string
	usrSystemd  string
	systemSlice string
}

// newPollingFixture creates one filesystem fixture for polling tests.
func newPollingFixture(t *testing.T) pollingFixture {
	t.Helper()

	root := t.TempDir()
	etcSystemd := filepath.Join(root, "etc", "systemd", "system")
	usrSystemd := filepath.Join(root, "usr", "lib", "systemd", "system")
	systemSlice := filepath.Join(root, "sys", "fs", "cgroup", "system.slice")

	fstest.MustMkdirAll(t, filepath.Join(etcSystemd, "multi-user.target.wants"), 0o755)
	fstest.MustMkdirAll(t, usrSystemd, 0o755)
	fstest.MustMkdirAll(t, filepath.Join(systemSlice, "alpha.service"), 0o755)

	fstest.MustWriteFile(t, filepath.Join(etcSystemd, "alpha.service"), "[Unit]\nDescription=Alpha Service\n\n[Install]\nWantedBy=multi-user.target\n", 0o755, 0o644)
	fstest.MustWriteFile(t, filepath.Join(usrSystemd, "beta.service"), "[Unit]\nDescription=Beta Service\n", 0o755, 0o644)
	fstest.MustWriteFile(t, filepath.Join(etcSystemd, "delta.service"), "[Unit]\nDescription=Delta Service\n\n[Service]\nRemainAfterExit=yes\n\n[Install]\nWantedBy=multi-user.target\n", 0o755, 0o644)
	fstest.MustSymlink(t, "../alpha.service", filepath.Join(etcSystemd, "multi-user.target.wants", "alpha.service"))
	fstest.MustSymlink(t, "/dev/null", filepath.Join(etcSystemd, "gamma.service"))

	fstest.MustWriteFile(t, filepath.Join(systemSlice, "alpha.service", "cgroup.procs"), "101\n102\n", 0o755, 0o644)
	fstest.MustWriteFile(t, filepath.Join(systemSlice, "alpha.service", "memory.current"), "4096\n", 0o755, 0o644)
	fstest.MustWriteFile(t, filepath.Join(systemSlice, "alpha.service", "cpu.stat"), "usage_usec 777\nuser_usec 700\nsystem_usec 77\n", 0o755, 0o644)

	return pollingFixture{
		etcSystemd:  etcSystemd,
		usrSystemd:  usrSystemd,
		systemSlice: systemSlice,
	}
}

// findServiceSnapshot returns one named service snapshot from a collection.
func findServiceSnapshot(t *testing.T, services []metrics.ServiceUnitSnapshot, name string) metrics.ServiceUnitSnapshot {
	t.Helper()
	for _, service := range services {
		if service.Name == name {
			return service
		}
	}
	t.Fatalf("service %q not found", name)
	return metrics.ServiceUnitSnapshot{}
}

// assertStringPointer verifies one string pointer field value.
func assertStringPointer(t *testing.T, value *string, want string) {
	t.Helper()
	if value == nil || *value != want {
		t.Fatalf("string pointer = %v, want %q", value, want)
	}
}
