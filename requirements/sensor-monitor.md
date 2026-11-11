# Sensor Monitor — Requirements

## Overview

Sensor Monitor collects available Linux host sensor readings in one process. It uses
the existing timer and polling workers, constructs its compiled collectors at
startup, and publishes one snapshot for each completed collection cycle.
Collection sources and readable names are configured separately so labels do
not change which host data is read.

This document follows `requirements/requirement-types.md`.

## Functional Requirements

### FR-001 Collect configured sensor readings

The packaged configuration MUST enable temperature, voltage, fan, throttling,
and clock sources by default. The service MUST poll enabled sources at a
configurable interval. Each configured collector MUST be
constructed at startup and called by the polling worker during each cycle.
Collectors MUST return their readings to the polling worker, which assembles
one timestamped snapshot. The snapshot MUST contain `temperatures`, `voltages`,
`fans`, and `clocks` arrays of `{key, name, value, unit}` readings, and a
`throttling` array of `{key, name, active}` readings. Collectors MUST NOT publish NATS messages or update
the history store themselves.

The service MUST read from available Linux kernel interfaces where they provide
the configured reading. Raspberry Pi firmware readings MAY use `vcgencmd` on
hosts where that command and its device are available. Command arguments MUST be selected
from validated, supported sensor operations rather than arbitrary config text.

### FR-002 Publish partial snapshots

An absent, unsupported, or unreadable sensor MUST NOT remove readings from
other sensors in the same polling cycle. The snapshot MUST omit a reading that
was not obtained. Collection failures MUST be observable in service logs and
MUST NOT be represented as a successful numeric reading.

### FR-003 Resolve human-readable sensor names

Each reading MUST have a stable machine-readable key. The service MUST use a
meaningful source-provided label when available, then a configured JSON name
mapping, and finally the raw key as fallback. Name mappings MUST be separate
from source selection. Dynamic sysfs directory indexes such as `hwmon0` MUST
NOT be used as stable sensor identities.

### FR-004 Maintain bounded history and latest state

The service MUST retain a configurable, bounded number of complete snapshots
in chronological order. The most recent snapshot MUST remain available even
when history retention is disabled.

### FR-005 Publish snapshots and provide NATS RPCs

After a completed collection cycle, the service MUST update its latest state
and history and publish a snapshot event through the shared messaging client.
It MUST provide request/reply handlers for the latest snapshot and retained
history, following the existing metrics-service response pattern. Publication
errors MUST be logged.

### FR-006 Configure source selection

The service MUST discover standard Linux sensor interfaces by default. It
MUST allow individual readings to be selected by stable key when a host needs
a narrower set. Raspberry Pi firmware collection MUST be optional on hosts
without the required command or device. The names JSON MUST map a category
and stable key to a human-readable name without selecting an input source.
