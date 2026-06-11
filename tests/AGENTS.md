# Agent Instructions For System Tests

These instructions apply to the top-level `tests/` project. They supplement the
repository-level `AGENTS.md`.

## System Test Scope

- Treat this directory as system-level HyperionTF tests for an installed or
  running LiteNAS system.
- Keep tests grouped under `infra/`, `cli/`, `api/`, and `ui/`.
- Use exactly one category marker on every system test: `infra`, `cli`, `api`,
  or `ui`.
- Add service or app domain markers where applicable.
- Give every system test a docstring with preparation, action, and expected
  result.
- Keep each system test focused on one verification point.

## UI Page Objects

- Store HyperionTF UI page objects under `ui/page_objects/`.
- Keep UI test suites directly under `ui/`.
- Model browser pages with HyperionTF page objects and widgets rather than
  placing locator details directly in tests.
- Model page-object composition as a hierarchy that follows the real UI
  implementation and naming, but prefer user-facing layers over strict React or
  DOM mirroring. Pages expose top-level regions and widgets; widgets expose
  their own child elements and nested widgets. Avoid flat page objects that list
  every descendant as a direct page member.
- Keep pages, navigation bars, sidebars, menus, forms, and meaningful widgets;
  omit structural wrappers that do not own a reusable user interaction.
- Page objects are not only locator assemblies. Add reusable UI-domain
  interactions to the page or widget object that owns that behavior, so tests
  can call named workflow methods instead of duplicating low-level click/fill
  sequences.
- Add docstrings to every page object class, decorated page-object member, and
  public or private page-object method. Document the modeled UI role,
  composition relationship, interaction contract, preconditions, and side
  effects where relevant.
- Page object members decorated with HyperionTF decorators return locators in
  source code but expose elements or widgets at runtime.
- For IDE lookup and readable tests, annotate decorated members as the runtime
  object they expose, such as `Element`, `Elements`, `Widget`, `Widgets`, or a
  project-specific widget class.
- This decorator/runtime mismatch is the only acceptable reason in system tests
  to temporarily disable static-analysis type checks. Keep each ignore narrow,
  local to the decorated page-object member, and explain the reason when the
  tool supports it.
- Do not add custom wrapper entities that behave like widgets while hiding
  collection lookup, deferred resolution, or stale-selection semantics. Keep
  those concerns explicit by using framework `Element` / `Elements` /
  `Widget` / `Widgets` types directly at the page-object boundary.
- Do not use the page-object type-check exception for test logic, fixtures,
  API clients, CLI clients, or helper code.
