# Agent Instructions

These instructions apply to all AI agents working in this repository, including
CLI-based agents and IDE-integrated agents.

## Instruction Precedence

Repository-level instructions in this file apply by default across the whole
monorepo.

For each subproject, such as an app, service, or shared component library, a
more specific `AGENTS.md` within that subproject should also be used when it is
available.

When both this repository-level file and a subproject-level `AGENTS.md` apply:

- agents must follow both when the instructions are compatible
- the subproject-level `AGENTS.md` takes precedence when the instructions
  conflict

## Git and Repository Operations

Agents may:

- Read repository state with commands such as `git status`, `git diff`,
  `git log`, and `git show`.
- Create, edit, and delete files as required by an explicit user request.
- Stage files with `git add` when preparing or inspecting the agent's own
  changes.
- Create a pull request only when there is already an appropriate pushed branch
  or pushed state available.

Agents must not:

- Create commits.
- Amend commits.
- Create tags.
- Push branches, tags, or commits.
- Force-push.
- Stage unrelated user changes.
- Run destructive Git commands such as `git reset --hard`, `git checkout --`,
  `git clean`, or equivalent history/worktree rewriting commands unless the
  user explicitly requests that exact operation.
- Push a branch only for the purpose of opening a pull request.

If a pull request would require pushing local changes first, do not create the
pull request. Instead, explain that a PR can be created after the user pushes the
branch or otherwise provides a pushed source branch.

## Change Handling

- Keep edits focused on the user's request.
- Preserve user changes already present in the worktree.
- Do not revert unrelated files.
- Report validation performed and any commands that could not be run.
- When tests repeat the same setup within a package, prefer package-local
  fixtures and helper files instead of copying the setup into each test.

## Tooling Conventions

- Use existing scripts and configs where available.
- Prefer `./scripts/install-dev-dependencies.sh` for local developer tooling.
- Prefer `./scripts/run-ci.sh` for the full local static-analysis check.
- Prefer scripts under `scripts/ci/` for CI-equivalent checks.
- Prefer scripts under `scripts/format/` for manual formatting.
- Shell scripts should resolve repository helpers and sourced script modules
  dynamically from their own location, for example with `$SCRIPT_DIR`.
- It is acceptable and preferred to disable ShellCheck `SC1091` inline for
  intentional dynamic `source` calls that load repository helpers or script
  modules. Keep the suppression directly above the affected `source` line and
  do not use it for unrelated missing-file warnings.
- Treat Debian packaging in this repository as a single-package workflow.
  Do not introduce or propose per-service or per-app `.deb` splits unless the
  user explicitly asks to redesign packaging.
- When the user refers to local build or local deployment work, prefer the
  local `scripts/build-*.sh`, `scripts/deploy-*.sh`, `scripts/build-all.sh`,
  and `scripts/deploy-all.sh` flows rather than changing Debian packaging.
- For CI workflow reuse, prefer composite actions under `.github/actions/` for
  repeated step sequences. Keep job dependencies and artifact upload/download
  explicit in top-level workflow files when downstream jobs depend on them.
- Prefer dedicated repository scripts for coordinated version or release-note
  maintenance work when such scripts exist, rather than editing many related
  files manually.

## Project Conventions

- Repository-wide coding and testing conventions are documented in
  [CONTRIBUTING.md](CONTRIBUTING.md).
- Follow those conventions by default unless the user explicitly requests a
  different style for the current task.
- This includes requirement-traceable tests when requirement documents exist;
  use source-qualified IDs such as `system-metrics-svc/FR-001`.
- Follow the release-note format documented in
  [docs/release-notes.md](docs/release-notes.md)
  when the task involves preparing or updating release-facing documentation.
- Use release-qualified subsection headings such as `RL-0.1.0 Summary` when
  editing release notes so repeated sections remain unique and lint-clean.
