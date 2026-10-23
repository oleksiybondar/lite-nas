# File Explorer

The application contract is documented in [requirements.md](requirements.md).

This is the separately built React app mounted at `/apps/file-explorer/` by
Nginx. It uses `@lite-nas/theme` and the admin panel's browser theme preference.
It starts at `/`, loads direct children from File Service's `/ls/` route,
opens files through `/f/` in a new tab, and uploads to the current directory
through the shared `@lite-nas/file-upload` dialog. A denied root listing is
shown as a File Service error. File Service enforces read, create, and replace
permissions.

Directory entries appear in a horizontally scrollable details grid with size,
type, octal permissions, owner, group, and local last-modified time. Size
formatting comes from the reusable `@lite-nas/formatting` package. The Show
hidden files control starts enabled and sits beside the one/two-panel switch;
turning it off hides dot-prefixed entries from both panels and their actions.
Preview supports images, PDFs, browser audio/video controls, and escaped text
for supported markup, data, and source formats, with syntax highlighting for
recognized grammars. Text previews never execute file contents.

The shared `@lite-nas/formatting` package exports `formatByteSize(size)` for
nullable byte counts. It uses 1024-byte unit steps, displays B/KB/MB/GB/TB, and
rounds scaled values to at most one decimal place. Null is displayed as an em
dash. UI apps can import it from `@lite-nas/formatting` to keep size values
consistent.

The admin panel embeds this Nginx-served app at `/applications/file-explorer`
and always shows its Applications entry.

From the repository root, install Node dependencies with `npm install`, then
run `./scripts/build-file-explorer.sh`. The static output goes to
`.build/file-explorer`. Use `./scripts/package/build-file-explorer-deb.sh` to
build the standalone `lite-nas-file-explorer` Debian package. Both commands
accept `--output-dir=PATH`; the package command also accepts
`--assets-dir=PATH` and `--version=VERSION`.

The package installs assets and its Nginx snippet in
`/usr/share/lite-nas/file-explorer`. An active link in
`/usr/share/lite-nas/nginx` exposes the app to the LiteNAS Nginx site.
The main `lite-nas` Debian package requires the matching File Explorer package.
Install them together through the normal Debian package manager; File
Explorer's maintainer scripts validate and reload Nginx. For local deployment,
`./scripts/deploy-all.sh` always builds and installs the assets and snippet
before it refreshes Nginx. Pass `--file-explorer-assets PATH` to use a prebuilt
Vite output directory.

The app uses the same origin as the admin panel, so File Service calls use the
existing browser cookie without copying it into app storage. The web gateway's
login route issues that cookie; if the gateway is unavailable, new logins are
unavailable, while an existing valid cookie can still authorize File Service
requests.
