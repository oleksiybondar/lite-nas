# File Explorer — Software Requirements

Requirement IDs use `file-explorer-app/<TYPE>-<NNN>`. File Explorer is a React
application served as static assets by Nginx at `/apps/file-explorer/`. It uses
File Service for filesystem access and the existing admin-panel login cookie;
it has no independent login or enable/disable setting.

## Functional Requirements

### file-explorer-app/FR-001 Browse authorized directories

The app MUST start each panel at host `/` and request direct children from
`GET /ls/{absolute-path...}`. It MUST show directories before files, sort
names consistently, and provide a parent-directory entry outside `/`. It MUST
validate each returned object's absolute `path`, child `name`, `kind`, and
nullable `size`, `owner`, `group`, `permissions`, and `last_modified` metadata. It MUST
show loading, empty, and File Service error states. A denied listing at `/`
MUST show the error; it MUST NOT silently choose another starting directory.
The Back control MUST remain available after a failed listing when panel
history contains an earlier directory.

Each entry MUST be presented in a horizontally scrollable details grid with
Name, Size, File type, Permissions, Owner, Group, and Last modified columns.
Column headers MUST remain aligned with their entry values while scrolling.
Missing metadata MUST display as an em dash. Size MUST use the shared
`@lite-nas/formatting` byte-size formatter, using binary 1024-byte units and
the labels B, KB, MB, GB, and TB, with at most one fractional digit.
Permissions MUST be displayed as a four-digit octal mode, including special
permission bits when present. Last modified values MUST use the browser's
local time zone.

### file-explorer-app/FR-002 Present one or two independent panels

The default view MUST show one panel and hide hidden files. A header control
MUST switch between one-panel and two-panel views. Both panels MUST remain
independently navigable, and the active panel MUST be highlighted. Switching
views MUST preserve both panels' directory and history. File operations MUST
use the active panel's selection. Copy and Move MUST appear only in two-panel
view and use the other panel's current directory as the target.

### file-explorer-app/FR-003 Select and inspect entries

A single click MUST select one file or directory. Ctrl-click or Command-click
and Space MUST toggle items in a multi-selection; arrow keys MUST navigate
entry focus. Each row MUST show an icon, name, and type detail. The icon set
MUST distinguish folders, PDFs, common image formats, audio, video, archives,
structured data, markup, code, and other files. Double-clicking a directory
MUST enter it. Double-clicking a file MUST open its `/f/{absolute-path...}`
resource in a new tab. Preview MUST show supported images and PDFs inline,
provide native HTML audio or video controls for supported media, and display
supported text, markup, data, and source files as escaped content with an
extension label. Files with a supported syntax grammar MUST show syntax
highlighting. Text previews MUST NOT execute markup or source code.
Unsupported file types MUST NOT offer Preview. The preview dialog MUST include
a visible close control and use the large dialog width.

Names beginning with `.` MUST be treated as hidden files or folders. The
workspace MUST provide a Show hidden files toggle beside the panel-view toggle.
The toggle MUST use distinct visible and crossed-out eye states. Show hidden
files MUST be enabled by default; disabling it hides dot-prefixed entries from
both panels and removes hidden entries from the active selection and its
available actions.

### file-explorer-app/FR-004 Upload through the shared dialog

The header and bottom action bar MUST provide Upload for the active directory.
The dialog MUST be the shared `@lite-nas/file-upload` component so the admin
panel and Moments can use the same interaction. It MUST accept selected files
and drag-and-drop files, submit them through File Service's proprietary upload
protocol, and show upload progress and service errors. File Service MUST decide
create and replace permissions. The dialog MUST default to "Readable by other
users" and send the inverse as a validated `private` boolean. The initial
chunk-size map based on original
file size MUST be: up to 8 MiB, 1 MiB chunks; above 8 MiB through 256 MiB,
4 MiB chunks; above 256 MiB through 2 GiB, 8 MiB chunks; and above 2 GiB,
16 MiB chunks. The final chunk may be shorter.

### file-explorer-app/FR-005 Create folders

New folder MUST request a validated single child name in the active panel
and send `POST /f/mkdir` with an absolute path and validated `private` boolean.
The dialog MUST default to allowing other users access and allow a private folder.
A configured known-path mask may select traverse-only mode for a non-private
request. A successful
request MUST refresh the directory listing. Invalid names, including path
separators and `.` or `..`, MUST be rejected before submission.

