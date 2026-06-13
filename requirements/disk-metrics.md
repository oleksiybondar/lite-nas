# Disk Metrics — Requirements

## Overview

The Disk Metrics service is responsible for collecting factual disk, mount, and
filesystem-observation data from the host system, composing one atomic snapshot
per polling cycle, maintaining a short in-memory history, and providing access
to current and historical snapshots.

The service follows the same snapshot-first model as the Network Metrics
service. Its snapshot contract separates block devices, observed filesystem
types, and active mounts into distinct sections so consumers can reason about
those domains independently.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Poll disk metrics and compose one atomic snapshot

#### FR-001 Description

The service MUST poll host disk-related metrics on a configurable interval and
compose one atomic disk snapshot for each successful polling cycle.

#### FR-001 Input

- Configured poll interval
- Host disk metric sources

#### FR-001 Output

- One `DiskMetricsSnapshot` value per successful polling cycle

#### FR-001 Acceptance Criteria

- Polling runs continuously at the configured interval
- The default poll interval is 1 second
- Each successful cycle produces exactly one snapshot
- All snapshot sections in one snapshot represent the same polling cycle

---

### FR-002 Collect block-device-like entries as one flat device array

#### FR-002 Description

The service MUST collect visible block-device-like host entries into one flat
`devices` array rather than exposing separate top-level collections per device
family.

#### FR-002 Input

- `/sys/block/*`
- Related sysfs metadata files for discovered block devices

#### FR-002 Output

- Snapshot `devices` array containing one entry per discovered disk-like or
  partition-like device

#### FR-002 Acceptance Criteria

- The snapshot uses one DTO type for all block-device-like entries
- The `devices` collection is a flat array rather than a map keyed by device
  name
- Device identity uses runtime kernel node names such as `sda`, `sda1`, and
  `nvme0n1`
- Device entries may represent disks, partitions, loop devices, device-mapper
  devices, MD devices, ZFS zvols, or unknown device kinds
- The service does not assume that runtime kernel node names are stable across
  reboots

---

### FR-003 Differentiate device entries by kind and parent relationship

#### FR-003 Description

The service MUST distinguish device entry roles with explicit kind and
parent/child relationship fields rather than separate DTO families.

#### FR-003 Input

- Discovered device topology from sysfs

#### FR-003 Output

- Device entries with explicit topology metadata

#### FR-003 Acceptance Criteria

- Each device entry includes `node`
- Each device entry includes `kind`
- Allowed `kind` values are `disk`, `partition`, `loop`, `dm`, `md`, `zvol`,
  and `unknown`
- Partition or derived-device entries may include `parent`
- Devices that own discovered partitions may include `partitions`
- Optional child and parent fields are omitted or null when not applicable

---

### FR-004 Collect factual device metadata without inventing values

#### FR-004 Description

The service MUST collect factual device metadata where available and MUST omit
or leave null fields that cannot be read from the host.

#### FR-004 Input

- `/sys/block/*/size`
- `/sys/block/*/queue/logical_block_size`
- `/sys/block/*/queue/rotational`
- `/sys/block/*/removable`
- `/sys/block/*/device/vendor`
- `/sys/block/*/device/model`
- `/sys/block/*/device/wwid`
- Related stable-ID sysfs files where available

#### FR-004 Output

- Device entries containing factual host metadata

#### FR-004 Acceptance Criteria

- Device entries include `major`, `minor`, and `size_bytes`
- Device entries may include `name`, `vendor`, `model`, `description`,
  `serial`, and `wwn`
- Device entries include `connection_type`
- Allowed `connection_type` values are `UNKNOWN`, `SATA`, `SAS`, `SCSI`,
  `NVME`, `USB`, `VIRTIO`, `IDE`, `MMC`, `LOOP`, `DM`, and `MD`
- Device entries may include `rotational` and `removable`
- Missing sysfs metadata files do not fail snapshot composition
- The service prefers stable identifiers such as serial or WWN when available
  but does not invent them when unavailable

---

### FR-005 Collect disk I/O counters from kernel disk statistics

#### FR-005 Description

The service MUST collect disk I/O counters from Linux kernel disk statistics
where available and attach them to the corresponding device entries.

#### FR-005 Input

- `/proc/diskstats`

#### FR-005 Output

- Optional device `io` section for devices represented in the snapshot

