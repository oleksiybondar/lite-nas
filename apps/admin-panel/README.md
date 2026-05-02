# Admin Panel

## Purpose

`admin-panel` is the LiteNAS browser application.

It is expected to be implemented as a Vite + TypeScript web app and to
be packaged into static frontend assets served by `web-gateway`.

## Role in the Platform

The app is intentionally separate from the gateway service:

- the app owns browser-side UI behavior
- `web-gateway` owns static asset serving and browser-facing API adaptation

This separation keeps the frontend build flow explicit and allows the
packaging process to copy built assets into the gateway-owned static
asset area.

## Current Slice

The first `admin-panel` slice is intentionally minimal, but the browser app
skeleton is now wired into the platform build, deployment, and package flow.

It establishes:

- Vite + React + TypeScript app wiring
- app-local providers, routes, hooks, contexts, and theme modules
- path aliases matching the frontend template style
- dark-mode default theme behavior
- stable build output under `.build/admin-panel` for gateway-owned static assets:
  - `.build/admin-panel/index.html`
  - `.build/admin-panel/assets/index.css`
  - `.build/admin-panel/assets/index.js`
- deployment/package handoff into `/usr/share/lite-nas/web-gateway/assets`

This gives future feature branches a ready frontend shell instead of another
round of project bootstrapping.

## Auth Boundary

The app uses `web-gateway` as a BFF. Access-token and refresh-token values are
transported through HTTP-only cookies and are not readable by browser
JavaScript.

Frontend auth state should be derived from gateway responses:

- call `/api/auth/me` to detect an existing access-token-backed session
- on `401`, call `/api/auth/refresh` with credentials included
- retry `/api/auth/me` after a successful refresh
- treat refresh failure as an anonymous session

The app should not store token values in local storage, session storage, React
state, or query caches.

The Vite dev server proxies `/api` to the configured gateway target. All other
development-server paths remain available to the SPA history fallback.

## Local Commands

Run commands from this directory or with `npm --prefix apps/admin-panel ...`
from the repository root.

```sh
npm install
npm run dev
npm run build
npm run test:unit
```

From the repository root, use the platform build wrapper when producing assets
for deployment or package assembly:

```sh
./scripts/build-admin-panel.sh
```
