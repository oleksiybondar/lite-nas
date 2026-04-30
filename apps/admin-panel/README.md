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

The first `admin-panel` slice is intentionally minimal.

It establishes:

- Vite + React + TypeScript app wiring
- app-local providers, routes, hooks, contexts, and theme modules
- path aliases matching the frontend template style
- dark-mode default theme behavior
- stable build output names for gateway-owned static assets:
  - `dist/assets/index.css`
  - `dist/assets/index.js`

## Local Commands

Run commands from this directory or with `npm --prefix apps/admin-panel ...`
from the repository root.

```sh
npm install
npm run dev
npm run build
npm run test:unit
```
