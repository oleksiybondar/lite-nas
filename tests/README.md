# LiteNAS System Tests

The `tests/` project contains LiteNAS system-level tests. These tests verify a
running installed LiteNAS system through externally visible behavior. They sit
above service-local unit, integration, and contract tests that live beside the
Go, TypeScript, or JavaScript code they cover.

## Test Categories

System tests are grouped into four top-level categories:

- `infra/`
  Host, package, service, daemon, filesystem, and system configuration checks.
- `cli/`
  Command-line behavior and terminal-facing workflows.
- `api/`
  Browser-facing and service-facing HTTP API behavior through the web gateway.
- `ui/`
  Browser UI, visual, and end-to-end administration flows.

Test files inside each category should represent suites for a focused feature
or workflow, for example `api/test_auth.py` or `infra/test_services.py`.

The Python test runner executes categories in this fast-fail order:

1. `infra`
2. `cli`
3. `api`
4. `ui`

## Required Markers

Every system test must have one category marker:

- `@pytest.mark.infra`
- `@pytest.mark.cli`
- `@pytest.mark.api`
- `@pytest.mark.ui`

Tests should also use domain markers when the behavior belongs to a specific
LiteNAS service or app, such as `Auth`, `SystemMetrics`, or `WebGateway`.
Parametrized tests should apply domain markers per parameter when each case
belongs to a different service.

## Test Case Docstrings

Every system test function must have a docstring. HyperionTF uses the docstring
as the test case description in HTML logs, so the docstring must explain the
test case intent rather than restating implementation details.

Use this structure:

```python
"""Test case: short behavior name.

Preparation:
- State what installed system state, account, service, or fixture is expected.

Action:
- State the user-visible or system-visible action.

Expected result:
- State the single behavior this test verifies.
"""
```

Keep the docstring useful for a reader who only opens the HTML report and wants
to understand what was proven, not every low-level command used to prove it.

## Verification Scope

System tests should usually have one verification point. If a workflow needs
to verify multiple independent outcomes, split it into separate tests or use
parametrization so each edge case remains a separate test case.

Use setup actions freely, but keep failing assertions tightly related to the
test case name and docstring. Do not assert unrelated service state, response
fields, or UI details just because they are available.

When repeated behavior is useful across tests, prefer fixtures, helpers, or
shared test-case steps instead of duplicating setup and assertions inline.
Duplication in system tests is maintenance debt and is checked by CI.

## Technology Boundary

Non-unit and non-integrity tests for LiteNAS services and apps should use the
Python HyperionTF suite by default. HyperionTF keeps infrastructure, CLI, API,
UI, and visual verification in one ecosystem and produces consistent HTML logs.

Service-local unit, integration, and contract tests should remain in the
service or app project using that project's native test framework.

## Logs

`scripts/test-python.sh` always runs with `tests/` as the working directory.
HyperionTF logs therefore resolve predictably to `tests/logs`.

The runner clears `tests/logs` once before the first category starts. CI uploads
that directory as a test artifact for post-failure review.