#### FR-005 Acceptance Criteria

- The `io` section may include `reads_completed`, `reads_merged`,
  `sectors_read`, `read_time_ms`, `writes_completed`, `writes_merged`,
  `sectors_written`, `write_time_ms`, `io_in_progress`, `io_time_ms`,
  `weighted_io_time_ms`, `discards_completed`, `discards_merged`,
  `sectors_discarded`, `discard_time_ms`, `flushes_completed`, and
  `flush_time_ms`
- The parser accepts diskstats lines that omit discard and flush fields on
  older or more limited kernels
- Missing optional diskstats fields do not fail the whole snapshot
- I/O counters remain factual/raw and are not converted into derived
  performance judgments

---

### FR-006 Compose observed filesystem summaries from active mounts

#### FR-006 Description

The service MUST derive the `filesystems` snapshot section from filesystem
types currently observed in active mounts rather than from installed kernel
filesystem drivers.

#### FR-006 Input

- Active mount entries collected for the current snapshot

#### FR-006 Output

- Snapshot `filesystems` object keyed by observed filesystem type name

#### FR-006 Acceptance Criteria

- `filesystems` is an object keyed by filesystem type such as `ext4` or `tmpfs`
- Each filesystem entry includes `type`
- Each filesystem entry includes `mount_count`
- The section summarizes currently observed mounts rather than installed or
  loadable filesystem drivers
- Filesystem categories use:
  - `local_block` for `ext4`, `xfs`, `btrfs`, `vfat`, `exfat`, `ntfs`, and
    `zfs`
  - `network` for `cifs`, `smb3`, `nfs`, and `nfs4`
  - `memory` for `tmpfs` and `ramfs`
  - `pseudo` for `proc`, `sysfs`, `devtmpfs`, `devpts`, `cgroup`, `cgroup2`,
    `securityfs`, `debugfs`, `tracefs`, `configfs`, `fusectl`, `mqueue`,
    `hugetlbfs`, `pstore`, `efivarfs`, and `bpf`
  - `fuse` for `fuse` and `fuse.*`
  - `unknown` as fallback for unclassified filesystem types

---

### FR-007 Collect all visible mounts as a separate snapshot section

#### FR-007 Description

The service MUST collect all currently visible mounts into a separate `mounts`
section regardless of whether the mount has a local block-device backing.

#### FR-007 Input

- `/proc/self/mountinfo`
- `/proc/mounts` as fallback when required
- Host filesystem usage data from `statfs` or `statvfs`

#### FR-007 Output

- Snapshot `mounts` object keyed by mountpoint

#### FR-007 Acceptance Criteria

- The snapshot includes all mounts visible to the process at collection time
- CIFS/NFS mounts appear even without a local block device
- `tmpfs` and `ramfs` mounts appear even without a local block device
- Mounted USB-backed filesystems appear once mounted even when the device
  existed earlier without a mount
- Each mount entry includes `mountpoint`, `source`, `filesystem`, `readonly`,
  `remote`, and `memory_backed`
- Each mount entry includes `total_bytes`, `used_bytes`, `free_bytes`,
  `available_bytes`, and `used_percent` when usage data can be read
- `device` resolves to the local kernel node when the source maps to
  `/dev/<node>` and is otherwise null or omitted
- `remote` is true for `cifs`, `smb3`, `nfs`, and `nfs4`
- `memory_backed` is true for `tmpfs` and `ramfs`

---

### FR-008 Keep devices, filesystems, and mounts as separate concepts

#### FR-008 Description

The service MUST keep device inventory, observed filesystem summaries, and
mount inventory as separate snapshot concepts.

#### FR-008 Input

- Collected device data
- Collected mount data
- Derived filesystem summaries

#### FR-008 Output

- Normalized snapshot contract suitable for downstream consumers

#### FR-008 Acceptance Criteria

- A device may exist in the snapshot without any mount
- A mount may exist in the snapshot without any local block device
- Filesystem summaries are derived from observed mounts rather than devices
- Device entries may include `mounts` references when the relationship is
  observable
- Device entries may include `zpool_memberships`, but the field remains empty
  unless ZFS membership data is available

---

### FR-009 Keep the snapshot contract snapshot-first and collector-free

#### FR-009 Description

The service MUST expose a snapshot-first contract and MUST NOT add collector
metadata to the root snapshot payload.

