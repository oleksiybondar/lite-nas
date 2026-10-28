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

Within `api/`, `cli/`, and `ui/`, service or app suites live in named folders,
for example `api/web_gateway/test_auth.py`, `api/file_service/`,
`cli/file_writer/`, and `ui/admin_panel/`. Shared logging manager and email
notifier CLI suites cover both system and security variants in their family
folders. Cross-service host checks remain directly under `infra/`.
UI page objects stay under `ui/page_objects/` and are imported by the UI suites.

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

Keep that single verification point user-visible and operationally simple even
when the exercised runtime chain is deeper. In the current platform slice, a
small CLI or API check may also transit service authentication, RBAC-backed
role resolution, transport-level token validation, and target-service
authorization before it reaches the final observable result.

That is expected. The test should still assert one main outcome while relying
on the integrated runtime path to provide broader system coverage implicitly.

## Security monitoring CLI tests

The authentication cases use the configured security-manager account for normal
account failures and a unique nonexistent account for unknown-account guessing.
Mount lifecycle, ramfs, and EICAR cases create temporary mounts below `/mnt`.
The clean ext4 fixture waits for Antivirus to report a normal informational scan
event before it unmounts, so cleanup cannot interrupt the scan under test.
Pytest runs as the current user. The mount fixture switches a separate shell to
the restricted `testsudouser` account. Its test-only sudoers rule permits
passwordless `sudo` for the mount utilities (`mkdir`, `mount`, `rmdir`, and
`umount`) and logging-manager CLI checks. The mount utilities accept their
normal arguments, so use this account only on disposable or dedicated test
hosts.

From the repository root, prepare the normal accounts and the sudo account with
their test-only permissions on a local test host:

```sh
sudo ./scripts/ci/prepare-system-test-users.sh
sudo ./scripts/ci/prepare-system-test-sudo-user.sh
```

The default accounts are `testuser`, `testoperator`, `testsecurityuser`,
`testadmin`, and `testsudouser`, with password `testpassword`. The scripts and
fixtures share the `LITENAS_*` overrides in `constants.py`; use the same values
for preparation and execution. CI uses `testsecurity` explicitly in both steps.

The SSH authentication test requires a running local SSH server accepting
password attempts. On a dedicated test host, prepare it with:

```sh
sudo bash ./scripts/ci/prepare-system-test-ssh.sh
```

This installs OpenSSH Server and enables password authentication for loopback
connections in `/etc/ssh/sshd_config.d/00-lite-nas-system-tests.conf`, then
restarts SSH. CI runs the same preparation after installing the package. These
scripts are test-host setup, not LiteNAS production deployment. Run pytest as
the normal user; the fixtures authenticate to the required account and use
restricted sudo for privileged commands:

```sh
cd tests
../.venv/bin/pytest -c pytest.ini cli/security_monitoring cli/logging_manager -q
```

An existing installation may retain older configuration during local redeploys.
The current authentication tests require `event_id_prefix` to be `secauth` in
`/etc/lite-nas/log-analysis/rules.json`; a newly detected `SEC_...` event means
the installed rules still use the older prefix. Update that setting and restart
`lite-nas-log-analysis` before rerunning. Historical stored IDs remain unchanged.
The EICAR tests also require `ext4` in the installed Antivirus `scan.fs_types`
setting. Keep installed monitor binaries and configuration aligned with the
checked-out source; the deployment scripts preserve existing configuration.

## UI Page Objects

UI system tests use HyperionTF page objects for browser automation. Store those
page objects under `ui/page_objects/`; keep test functions under `ui/`.

Page object classes should extend HyperionTF `WebPage` or `Widget` classes and
declare locators with HyperionTF decorators such as `@element`, `@elements`,
`@widget`, and `@widgets`. The method bodies return `By...` locators because
HyperionTF converts the decorated members into elements or widgets at runtime.

Model page-object composition as a hierarchy that follows the current UI
implementation and naming, but do not mirror every React component or DOM
wrapper. A page object should expose top-level regions and widgets; those
widgets should expose their own child elements and nested widgets. For example,
a login page owns a top bar and a login form, while the login form owns the
login field, password field, and submit button. Avoid flat page objects that
list every descendant as a direct page member.

Preserve meaningful user-facing layers such as pages, navigation bars,
sidebars, menus, forms, and reusable widgets. Omit purely structural wrappers
when they do not represent a distinct user concept or reusable interaction
boundary.

Page objects are also the home for reusable UI-domain interactions around the
page or widget they model. Prefer named page-object methods such as
`login_as(...)`, `open_user_menu()`, or `submit_form()` when the action is a
meaningful user workflow or repeated widget interaction. Tests should read in
terms of behavior, while the page object keeps locator details and low-level
click/fill sequences close to the UI surface they belong to.

Every page object class, decorated page-object member, and public or private
page-object method must have a docstring. Class docstrings should explain the
page or widget role. Decorated member docstrings should describe the represented
UI element, widget, or composition relationship. Method docstrings should
describe the user-facing interaction, expected preconditions, and relevant side
effects when they are not obvious from the method name.

For IDE navigation and readable test code, annotate decorated page-object
members as the runtime object they expose to tests, such as `Element`,
`Elements`, `Widget`, `Widgets`, or a project-specific widget class. This is
the only accepted reason in system tests to add a narrow static-analysis type
ignore around a page-object member: the source returns a locator, while the
decorator intentionally exposes a richer runtime object.

Keep ignores local to the affected HyperionTF-decorated member and include a
short reason when the tool supports it. Do not use this exception for regular
test logic, fixtures, API clients, CLI clients, or helper code.

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
those logs as `system_tests_logs.zip` for post-failure review.

## File Service API checks

The `api/file_service/` suites log in through `/api/auth/login` and send file
requests through Nginx's `/f/`, `/d/`, and `/ls/` mounts. Upload, operation,
and listing tests create unique paths in the test operator's home directory
and remove them afterward. Set
`LITENAS_FILE_SERVICE_ADMIN_LOGIN` and `LITENAS_FILE_SERVICE_ADMIN_PASSWORD`
to a dedicated admin test account before running the admin cases. The suite
fails clearly when either value is missing rather than treating the default
`testuser` account as an admin. CI provisions `testadmin` in the `admin` group
through `scripts/ci/prepare-system-test-users.sh`; it remains a nonroot Unix
account, so admin reads still obey Unix file and directory permissions.
Critical-path upload checks use the configured admin and operator accounts.
REST client logs redact access and refresh cookie
values and shorten large transfer bodies before they are written to the
HyperionTF report.

## Moments API and REST contracts

`api/moments/` checks the authenticated working directory on standalone and
embedded hosts, gallery search, album and collection lifecycle, collection
file and directory membership, and search/path injection behavior on an installed
Moments host. It creates operator-owned temporary source paths and removes the
gallery metadata and paths after each case. Set `LITENAS_MOMENTS_API_URL` when
the standalone Moments host is not `https://moments.localhost`. The suite uses
the configured `LITENAS_OPERATOR_LOGIN` and `LITENAS_OPERATOR_PASSWORD` account.

`api/test_rest_contracts.py` marks live JSON Schema checks with `contract`. The
gateway current-user, File Service stat, and Moments gallery-list checks
validate actual JSON responses with HyperionTF's
`response.validate_json_schema()`. Stream Service serves media bytes, so its
contract case validates its published OpenAPI JSON document instead. Schemas
live in `schemas/rest_contracts.py` and reflect the service DTOs and published
route definitions. From `tests/`, run
`../.venv/bin/pytest -c pytest.ini api/moments api/test_rest_contracts.py` on an
installed system.
