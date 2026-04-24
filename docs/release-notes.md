# Release Notes

## Why release notes are needed early

LiteNAS is still in the platform-skeleton stage, but release notes are useful
already.

At this stage, many changes do not add much immediate business value. They
still change the shape of the platform in important ways:

- service and app skeletons
- runtime and messaging wiring
- packaging and installation behavior
- deployment expectations
- CI/CD and release reproducibility

Release notes should make those changes visible so platform progress can be
tracked intentionally rather than inferred from commit history.

## Format

Use one section per release.

Recommended structure:

```md
## X.Y.Z - YYYY-MM-DD

### RL-X.Y.Z Summary

- One short paragraph or 2-4 bullets describing the release intent.

### RL-X.Y.Z Added

- New service, app, module, script, or packaging capability.

### RL-X.Y.Z Changed

- Important behavior or structure changes.

### RL-X.Y.Z Fixed

- Important defects resolved.

### RL-X.Y.Z Platform

- CI/CD, packaging, deployment, reproducibility, or developer-workflow changes.

### RL-X.Y.Z Notes

- Optional limitations, follow-up work, or intentionally incomplete areas.
```

## Guidance

- Prefer business-meaningful summaries over commit-by-commit narration.
- It is acceptable for early releases to emphasize platform and infrastructure
  value over direct end-user value.
- Mention intentionally incomplete slices when a release mainly prepares later
  product work.
- When a release introduces or changes installation, packaging, runtime
  dependencies, or deployment behavior, record that explicitly.
- Use release-qualified subsection headings such as `RL-0.1.0 Summary` and
  `RL-0.1.0 Platform` so markdown headings stay globally unique inside the
  document.
- Keep wording factual and concise.
