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

## Critical Working Rules

These rules are mandatory and should be treated as the highest-priority
repository working style after explicit user instructions.

- Never guess API contracts, interfaces, DTO shapes, route behavior, provider
  contracts, or similar integration boundaries. If the contract is not already
  clear from user instructions or existing code, ask the user before
  implementing it.
- Do not add artificial entities, states, abstractions, hooks, providers,
  helpers, files, or architectural layers that the user did not ask for and
  that are not clearly required by existing code.
- Prefer direct, reviewable, human-readable code first. Static-analysis
  compliance matters, but do not spend time optimizing, abstracting, or shaping
  code primarily to satisfy tooling before the user has reviewed and accepted
  the design.
- Keep initial implementations close to the requested behavior. Make the
  smallest coherent change that lets the user review the intended direction
  before expanding the solution.
- If a design choice is uncertain, stop and ask a focused question instead of
  filling the gap with assumptions.

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
- Validate every user-facing or service-facing input boundary before doing
  further message processing or invoking business logic.
- Use Huma DTO validation for HTTP and HTTPS routes. Use
  `go-playground/validator/v10` by default for non-HTTP Go boundaries such as
  CLI inputs, NATS RPC handlers, NATS subscription handlers, and other
  transport/message handlers.
- Use Zod by default for frontend runtime input validation. Keep schemas outside
  React components where practical so forms, providers, and API actions can reuse
  the same rules.
- Do not treat TypeScript DTO types as runtime validation. Validate submitted
  frontend payloads before sending requests from context/provider/API layers,
  even when forms also show user-facing validation messages.
- Keep validation at the adapter boundary and pass validated plain values or
  service-owned request structs into lower-level services.
- When tests repeat the same setup within a package, prefer package-local
  fixtures and helper files instead of copying the setup into each test.
- Treat test duplication as repository-wide maintenance debt, not as a
  package-local issue only. Reuse that crosses Go module boundaries should move
  into shared test helpers instead of being copied.
- In tests, prefer `must...` helpers not only for object construction but also
  for controlled helper-driven actions that are expected to succeed, such as
  invoking registered handlers or executing fixture flows. Keep inline error
  assertions only for expected failure behavior.
- Treat top-level `tests/` as system-level HyperionTF tests for an installed or
  running LiteNAS system, not as service-local unit or contract tests.
- Keep top-level system tests grouped under `tests/infra`, `tests/cli`,
  `tests/api`, and `tests/ui`. Use exactly one matching category marker on
  every system test, and add service/app domain markers where applicable.
- Store HyperionTF UI page objects under `tests/ui/page_objects`, and keep UI
  test suites under `tests/ui`.
- Model UI page-object composition as a hierarchy that follows the real
  implementation and naming, while favoring user-facing layers over strict
  React or DOM mirroring. Pages should expose top-level regions and widgets;
  widgets should expose their own child elements and nested widgets. Avoid flat
  page objects that list every descendant as a direct page member.
- Keep pages, navigation bars, sidebars, menus, forms, and meaningful widgets;
  omit structural wrappers that do not own reusable user interactions.
- Treat UI page objects as reusable domain-facing test support, not just
  locator assemblies. Put repeated page or widget interactions on the owning
  page object so tests describe user workflows instead of duplicating low-level
  click, fill, and wait sequences.
- Every UI page object class, decorated page-object member, and public or
  private page-object method must have a docstring describing the modeled UI
  role, composition relationship, interaction contract, preconditions, and side
  effects where relevant.
- In HyperionTF page objects, decorated members may return locators in source
  code while being annotated as the runtime element, elements collection,
  widget, widgets collection, or specific widget class exposed by the
  decorator. This is the only accepted system-test reason to temporarily
  disable static-analysis type checks; keep such ignores narrow and local to
  the affected decorated member.
- Give every system test a docstring with preparation, action, and expected
  result. HyperionTF uses the docstring as the HTML report test case.
- Keep each system test focused on one verification point. Use parametrization,
  fixtures, helpers, or shared test-case steps instead of asserting unrelated
  outcomes or duplicating setup.

## Tooling Conventions

- Use existing scripts and configs where available.
- Prefer `./scripts/install-dev-dependencies.sh` for local developer tooling.
- Prefer `./scripts/run-ci-checks.sh` for the full local static-analysis check.
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
- Prefer `./scripts/ci/go-duplication-analysis.sh` and
  `./scripts/ci/bash-duplication-analysis.sh` when validating repo-wide
  duplication changes locally.

## Project Conventions

- Repository-wide coding and testing conventions are documented in
  [CONTRIBUTING.md](CONTRIBUTING.md).
- Follow those conventions by default unless the user explicitly requests a
  different style for the current task.
- This includes requirement-traceable tests when requirement documents exist;
  use source-qualified IDs such as `system-metrics-svc/FR-001`.
- Follow the repository testing split for helper placement:
  - package-local or subproject-local `testutil` for local reuse
  - `shared/go/testutil/...` only for cross-module reusable testing primitives
- Follow the system testing rules in [tests/README.md](tests/README.md) when
  creating or updating top-level Python HyperionTF tests.
- Follow the release-note format documented in
  [docs/release-notes.md](docs/release-notes.md)
  when the task involves preparing or updating release-facing documentation.
- Use release-qualified subsection headings such as `RL-0.1.0 Summary` when
  editing release notes so repeated sections remain unique and lint-clean.

## Documentation Conventions

- New code should include meaningful docstrings for public and private
  structures, classes, methods, functions, type aliases, and interface members.
- Docstrings should describe the contract, correct usage, and architectural
  role when that role is relevant to understanding the code.
- When a design choice is non-obvious, add a concise inline comment near the
  implementation that explains why the code is shaped that way.
- Avoid comments that merely repeat the code. Prefer comments that clarify API
  behavior, invariants, side effects, lifecycle expectations, or integration
  boundaries.
