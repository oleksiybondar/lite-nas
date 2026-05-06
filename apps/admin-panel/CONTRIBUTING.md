# Contributing — Admin Panel

This document records frontend conventions for the `admin-panel` React/TypeScript application.
It supplements the repository-level [CONTRIBUTING.md](../../CONTRIBUTING.md), which covers
Go services, CI, shell scripts, and cross-cutting rules.

## Component File Organization

- One component per file. A file named `AppSidebar.tsx` contains exactly one exported component.
- A props type or interface that is only used by a single component stays inline in that component's
  file.
- When a component requires private sub-components, helper functions, or types shared across
  multiple files, move everything into a named folder with the same name as the public component.

  ```text
  navigation/
    AppSidebar/
      AppSidebar.tsx          ← public component
      AppSidebarTree.tsx      ← private sub-components (grouped here due to mutual recursion)
      index.ts                ← re-exports the public API only
  ```

- Use `helpers.ts` for pure functions with no JSX.
- Use `types.ts` for types that are either exported publicly or shared across multiple files in
  the folder. Keep types that belong exclusively to one file inline in that file.
- The `index.ts` at the folder root is the public API surface. Everything not listed there is
  an implementation detail.
- Mutually recursive sub-components that cannot be separated without creating circular imports
  may co-locate in a single file named after their logical role (e.g. `AppSidebarTree.tsx`).

## Contexts, Providers, Hooks

- One context definition per file under `src/contexts/`.
- One provider per file under `src/providers/`.
- One hook per file under `src/hooks/`.
- Barrel files (`index.ts`) are the public import surface for each directory.

## Schemas and Validation

- Zod schemas live in `src/schemas/` organized by domain (e.g. `schemas/auth/`).
- Never define Zod schemas inside React components or hooks — keep them importable and reusable.
- Runtime validation happens at the API boundary. Components and context layers receive
  already-validated values.

## Pages

- One page component per file under `src/pages/`.
- Pages are thin: they compose layout, navigation components, and feature components. They do
  not own business logic or data-fetching state directly.

## Imports

- Use the `@components`, `@pages`, `@hooks`, `@contexts`, `@dto`, `@schemas`, `@helpers`, and
  `@routes` path aliases defined in `tsconfig.json`. Avoid relative `../../` paths when an alias
  covers the target.
