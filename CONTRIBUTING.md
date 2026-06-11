# Contributing

This document records intentional repository conventions for contributors and
automated code assistants. These rules are project-specific defaults chosen for
readability, maintenance, and consistent review quality.

## General Principles

- Prefer code that is easy to scan over code that is merely compact.
- Keep logic flat where possible. Reducing nesting usually improves readability.
- Prefer explicit, named helpers over clever inline implementation details.
- Keep changes focused. Avoid mixing unrelated cleanup with feature work unless
  the cleanup is necessary to make the change correct and maintainable.

## Modules and Runtime Wiring

- Prefer a module-based construction approach for service wiring when bootstrap
  code starts to grow.
- Modules are a construction/composition layer. They should assemble related
  dependencies and return them as a small bundle.
- Modules should remain as close to pure construction as practical:
  - do not start goroutines in module constructors
  - do not register background loops implicitly
  - do not hide side effects behind construction helpers
- Runtime startup and shutdown actions should stay outside modules in explicit
  runtime/orchestration functions.
- Prefer splitting service wiring into small modules by responsibility, for example:
  - infrastructure modules for config, logger, and messaging setup
  - worker modules for channels and worker construction
  - state or runtime modules for in-memory stores and runtime-owned state
- Prefer creating the most dependency-free modules first, such as channels, and
  let later modules receive the previously constructed dependencies explicitly.
- This split is preferred because it reduces cognitive load and complexity
  scores while preserving testability and a mostly pure composition layer.
- Keep module constructors deterministic and easy to test. They should mostly
  transform explicit inputs into assembled outputs.

## Software Requirements and Traceability

- All services and apps must be defined by traceable and testable software
  requirements before implementation begins.
- Requirements documents must exist in the `requirements/` directory and
  provide a clear foundation for drafting the service or app.
- This approach ensures that all functionality is traceable back to a specific
  requirement and is inherently testable.

## Boundary Validation

- Every user-facing or service-facing input boundary must validate received
  data before doing further message processing or invoking business logic.
- HTTP and HTTPS browser-facing or service-facing routes should use Huma DTO
  validation tags for request bodies, parameters, headers, and cookies.
- Non-HTTP Go boundaries should use `go-playground/validator/v10` by default.
  This includes CLI command inputs, config-shaped command payloads, NATS RPC
  request handlers, NATS subscription handlers, and other message handlers that
  receive data from outside the current package.
- Frontend runtime input validation should use Zod by default. TypeScript DTO
  types are compile-time contracts only and do not replace runtime validation.
- Frontend schemas should live outside React components where practical, such as
  under app-local `validation/` modules, so forms, providers, and API actions can
  reuse the same rules.
- Validation should happen at the transport or adapter boundary. Controllers,
  command handlers, RPC handlers, and subscription handlers should extract or
  map validated values before calling lower-level services.
- Services should receive already-validated plain values or service-owned
  request structs. Do not pass HTTP DTOs, raw NATS envelopes, or unvalidated
  transport-shaped objects into service logic.
- Validation rules should cover semantic constraints, not only type shape. Use
  required fields, string length limits, allowed values, patterns, and custom
  validators where the contract requires them.
- Form components may keep user-facing field state and display messages, but
  context/provider/API layers should still validate submitted payloads before
  sending requests.

## Testing

- Prefer requirement-traceable tests; use source-qualified IDs such as
  `system-metrics-svc/FR-001`.
- Prefer small, single-purpose tests. A test should usually verify one
  behavior and use at most 2-3 closely related assertions.
- When a matching requirement exists, annotate the test with the requirement
  identifier in a short comment directly above the test.
- Requirement references in tests should use a source-qualified form so the ID
  stays globally unambiguous, for example `system-metrics-svc/FR-001` instead
  of only `FR-001`.
- Prefer comment-based traceability such as `// Requirements:
  system-metrics-svc/FR-001, system-metrics-svc/IR-002` unless a language or
  framework already provides a clearly better native mechanism.
- Prefer package-local tests next to the code they cover for unit tests. Use a
  separate top-level test area only for true integration or end-to-end tests.
- Prefer fixtures and helper builders for repetitive test data preparation.
  Helpers should create valid baseline inputs by default, and tests should
  override only the fields relevant to the behavior under test.
- When multiple tests in one package repeat the same setup shape, extract that
  setup into package-local fixtures or helper functions instead of duplicating
  inline construction across test bodies.