#### FR-009 Input

- Composed disk metrics from FR-002 through FR-008

#### FR-009 Output

- Stable disk snapshot contract

#### FR-009 Acceptance Criteria

- The snapshot root shape is:
  - `timestamp`
  - `devices`
  - `filesystems`
  - `mounts`
- The snapshot does not include a `collector` field
- Repeated device entities are represented by an array
- Filesystem summaries and mounts are represented by objects keyed by their
  natural names

---

### FR-010 Maintain bounded snapshot history

#### FR-010 Description

The service MUST maintain a bounded in-memory history of collected disk
snapshots.

#### FR-010 Input

- Stream of successfully collected snapshots from FR-001

#### FR-010 Output

- Chronological retained snapshot history

#### FR-010 Acceptance Criteria

- The default retained history contains at most 120 snapshots
- Older snapshots are discarded automatically when the bound is reached
- History retrieval returns snapshots in chronological order
- Runtime memory usage for history remains bounded by configuration

---

### FR-011 Provide the latest disk snapshot

#### FR-011 Description

The service MUST provide access to the latest successfully collected disk
snapshot.

#### FR-011 Input

- Latest snapshot retained from FR-001

#### FR-011 Output

- Single latest disk snapshot

#### FR-011 Acceptance Criteria

- The latest snapshot always reflects the most recently completed successful
  polling cycle
- The latest snapshot can be returned independently of history retrieval
- When no snapshot has been collected yet, the service can report that no
  snapshot is available

---

### FR-012 Provide retained disk snapshot history

#### FR-012 Description

The service MUST provide access to retained disk snapshot history within the
configured retention bound.

#### FR-012 Input

- Request for snapshot history

#### FR-012 Output

- Chronological collection of retained disk snapshots

#### FR-012 Acceptance Criteria

- Returned history does not exceed the retained in-memory bound
- Empty history results are allowed when no snapshots have been collected
- History output remains JSON-serializable

---

### FR-013 Publish snapshot update events

#### FR-013 Description

The service MUST publish a snapshot update event for each successfully composed
disk snapshot.

#### FR-013 Input

- Successfully composed snapshot from FR-001

#### FR-013 Output

- Snapshot update event containing the current disk snapshot

#### FR-013 Acceptance Criteria

- One event is published for each successful polling cycle
- Event payload contains the composed disk snapshot
- Event production frequency follows the successful polling cadence
- Event publication does not depend on change detection between snapshots

---

## Interface Requirements

### IR-001 Expose messaging contracts for latest snapshot and history

#### IR-001 Description

The service MUST expose messaging contracts for retrieving the latest disk
snapshot and retained disk snapshot history.

#### IR-001 Input

- Messaging requests for latest snapshot
- Messaging requests for retained history

#### IR-001 Output

- Request/reply responses containing latest snapshot or retained history

#### IR-001 Acceptance Criteria

- Latest snapshot access is exposed through a dedicated request/reply contract
- History access is exposed through a dedicated request/reply contract
- When no latest snapshot exists, the latest-snapshot response reports that no
  snapshot is available rather than returning invented data
- Response payloads conform to the shared disk-metrics contract package

---

### IR-002 Publish shared snapshot update events

#### IR-002 Description

The service MUST publish disk snapshot update events through the shared
disk-metrics event contract.

#### IR-002 Input

- Snapshot update events from FR-013

#### IR-002 Output

- Messaging events containing the current disk snapshot

#### IR-002 Acceptance Criteria

- Published event payloads conform to the shared disk-metrics contract package
- Event payloads are JSON-serializable
- Consumers can treat each event as one complete disk snapshot

---

### IR-003 Use stable field naming and project timestamp formatting

#### IR-003 Description

The service MUST expose a stable disk snapshot JSON contract with explicit
field naming and project-standard timestamp formatting.

#### IR-003 Input

- Composed disk snapshot from FR-009

#### IR-003 Output

- Stable JSON-serializable snapshot payload

#### IR-003 Acceptance Criteria

- The snapshot uses the field name `timestamp` at the root
- The snapshot timestamp format matches project timestamps such as
  `2026-06-09T10:34:28.461988712+02:00`
- The snapshot uses the section names `devices`, `filesystems`, and `mounts`
- Optional fields use null or omission rather than fake placeholder values

---

## Reliability Requirements

