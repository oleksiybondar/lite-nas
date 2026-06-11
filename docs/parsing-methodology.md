# Unified Parsing Methodology

## Purpose

This document defines the repository-wide parsing approach for LiteNAS.

The goal is to keep parsing strict, reviewable, and consistent across domains
instead of mixing ad-hoc parsing styles.

## Decision

LiteNAS uses parser generators as the default and common parsing pattern.

ANTLR4 is the selected foundation for new parser implementations.

Handwritten parsers are not the default path for domain grammars in LiteNAS.
They introduce higher long-term support cost, more drift in parser behavior,
and inconsistent error handling between teams and services.

## Scope

This methodology applies to:

- CLI output parsers, including `zpool` and `zfs` outputs used for metrics and
  diagnostics snapshots.
- Configuration parsers used by hardening and compliance workflows, including
  domains such as `sshd`, `pam`, `ufw`, `aide`, `usbguard`, and similar tools.

## Rationale

Using one grammar-driven approach gives LiteNAS:

- one consistent parser architecture across services and apps
- strict, explicit grammar contracts per domain
- predictable diagnostics (line/column, token context, parse failures)
- repeatable AST generation and evaluator integration
- reduced maintenance overhead versus many custom handwritten parsers
- easier onboarding and review because grammars follow one style

## Repository Policy

All new domain parsers should follow these rules unless a user explicitly asks
for a different approach for a specific task:

1. Define grammar with ANTLR4 in a domain-owned grammar file.
2. Generate parser artifacts for Go targets.
3. Build a typed AST layer from parse trees.
4. Keep evaluation and policy logic separate from parsing logic.
5. Expose a consistent parse result contract for downstream services.

## Unified Output Contract Direction

To support cross-domain reuse, parser packages should converge on a common
result shape:

- parsed domain payload (typed AST and/or mapped model)
- diagnostics collection (errors and warnings with source location)
- parser mode metadata (for example strict vs tolerant mode)

This keeps all services aligned on one parsing methodology even when grammars
and business domains differ.

## Immediate Application to ZFS/Zpool

For upcoming ZFS metrics preparation work:

- `zpool list` and `zpool iostat` remain part of the same parsing strategy
  rather than being treated as one-off string split utilities.
- `zpool status` parsing should produce typed AST that can be evaluated into
  snapshot-ready health and error details.
- The combined snapshot model should be assembled from consistently parsed
  command outputs.
