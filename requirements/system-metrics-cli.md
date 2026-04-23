# System Metrics CLI — Requirements

## Overview

The System Metrics CLI is a read-only command-line client for the System
Metrics service. It requests either the current system metrics snapshot
or the recent metrics history and renders the result for terminal use.

The CLI is intended for direct operator usage on a LiteNAS host. By
default it presents the current snapshot in a human-readable format
instead of raw JSON.

This document follows the requirement classification defined in
`requirement-types.md`.

---

## Functional Requirements

### FR-001 Request the current metrics snapshot by default

#### FR-001 Description

The CLI MUST request the current system metrics snapshot when it is
invoked without a history flag.

#### FR-001 Input

- CLI invocation without `--history`
- CLI configuration for messaging access

#### FR-001 Output

- Request for the current system metrics snapshot
- Current snapshot response data

#### FR-001 Acceptance Criteria

- Running the CLI with no feature flags requests the current snapshot
- Running the CLI with `--cpu` requests the current snapshot
- Running the CLI with `--ram` requests the current snapshot
- Running the CLI with both `--cpu` and `--ram` requests the current snapshot

---

### FR-002 Render the current snapshot in a human-readable format

#### FR-002 Description

When rendering the current metrics snapshot, the CLI MUST print a
human-readable terminal format instead of JSON.

#### FR-002 Input

- Current system metrics snapshot from FR-001

#### FR-002 Output

- Formatted terminal output containing CPU and RAM sections

#### FR-002 Acceptance Criteria

- Default output includes a CPU section and a RAM section
- The CPU section includes total CPU load and per-core load entries
- The RAM section includes total RAM, used RAM, and available RAM values
- The current snapshot timestamp is not printed in the default human-readable output
- Output is formatted for direct terminal reading rather than JSON parsing

---

### FR-003 Filter current snapshot sections

#### FR-003 Description

The CLI MUST support section-level filtering for the current snapshot so
operators can request only CPU output, only RAM output, or both.

#### FR-003 Input

- `--cpu` flag
- `--ram` flag
- Current system metrics snapshot from FR-001

#### FR-003 Output

- Human-readable CPU section
- Human-readable RAM section
- Combined human-readable output when both sections are selected

#### FR-003 Acceptance Criteria

- `--cpu` prints only the CPU section
- `--ram` prints only the RAM section
- Providing both `--cpu` and `--ram` prints both sections
- When both sections are printed explicitly, they are separated by a visual separator
- When neither `--cpu` nor `--ram` is provided, both sections are printed

---

### FR-004 Request metrics history

#### FR-004 Description

The CLI MUST request metrics history when invoked with the history flag.

#### FR-004 Input

- CLI invocation with `--history`
- CLI configuration for messaging access

#### FR-004 Output

- Request for metrics history
- Metrics history response data

#### FR-004 Acceptance Criteria

- Running the CLI with `--history` requests metrics history instead of the current snapshot
- History retrieval uses the service-provided retention window
- Empty history results are allowed

---

### FR-005 Render history as pretty-printed JSON

#### FR-005 Description

When rendering metrics history, the CLI MUST print the history payload as
pretty-printed JSON.

#### FR-005 Input

- Metrics history from FR-004

#### FR-005 Output

- Pretty-printed JSON history output

#### FR-005 Acceptance Criteria

- History output is valid JSON
- History output is indented for readability
- History entries retain their timestamps in the JSON output
- History output is written to standard output

---

## Interface Requirements

### IR-001 Provide a flag-based command-line interface

#### IR-001 Description

The CLI MUST expose a flag-based interface for selecting snapshot mode,
history mode, and current-snapshot section filters.

#### IR-001 Input

- Process arguments

#### IR-001 Output

- Parsed execution mode and output options

#### IR-001 Acceptance Criteria

- The CLI accepts `--history` to select history mode
- The CLI accepts `--cpu` to select the CPU section for current snapshot output
- The CLI accepts `--ram` to select the RAM section for current snapshot output
- The CLI accepts invocation without feature flags
- The CLI rejects unknown flags with a non-zero exit status

---

### IR-002 Write user-facing output to standard output

#### IR-002 Description

The CLI MUST write successful command results to standard output in the
format required by the selected mode.

#### IR-002 Input

- Current snapshot output from FR-002 and FR-003
- History output from FR-005

#### IR-002 Output

- Human-readable snapshot text on standard output
- Pretty-printed JSON history on standard output

#### IR-002 Acceptance Criteria

- Successful current snapshot output is written to standard output
- Successful history output is written to standard output
- Error text is not mixed into successful standard output content

---

## Reliability Requirements

### RR-001 Fail clearly when snapshot retrieval cannot complete

#### RR-001 Description

The CLI MUST fail with a clear non-zero result when it cannot load
configuration, connect to the messaging system, or retrieve the requested
data.

#### RR-001 Acceptance Criteria

- Invalid configuration causes a non-zero exit status
- Messaging request failures cause a non-zero exit status
- Response decoding failures cause a non-zero exit status
- Failures produce an error message on standard error

---

## Operational Requirements

### OR-001 Support an explicit configuration file path

#### OR-001 Description

The CLI MUST support selecting a specific configuration file path at
runtime.

#### OR-001 Acceptance Criteria

- The CLI supports an explicit config-path option
- The CLI uses a default config path when no explicit path is provided
- Configuration path selection does not change command behavior apart from
  which config file is loaded

---

## Testability Requirements

### TR-001 Keep rendering testable without live infrastructure

#### TR-001 Description

The CLI MUST keep output rendering and argument handling testable without
requiring a live messaging system.

#### TR-001 Acceptance Criteria

- Argument parsing can be tested independently of runtime wiring
- Human-readable snapshot rendering can be tested with fixture snapshots
- History JSON rendering can be tested with fixture history entries
- Messaging interactions can be replaced by test doubles in automated tests

---

## Notes

- This document uses the existing repository component name
  `system-metrics-cli`
- This document defines the target user-facing CLI behavior and does not
  preserve the current `current` and `history` positional-command interface
- Open question: whether `--history` should be rejected when combined with
  `--cpu` or `--ram`, or whether those flags should be ignored in history mode
