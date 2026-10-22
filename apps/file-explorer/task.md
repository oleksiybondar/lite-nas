# File Explorer — Implementation Task

## Goal

Create a small React single-page application for browsing files through the
existing File Service. Deploy it as a separately built application and expose
it inside the admin panel at `/apps/file-explorer`. File Explorer is an
admin-panel utility and is not a standalone product.

## User behavior

- Start at the initial directory supplied by the existing authorized file
  browser configuration and use the File Service
  `GET /ls/{absolute-path...}` contract. If no starting-directory setting
  exists, settle its value with the File Service owner before implementation;
  do not assume that every user may list host `/`.
- Show directories first and files below them. Include a `..` entry when the
  current location has a parent the user may access.
- Use a folder icon for directories. Use recognizable icons for PDFs, video,
  images, archives, and text; use a generic binary-file icon for other types.
- Double-clicking a directory navigates into it. Double-clicking a file opens
  the File Service `/f/{absolute-path...}` URL in a new browser tab so the
  service's content headers determine inline rendering or download behavior.
- Use the admin panel's shared preferred UI settings when available. Keep any
  File Explorer-specific UI changes in an app-specific preference key.
- Use the existing `lite-nas-at` cookie through same-origin Nginx composition.
  Do not copy tokens into application storage, iframe messages, or URLs.

## Implementation instructions

1. Inspect File Service listing response fields, path encoding rules, browser
   authentication, Nginx packaging conventions, and admin-panel theme APIs.
   Do not invent or weaken the `/ls/` or `/f/` contracts.
2. Create the React app under `apps/file-explorer` using repository frontend
   validation and documentation conventions. Keep runtime schemas outside
   React components and validate values received from HTTP.
3. Implement directory navigation, parent navigation, stable sorting, loading,
   empty, and error states. Treat all names and paths from the service as
   untrusted display/input data.
4. Build static assets and add the Nginx mount at `/apps/file-explorer/` using
   the repository's deployment and single-package integration flows. Ensure
   deep links fall back to this SPA's entry point without intercepting API
   routes.
5. Add an independent enable/disable setting for the app, defaulting to
   enabled. When disabled, hide its Applications entry and do not serve its SPA
   through the admin-panel mount.
6. Confirm that theme inheritance works in the same-origin embedded route and
   that opening a file preserves the intended new-tab behavior.

## Acceptance criteria

- The app is available at `/apps/file-explorer` through Nginx and can be
  embedded by the admin panel.
- File Explorer can be disabled independently and is enabled by default.
- Directory navigation uses `/ls/`; file activation opens the corresponding
  `/f/` resource in a new tab.
- Directory and file sorting and icon categories follow the behavior above.
- Authentication is provided by the shared secure browser cookie; no duplicate
  token storage is introduced.
- Build/deploy/package integration follows existing scripts and the single
  Debian package workflow.
