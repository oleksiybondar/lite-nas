# Admin Panel Agent Instructions

These instructions apply to work under `apps/admin-panel/`.

They supplement the repository-level `AGENTS.md`. When these instructions
conflict with the repository-level file, these local instructions take
precedence.

## Component File Organization

- **One component per file.** A `.tsx` file contains exactly one exported component.
- If a component needs private sub-components, helper functions, or shared types, create a
  folder with the same name:

  ```text
  ComponentName/
    ComponentName.tsx    ← the one public component
    SubPart.tsx          ← private sub-component (if needed)
    helpers.ts           ← pure functions, no JSX
    types.ts             ← shared or exported types; inline-only types stay in their own file
    index.ts             ← public API (only what callers should import)
  ```

- Mutually recursive sub-components that would create circular imports may share one file,
  named after their logical role (e.g. `AppSidebarTree.tsx`), not the parent component.
- Never add a second exported component to an existing `.tsx` file — create a new file instead.
- Do not import internal files of another component's folder. Use the folder's `index.ts`.
- Always use path aliases (`@components`, `@hooks`, `@contexts`, `@dto`, `@schemas`,
  `@helpers`, `@routes`, `@pages`). Avoid `../../` relative paths when an alias works.

## React Context Usage

- Prefer destructuring values returned by context hooks at the call site, for
  example `const { get } = useApi();` or `const { login } = useAuth();`.
- Avoid storing a whole context object only to call `context.method(...)` when
  the required members are known locally.
- Keep provider values focused on the stable contract exposed by the context,
  while implementation helpers can stay private to the provider module.

## Routing And Navigation

- Keep route definitions decomposed by domain. The root router should assemble
  child route modules instead of owning every route inline.
- For nested domains, use nested route modules as well. For example, a
  `system` route module can assemble `performance` and `sensors` route modules.
- Keep the route tree and sidebar navigation tree related but decoupled. Routes
  define renderable URLs; navigation defines labels, icons, grouping, and
  sidebar/flyout behavior for those URLs.
- Sidebar navigation should be modeled as a tree, not as a flat list plus
  repeated section headers. Parent categories such as `System`, `Performance`,
  and `Sensors` should be represented explicitly and may have child items.
- Parent navigation categories should have real landing routes when they
  represent meaningful areas, for example `/system`, `/system/performance`, and
  `/system/sensors`.
- Category landing pages must be designed as useful overview pages with cards,
  descriptions, status previews, or entry points. Do not duplicate the sidebar
  menu as plain icon-and-label lists inside page content.
- Desktop expanded sidebars should render nested, expandable navigation.
- Desktop collapsed sidebars should use icon-first navigation with flyout or
  popover access to child items.
- Mobile navigation should not depend on hover. Use tap-driven drawer,
  collapse, menu, or landing-page behavior for nested navigation.
- Selected and expanded navigation state should be derived from the current
  location path whenever possible instead of being duplicated in local component
  state.

## Coverage Scope

- Admin-panel coverage must include the whole `src` tree by default.
- Do not narrow coverage includes or add exclusions only to make coverage
  thresholds pass.
- Exclusions are allowed only for a clearly documented technical reason, such as
  generated files or files that cannot be executed in the Vitest environment.
