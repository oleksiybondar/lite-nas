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

## Scope of This Stage

The first `admin-panel` slice is expected to be minimal.

Its initial value is mainly platform-completion value:

- establish the frontend app location and name
- establish the frontend build output flow
- connect browser-facing assets to package assembly
- prepare the platform for later, richer UI implementation
