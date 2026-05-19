# Zpool Status Parser Specification

## Purpose

This document defines the grammar and AST contract for parsing `zpool status`
 output in LiteNAS.

It is the implementation blueprint for ANTLR4-based parsing in the ZFS parser
domain.

## Input Scope

Initial scope targets standard `zpool status` output blocks:

- `pool: <name>`
- `state: <state>`
- `scan: <scan-summary>`
- `config:` section
- `errors: <errors-summary>`

Initial scope does not include:

- verbose/history script output variants outside standard status blocks
- non-status command output mixed into the same stream

## Grammar Coverage (V1)

The grammar should parse one or more pool status blocks from a single input.

Each pool block contains:

1. `pool` header line
2. optional metadata lines (`state`, `status`, `action`, `see`, `scan`)
3. `config:` section with a table header and indented rows
4. trailing `errors:` line

### Config Table Rules

The `config:` table must support:

- a header line containing at least `NAME` and `STATE`
- optional columns such as `READ`, `WRITE`, `CKSUM`
- indented row hierarchy representing pool -> vdev -> leaf devices
- multi-level nesting based on indentation depth

The parser should preserve:

- original row order
- per-row indentation depth
- raw token values for each parsed column

Current grammar file:

- `shared/go/parsers/zfs/status/grammar/ZpoolStatus.g4`

Note: start-of-line indentation is handled in parser-side mapping rather than
relying on a pure `INDENT` token rule.

## AST Contract (V1)

The parser should return a typed AST with this shape direction:

- `StatusDocument`
  - `Pools []PoolBlock`
- `PoolBlock`
  - `PoolName string`
  - `Metadata PoolMetadata`
  - `Config ConfigTree`
  - `ErrorsSummary string`
- `PoolMetadata`
  - `State string`
  - `Status string` (optional)
  - `Action string` (optional)
  - `See string` (optional)
  - `Scan string` (optional)
- `ConfigTree`
  - `Header []string`
  - `Roots []ConfigNode`
- `ConfigNode`
  - `Name string`
  - `Columns map[string]string`
  - `Indent int`
  - `Children []ConfigNode`

## Parse Modes

The parser API should support:

- `Strict`: grammar or structure mismatch returns an error and diagnostics.
- `Tolerant`: parser keeps recoverable fragments and returns diagnostics.

Both modes must always return structured diagnostics with line and column.

## Diagnostics Contract

Each diagnostic should contain:

- severity (`error` or `warning`)
- message
- line
- column
- optional parser rule context

## Generated Code Layout

ANTLR-generated files should be emitted only under:

- `shared/go/parsers/generated/zfs/zpoolstatus/`

Handwritten code for this parser remains under:

- `shared/go/parsers/zfs/status/`

This split keeps generated files in one exclusion-friendly location while
keeping domain logic and AST mapping reviewable.

## Public Parser API Direction

Initial API direction:

- `ParseZpoolStatus(input string, mode ParseMode) (StatusDocument, []Diagnostic, error)`

Behavior:

- `error` is reserved for non-recoverable parser/runtime failures.
- grammar and structure issues should primarily be represented via diagnostics.

## Test Fixtures (V1)

Add fixture-driven tests for:

1. single healthy pool (`errors: No known data errors`)
2. mirrored pool with nested vdev rows
3. degraded/faulted sample with non-zero error counters
4. multi-pool input in one text payload
5. malformed payload cases for strict/tolerant behavior comparison

## Open Items

1. Confirm if `zpool status -P` output is in-scope for V1.
2. Confirm if numeric columns must be normalized in parser layer or evaluator
   layer.
3. Confirm whether `errors` line should be preserved verbatim in addition to
   parsed fields.