- Treat test duplication as real duplication. Do not leave repeated setup,
  request builders, or assertion scaffolding inline just because the code is
  test-only.
- Prefer package-local or subproject-local `testutil` helpers when the reuse is
  local to one package or subproject.
- Prefer `shared/go/testutil/...` only when the same testing primitive is
  reused across multiple Go modules or subprojects.
- For success-path setup that is expected to succeed, prefer `must...` helpers
  so test bodies stay focused on the behavior under test.
- Apply the same `must...` rule to controlled helper-driven operations during
  tests, such as invoking registered handlers, executing fixture flows, or
  calling setup helpers that are expected to succeed. Keep explicit `err`
  assertions inline only when failure is the behavior under test.
- For repeated positive and negative setup shapes, prefer named builders such
  as `newValidX`, `newFailingX`, `newUnauthorizedX`, or similarly explicit
  helpers instead of repeated direct struct literals.
- Keep fixtures focused on data preparation. Avoid helpers that both build data
  and perform assertions unless the helper is a narrow assertion helper.
- For config parsing and similar tests, separate success-path assertions from
  error-path assertions:
  - success tests should assert the returned value
  - error tests should assert the returned error
  - avoid mixing successful field validation and invalid-input validation in
    the same test
- When validating multiple boundary values or multiple variants of the same
  behavior, prefer table-driven tests instead of long ad hoc tests.
- When a returned struct has several fields, prefer table-driven tests or
  focused helper-backed tests so each test remains small and readable.
- Name helpers clearly and idiomatically based on purpose, for example valid
  fixture builders such as `validConfigFixture` or `newTestEnvelope`, and
  focused assertion helpers such as `assertTimeout` or `assertInvalidConfig`.
- Repository-wide duplication is checked in CI and local analysis scripts for
  Go and shell code, so repeated test scaffolding can fail validation even when
  it is spread across separate Go modules.

## System Tests

- Top-level `tests/` are system-level tests for an installed or otherwise
  running LiteNAS system. They are higher integrity than service-local unit,
  integration, and contract tests and should verify externally visible behavior.
- Non-unit and non-integrity tests for LiteNAS services and apps should use the
  Python HyperionTF suite by default so infra, CLI, API, UI, and visual checks
  share one logging and execution ecosystem.
- Keep system tests grouped by category under `tests/infra`, `tests/cli`,
  `tests/api`, and `tests/ui`. Test files inside those folders should be
  focused suites for one feature, service, or workflow.
- Store HyperionTF UI page objects under `tests/ui/page_objects`. UI tests
  should import those page objects instead of declaring browser locators inline.
- Model UI page-object composition as a hierarchy that follows the real
  implementation and naming, while favoring user-facing layers over strict
  React or DOM mirroring. Pages should expose top-level regions and widgets;
  widgets should expose their own child elements and nested widgets. Avoid flat
  page objects that list every descendant as a direct page member.
- Keep pages, navigation bars, sidebars, menus, forms, and meaningful widgets;
  omit purely structural wrappers when they do not own reusable user
  interactions.
- Treat UI page objects as reusable domain-facing test support, not just
  locator assemblies. Put meaningful repeated interactions on the page or
  widget object that owns the UI behavior, so tests describe workflows without
  duplicating low-level click, fill, and wait sequences.
- Every UI page object class, decorated page-object member, and public or
  private page-object method must have a docstring. Use those docstrings to
  document the modeled UI role, element or widget composition relationship,
  interaction contract, preconditions, and side effects where relevant.
- Every system test must use exactly one category marker: `infra`, `cli`,
  `api`, or `ui`. Add domain markers such as `Auth`, `SystemMetrics`, or
  `WebGateway` when the behavior belongs to a specific service or app.
- Every system test must have a docstring that describes the test case for the
  HTML report. Include preparation, action, and expected result at intent level
  rather than listing implementation details.
- System tests should usually have one verification point. Split independent
  outcomes into separate tests, parametrized cases, or shared test-case steps
  instead of adding unrelated assertions to one test.
- Resolve system-test duplication with fixtures, helpers, parametrization, or
  shared test-case steps. Do not copy setup or assertion scaffolding between
  category suites.
- HyperionTF page-object members decorated with `@element`, `@elements`,
  `@widget`, `@widgets`, or similar decorators intentionally return locators in
  source code while exposing elements or widgets at runtime. Annotate those
  members as the runtime object exposed to tests for IDE lookup and readable
  test code.