### RR-001 Continue polling after collection failures

#### RR-001 Description

The service MUST continue future polling cycles when one polling cycle fails.

#### RR-001 Acceptance Criteria

- Failure to collect one or more source files for one cycle does not terminate
  the service process
- Polling failures are logged with enough detail for diagnosis
- A later successful cycle still produces a snapshot and publication event

---

### RR-002 Tolerate partial host metadata loss within one successful snapshot

#### RR-002 Description

The service MUST tolerate missing optional host metadata within a successful
snapshot rather than failing the entire snapshot composition.

#### RR-002 Acceptance Criteria

- Missing sysfs files for optional metadata do not fail snapshot composition
- Missing vendor, model, serial, WWN, rotational, or removable data do not
  prevent the device from appearing in the snapshot
- Missing usage data for one mount does not require dropping unrelated mounts
  from the snapshot
- Older kernel diskstats rows that omit discard or flush counters remain
  parseable

---

### RR-003 Isolate publication and request handling from snapshot retention

#### RR-003 Description

The service MUST keep latest-snapshot retention and history retention resilient
to downstream publication failures.

#### RR-003 Acceptance Criteria

- A snapshot that is successfully composed is still retained even if event
  publication fails
- Failure to publish one snapshot event does not block future polling cycles
- Request/reply access to retained snapshots does not depend on successful
  publication of prior events

---

## Operational Requirements

### OR-001 Support configurable polling and history bounds

#### OR-001 Description

The service MUST provide configuration for polling cadence and retained history
capacity.

#### OR-001 Acceptance Criteria

- Poll interval is configurable
- Retained history capacity is configurable
- Default configuration supports 1-second polling and 120 retained snapshots

---

### OR-002 Prefer Linux procfs/sysfs mount sources defined by the contract

#### OR-002 Description

The service MUST collect disk snapshot data from the Linux host sources defined
by the contract unless a documented fallback is required.

#### OR-002 Acceptance Criteria

- Disk I/O counters are sourced from `/proc/diskstats`
- Device inventory and metadata are sourced from `/sys/block/*` and related
  sysfs files
- Mount inventory is sourced from `/proc/self/mountinfo` when available
- `/proc/mounts` is used only as a fallback when required for mount discovery
- Filesystem usage values are sourced from `statfs` or `statvfs` where possible

---

### OR-003 Operate with bounded in-memory retention

#### OR-003 Description

The service MUST operate with bounded in-memory retention rather than
unbounded accumulation of historical disk snapshots.

#### OR-003 Acceptance Criteria

- Runtime memory use for retained snapshot history is bounded by configuration
- The service does not retain an unbounded list of prior snapshots in memory
- Snapshot contract design does not require unbounded growth from duplicative
  device/mount bookkeeping across history

---

## Testability Requirements

### TR-001 Keep collection and composition testable without live host state

#### TR-001 Description

The service MUST keep source reading, parsing, snapshot composition, and
messaging publication testable without requiring live host disk state in unit
tests.

#### TR-001 Acceptance Criteria

- Source readers can be substituted with controlled test inputs
- Diskstats parsing can be tested from deterministic fixtures
- Mount parsing can be tested from deterministic fixtures
- Snapshot composition can be tested from deterministic sysfs, diskstats, and
  mount fixtures
- Messaging publication and request/reply flows can be tested with injected test
  doubles

---

### TR-002 Cover representative host-shape fixtures

#### TR-002 Description

The service MUST include fixture-driven tests that cover representative device,
mount, and metadata combinations present in the contract.

#### TR-002 Acceptance Criteria

- Tests cover a basic SATA disk with partitions
- Tests cover an NVMe disk with `nvme0n1p1`-style partitions
- Tests cover a `tmpfs` mount
- Tests cover a `CIFS` or `NFS` mount
- Tests cover an unmounted USB or other block device
- Tests cover missing vendor, model, or other optional sysfs metadata files
- Tests cover diskstats rows both with and without discard/flush fields

---

## Notes

- The snapshot contract intentionally separates `devices`, `filesystems`, and
  `mounts` because these sections describe related but distinct host-state
  concepts
- Device nodes are runtime identifiers, while serial and WWN are preferred
  stable identifiers when available
- This document describes the snapshot contract and behavior, not the internal
  implementation structure of readers, parsers, or builders