### file-explorer-app/FR-006 Copy and move selections

Copy and Move MUST submit one task per selected entry, using the entry's name
under the opposite panel's directory as the complete destination path. They
MUST call `POST /f/copy` or `POST /f/move` with `{source,destination}` and
allow the server to merge existing directories and replace existing regular
files. The UI MUST present a confirmation dialog before submitting either
action.

### file-explorer-app/FR-007 Delete selections

Delete MUST submit one task per selected entry through
`DELETE /f/{absolute-path...}` after confirmation. It MUST handle immediate
`204 No Content` for regular files and `202 Accepted` jobs for directories.
A successful deletion MUST refresh affected listings.

### file-explorer-app/FR-008 Require typed confirmation

Copy, Move, and Delete MUST use a reusable confirmation dialog. The dialog
MUST identify the action and selected item count, and MUST enable submission
only when the user enters the exact lowercase text `yes`. The component MUST
have an action-independent interface so it can move into a shared TypeScript
package without changing its interaction contract.

### file-explorer-app/FR-009 Show and resume the action queue

The app MUST show pending, running, done, and error states for submitted
actions. It MUST show overall queue progress, current operation progress, and
an expandable list of task statuses. Server jobs MUST be polled at
`GET /f/jobs/{id}` for server-reported percentage, bytes, and item counts.
The browser MUST cache only job IDs and display labels in local storage so it
can resume polling after reload while File Service remains running. A missing
server job after service restart MUST be shown as an error. An immediate
`204` file deletion is complete without a server job.

## Interface and Security Requirements

### file-explorer-app/IR-001 Validate File Service data and requests

The app MUST validate directory listings, job responses, host paths, new
folder names, and submitted operation payloads at runtime with Zod outside
React components where practical. File Service errors MUST be shown without
assuming a successful listing or mutation. Client validation MUST NOT replace
File Service authorization.

### file-explorer-app/SR-001 Use the shared browser session

Requests MUST use the same-origin `lite-nas-at` cookie issued by the existing
login flow. The app MUST NOT copy access tokens into JavaScript storage,
iframe messages, or URLs. New login depends on the admin-panel login flow;
an existing valid cookie may still authorize File Service requests when that
flow is unavailable.

## Integration and Operational Requirements

### file-explorer-app/IR-002 Share the admin-panel theme

The app MUST use Material UI and the shared `@lite-nas/theme` package. It MUST
read the admin panel's preferred theme from the same-origin preference key and
keep File Explorer-specific UI preferences separate from shared theme data.

### file-explorer-app/IR-003 Reuse byte-size formatting

The app MUST format file sizes through the shared `@lite-nas/formatting`
package so other browser UI apps can use the same byte-size display contract.
The formatter MUST accept nullable byte counts, display null as an em dash,
and use binary unit steps with B, KB, MB, GB, and TB labels.

### file-explorer-app/OR-001 Build and serve the static app

The app MUST have a local build script producing static assets and an Nginx
mount at `/apps/file-explorer/` with SPA deep-link fallback. The admin panel
MUST expose an Applications entry and embed the Nginx-served app at
`/applications/file-explorer`. `deploy-all` MUST always deploy it. The
existing local package workflow MUST build the separately requested
`lite-nas-file-explorer` Debian package containing the assets and Nginx
snippet; the app has no deployment enable/disable flag. File Service requests
MUST pass directly through Nginx's `/f/` and `/ls/` routes, without the web
gateway proxying the static app.

## Testability Requirements

### file-explorer-app/TR-001 Verify user workflows

The UI MUST support verification of denied root listing, failed-listing Back
navigation, independent panel histories, view switching, active selection,
keyboard multi-selection, metadata grid values and horizontal scrolling,
hidden-file visibility and selection behavior, file opening, image/PDF/media/
text previews, escaped text rendering, folder creation, upload, confirmation
gating and preview dismissal, copy and move destination paths, file and
directory delete responses,
queue progress, and job restoration after browser reload. Size formatting MUST
be covered for null, byte, and supported unit boundaries. Build and deployment
checks MUST verify the Nginx mount, admin-panel embed, shared theme, shared
formatting package, and deployed static assets.
