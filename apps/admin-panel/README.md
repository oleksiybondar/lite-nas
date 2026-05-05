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

## Architecture Notes

- [API request architecture](docs/api-request-architecture.md) documents the
  intended split between request lifecycle builders, fetch transport behavior,
  and app-facing API actions.

## Local Commands

Run commands from this directory or with `npm --prefix apps/admin-panel ...`
from the repository root.

```sh
npm install
npm run dev
npm run build
npm run test:unit
```

During `npm run dev`, Vite proxies `/api` to the local web gateway at
`http://127.0.0.1:9090` while preserving the `/api/...` path. Override the
gateway origin with `LITE_NAS_WEB_GATEWAY_ORIGIN` when using a different local
address.

From the repository root, use the platform build wrapper when producing assets
for deployment or package assembly:

```sh
./scripts/build-admin-panel.sh
```
