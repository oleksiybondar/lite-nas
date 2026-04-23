package metrics

import "time"

// CPUCoreRawSample contains raw CPU counters for a single CPU context.
//
// It is used both for the aggregated system-wide CPU line and for
// individual per-core CPU lines read from the host metrics source.
//
// Total and Idle are cumulative counters, not percentages.
type CPUCoreRawSample struct {
	// Total is the cumulative total CPU time.
	Total uint64

	// Idle is the cumulative idle CPU time.
	Idle uint64
}

// CPURawSample contains raw CPU counters for the whole system and all cores.
//
// CPU usage percentages are derived by comparing two CPURawSample values
// collected at different times.
type CPURawSample struct {
	// Total contains the aggregated CPU counters for the whole system.
	Total CPUCoreRawSample

	// Cores contains per-core CPU counters in core index order.
	Cores []CPUCoreRawSample
}

// CPUSample contains computed CPU usage percentages.
//
// Values in this struct are derived from two raw CPU samples and represent
// usage at a specific collection interval.
type CPUSample struct {
	// TotalUsagePct is the computed total CPU usage percentage for the system.
	TotalUsagePct float64

	// PerCoreUsage contains computed CPU usage percentages for each core
	// in the core index order.
	PerCoreUsage []float64
}

// MemSample contains computed memory usage values for the system.
//
// All values are point-in-time values for a single collection cycle.
type MemSample struct {
	// TotalBytes is the total system memory in bytes.
	TotalBytes uint64

	// UsedBytes is the used system memory in bytes.
	UsedBytes uint64

	// UsedPct is the used system memory as a percentage of total memory.
	UsedPct float64
}

// RawSystemSnapshot represents one polling cycle result before CPU usage is computed.
//
// It contains raw CPU counters and computed memory values collected at a
// specific point in time. CPU usage percentages are not yet calculated and
// require comparison with a previous RawSystemSnapshot.
//
// This structure is used as the output of the polling service and serves as
// the input for the CPU usage calculator stage in the processing pipeline.
type RawSystemSnapshot struct {
	// Timestamp is the time at which the raw snapshot was collected.
	Timestamp time.Time

	// CPU contains raw cumulative CPU counters.
	CPU CPURawSample

	// Mem contains computed memory usage values for the same polling cycle.
	Mem MemSample
}

// SystemSnapshot represents one complete system metrics reading.
//
// It is the main data structure used for the latest snapshot, history
// entries, event payloads, and request/reply responses.
type SystemSnapshot struct {
	// Timestamp is the time at which the snapshot was collected.
	Timestamp time.Time

	// CPU contains computed CPU usage values.
	CPU CPUSample

	// Mem contains computed memory usage values.
	Mem MemSample
}