- The HyperionTF decorated page-object mismatch is the only allowed reason in
  system tests to temporarily disable static-analysis type checks. Keep each
  ignore narrow, local to the affected decorated member, and do not apply it to
  fixtures, helper code, API tests, CLI tests, or ordinary test logic.
- `scripts/test-python.sh` is the canonical runner. It executes categories in
  fast-fail order: infra, CLI, API, then UI. It also fixes the working directory
  to `tests/` so HyperionTF logs are always written under `tests/logs`.

## Large Test Suites

- When a test suite grows large enough that support code obscures the behavior
  under test, split test support by role into separate package-local `*_test.go`
  files.
- As a practical rule of thumb, this usually makes sense once a test suite
  reaches roughly 200+ lines or otherwise becomes hard to scan.
- Apply this split only when it improves readability. Do not create multiple
  near-empty files for tiny packages, one-off helpers, or very small local
  fixtures.
- Keep Go's idiomatic `*_test.go` suffix. Do not prefer `test_*.go` naming.
- When splitting test files by role, prefer names such as:
  - `module_test.go` for the actual test functions
  - `module_helpers_test.go` for assertion helpers and small reusable test utilities
  - `module_fixtures_test.go` for valid baseline builders and fixture loaders
  - `module_testcases_test.go` for large reusable table-driven case sets
  - `module_testdata_test.go` for large inline sample values or code-based test data

## Nested Function Literals

- Avoid nested function declarations by default.
- Inline function literals increase cognitive load by forcing readers to parse
  local implementation structure instead of following the main logic.
- Prefer standalone pure functions or named helpers for reusable or non-trivial
  logic.
- Standalone helpers are easier to document, test, reuse, and maintain.
- Inline function literals are allowed only when intentionally building,
  wrapping, or decorating and returning a function.
- For decorator-style helpers, prefer names such as `buildX`, `wrapX`, or
  `decorateX`.
- Small function literals used only as compact selectors in table-driven tests
  are acceptable when extracting them would clearly reduce readability.

## Shell Scripts

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

## Packaging And Local Deployment

- Treat Debian packaging in this repository as a single-package workflow.
- Do not split services or apps into separate `.deb` outputs unless packaging
  redesign is an explicit goal of the task.
- When the goal is local build or local redeploy work, prefer updating the
  local build and deployment scripts instead of changing Debian packaging.
- Prefer ordered aggregate orchestration scripts for full local redeploy work
  when shared infrastructure steps such as config deployment, certificate
  rotation, and NATS restart must run only once.

## CI Workflow Shape

- Prefer CI graphs that reflect real dependency gates instead of flattening all
  build and test work into one broad stage.
- Keep shared-module validation as an explicit upstream gate when services or
  apps depend on shared modules.
- Prefer combining closely coupled build and test work into the same CI job
  when repeated environment setup would dominate runtime, especially for Go
  jobs where toolchain setup is a large fraction of job cost.
- Prefer splitting CI jobs by category, such as shared modules, services,
  apps, and packaging, when that improves graph readability and failure
  localization.
- Use matrices inside a category to keep parallelism, but keep the category as
  a distinct job boundary when it represents a meaningful stage in the
  pipeline.
- Keep artifact upload and download explicit in top-level workflow files when
  later jobs depend on them. Do not hide artifact flow behind reusable
  workflow boundaries.
- Prefer composite actions under `.github/actions/` for repeated CI step
  sequences such as toolchain setup, Go build/test execution, and package
  build/validation, while leaving job dependencies and artifact flow visible
  in the workflow YAML.

## Release Notes

- Keep release notes in a dedicated repository document rather than inferring
  release intent from commit history alone.
- Use the format documented in [docs/release-notes.md](docs/release-notes.md).
- Use release-qualified subsection headings such as `RL-0.1.0 Summary` so
  repeated sections remain markdownlint-compliant and easy to scan.
- Early LiteNAS releases may legitimately emphasize platform, packaging,
  deployment, and reproducibility work even when direct end-user value is low.
- When a change mainly completes infrastructure needed for later product work,
  say that explicitly in release notes instead of overstating user-facing value.

## Versioning Workflows

- Prefer a dedicated developer script for coordinated version bumps instead of
  manual multi-file edits.
- Use `scripts/dev/bump-versions.sh` when updating the default Debian package
  version and the `lite-nas/shared` dependency version referenced by services
  and apps.
