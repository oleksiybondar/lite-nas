# File Explorer — Implementation Task

## Goal

Create a small React single-page application for browsing files through the
existing File Service. Deploy it as a separately built application and expose
it inside the admin panel at `/apps/file-explorer`. File Explorer is an
admin-panel utility and is not a standalone product.

## User behavior

- Start at host `/` and use the File Service `GET /ls/{absolute-path...}`
  contract. If the user cannot list `/`, show the service error on the page.
- Show directories first and files below them. Include a `..` entry when the
  current location has a parent the user may access.
- Use a folder icon for directories. Use recognizable icons for PDFs, video,
  images, archives, and text; use a generic binary-file icon for other types.
- Double-clicking a directory navigates into it. Double-clicking a file opens
  the File Service `/f/{absolute-path...}` URL in a new browser tab so the
  service's content headers determine inline rendering or download behavior.
- Provide an Upload button in the header. Its shared modal accepts selected or
  dropped files and uses File Service's proprietary upload protocol; the
  service decides create and replace permissions.
- Use the admin panel's shared preferred UI settings when available. Keep any
  File Explorer-specific UI changes in an app-specific preference key.
- Use the existing `lite-nas-at` cookie through same-origin Nginx composition.
  Do not copy tokens into application storage, iframe messages, or URLs.
- Keep two independent panels by default, with a header control to show one or
  two panels. Highlight the active panel. A single click selects an item;
  Ctrl or Command click and Space toggle a multi-selection, and arrow keys
  move keyboard focus. Keep a Back control visible after listing errors.
- Show Copy, Move, Preview, Delete, Upload, and New folder in the bottom action
  bar. Copy and Move use the opposite panel as their destination and appear in
  two-panel mode. Copy, Move, and Delete require typing `yes` in a reusable
  confirmation dialog. Display a queue with overall and current job progress
  and expandable task statuses. Cache only job IDs and labels in local storage
  so a browser reload resumes polling server-side jobs.

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
   a separate `lite-nas-file-explorer` Debian package. Ensure
   deep links fall back to this SPA's entry point without intercepting API
   routes.
5. Include File Explorer in every `deploy-all` run and keep its Applications
   entry available in the admin panel. The app has no enable/disable setting.
6. Confirm that theme inheritance works in the same-origin embedded route and
   that opening a file preserves the intended new-tab behavior.

## Acceptance criteria

- The app is available at `/apps/file-explorer` through Nginx and can be
  embedded by the admin panel.
- `deploy-all` always installs the File Explorer assets and Nginx mount.
- Directory navigation uses `/ls/`; file activation opens the corresponding
  `/f/` resource in a new tab.
- Directory and file sorting and icon categories follow the behavior above.
- Authentication is provided by the shared secure browser cookie; no duplicate
  token storage is introduced.
- Build/deploy/package integration gives File Explorer a separate Debian
  package and installation path.
