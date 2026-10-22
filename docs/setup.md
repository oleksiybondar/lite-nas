# Developer setup

This repository keeps shared developer tooling configuration in `dev-configs/`.
Thin root-level config files are present so tools can discover their configs
without extra command-line flags.

## Required tools

### Developer tools

- Node.js and npm
- Go
- Git
- `shellcheck`
- Go-installed tools:
  - `actionlint`
  - `gofumpt`
  - `goimports`
  - `golangci-lint`
  - `shfmt`
- Node-installed tools:
  - Biome
  - Lefthook
  - jscpd
  - markdownlint-cli2

### Runtime dependencies

- NATS Server
- OpenSSL

## Install developer dependencies

Run:

```bash
sudo ./scripts/install-dev-dependencies.sh
```

The script is safe to re-run. It installs Node dependencies, Go tools, and
Lefthook Git hooks. On Debian/Ubuntu, it also installs missing base packages
with `apt-get`:

```bash
git nodejs npm golang-go shellcheck shfmt
```

Run the script with `sudo`. System packages are installed as root and
repo-local installs are run as the original sudo user.

Go-based developer tools are installed into the repo-local `.bin/` directory so
Git hooks and scripts can find consistent tool versions without relying on a
global Go bin path.

On macOS, install base tools with:

```bash
brew install node go shellcheck shfmt actionlint
```

## Install runtime dependencies

Run:

```bash
sudo ./scripts/install-runtime-dependencies.sh
```

The script is safe to re-run. On Debian/Ubuntu, it installs missing runtime
packages with `apt-get`:

```bash
nats-server openssl
```

On macOS, install runtime dependencies with:

```bash
brew install nats-server openssl
```

## Deploy runtime configs

Deploy repository-managed files from `configs/etc` into `/etc`:

```bash
sudo ./scripts/deploy-configs.sh
```

The deployment script overwrites matching files under `/etc`, normalizes
permissions for LiteNAS-managed NATS config paths, and restarts affected
services. The currently affected service list contains `nats-server`. This also
deploys the LiteNAS logrotate policy at `/etc/logrotate.d/lite-nas`, including
the File Service and Stream Service logs.

For validation without touching `/etc`, deploy to another target directory and
skip service restarts:

```bash
sudo ./scripts/deploy-configs.sh --target-dir /tmp/lite-nas-etc --no-restart
```

## Build the system-metrics binary

Build an installable Linux binary artifact for the local machine architecture:

```bash
./scripts/build-system-metrics-binary.sh
```

Override the output path when needed:

```bash
./scripts/build-system-metrics-binary.sh --output=/tmp/system-metrics
```

The default output path is:

```text
.build/system-metrics/linux-<arch>/system-metrics
```

## Build the system-metrics CLI binary

Build the read-only NATS client used to request the latest snapshot or the
snapshot history for the local machine architecture:

```bash
./scripts/build-system-metrics-cli-binary.sh
```

Override the output path when needed:

```bash
./scripts/build-system-metrics-cli-binary.sh --output=/tmp/system-metrics-cli
```

The default output path is:

```text
.build/system-metrics-cli/linux-<arch>/system-metrics-cli
```

The default client config template is deployed to:

```text
/etc/lite-nas/system-metrics-cli.conf
```

The deployed CLI config and `lite-nas-system-metrics-cli` transport client
certificate are readable by local users so the CLI can be run directly from a
normal shell. The deployment also initializes:

```text
/var/log/lite-nas/system-metrics-cli.log
```

Usage:

```bash
./.build/system-metrics-cli/linux-<arch>/system-metrics-cli
./.build/system-metrics-cli/linux-<arch>/system-metrics-cli --cpu
./.build/system-metrics-cli/linux-<arch>/system-metrics-cli --ram
./.build/system-metrics-cli/linux-<arch>/system-metrics-cli --history
```

## Deploy the system-metrics CLI app

Install the binary, deploy the client config, and prepare the LiteNAS
bootstrap prerequisites:

```bash
sudo ./scripts/deploy-system-metrics-cli.sh
```

Install an already-built binary instead of building during deployment:

```bash
sudo ./scripts/deploy-system-metrics-cli.sh --binary /tmp/system-metrics-cli
```

## Deploy the system-metrics service

Install the binary, deploy the service config, prepare the service account and
log file, install a hardened systemd unit, and start the service:

```bash
sudo ./scripts/deploy-system-metrics.sh
```

Install an already-built binary instead of building during deployment:

```bash
sudo ./scripts/deploy-system-metrics.sh --binary /tmp/system-metrics
```

To validate file installation and unit deployment without starting the service:

```bash
sudo ./scripts/deploy-system-metrics.sh --no-start
```

By default, the deployment uses:

- binary path `/usr/libexec/lite-nas/system-metrics`
- config path `/etc/lite-nas/system-metrics.conf`
- log path `/var/log/lite-nas/system-metrics.log`
- systemd service `lite-nas-system-metrics.service`
- runtime user `lite-nas-system-metrics`
- runtime group `lite-nas-system-metrics`
- supplementary config group `lite-nas`

The installed systemd unit is deployed under `/etc/systemd/system/` with the
repository-managed LiteNAS defaults.

## Build Debian packages

Build the current LiteNAS package set:

```bash
./scripts/package/build-all-debs.sh --version=0.2.0~alpha.1
```

Build the native-architecture LiteNAS package directly when needed:

```bash
./scripts/package/build-lite-nas-deb.sh --version=0.2.0~alpha.1
./scripts/package/build-lite-nas-deb.sh --system-metrics-binary=/tmp/system-metrics --resources-monitor-binary=/tmp/resources-monitor
```

Build only the browser app assets:

```bash
./scripts/build-admin-panel.sh
```

By default this writes the Vite build output to `.build/admin-panel`. The
gateway deploy and package flows normalize that output into the flat static
asset layout served from `/usr/share/lite-nas/web-gateway/assets`.

The package output currently contains one native-architecture package:

- `lite-nas`: bootstrap/profile package that bundles the currently packaged
  LiteNAS services, CLIs, and web assets

The `lite-nas` package:

- depends on `aide`, `libpam0g`, `zfsutils-linux`, `nats-server`, `nginx`,
  `openssl`, `systemd`, and `ufw`
- installs the AIDE binary as a platform dependency, but LiteNAS does not
  initialize an AIDE database or use the distribution default AIDE config during
  package installation
- recommends future hardening packages such as `clamav` and `usbguard`
- shows a GPLv3 notice during installation
- explains that it applies a managed LiteNAS host profile
- asks whether LiteNAS may replace the local NATS configuration
- rotates or creates NATS certificates only when one or more expected files are missing or stale
- rotates or creates the auth token signing certificate when it is missing
- stores the shared transport CA under `/etc/lite-nas/certificates/transport/root-ca.crt`
- stores service NATS client certificates under
  `/etc/lite-nas/certificates/transport/<service>/`
- stores auth token signing material under `/etc/lite-nas/certificates/auth/`
- grants the `users` group read-only access to the packaged CLI configs and
  transport client certificates intended for normal-shell usage

The `lite-nas` package currently installs:

- `auth-service`
- `rbac-service`
- `system-logging-manager`
- `security-logging-manager`
- `system-metrics`
- `resources-monitor`
- `system-logging-manager-cli`
- `security-logging-manager-cli`
- `system-metrics-cli`
- `lite-nas-file-writer`
- `web-gateway`
- packaged `admin-panel` assets

`zfs-metrics` and `zfs-metrics-cli` already exist in the repository and local
deploy flows, but they are not yet part of the packaged runtime path described
in this section.

The File Service and Stream Service are currently specified in
`services/file-service/` and `services/stream-service/` and are not yet
installed. The package includes `lite-nas-file-writer` for user-scoped upload
stage creation, bounded writes, inspection, checksum, truncation, publication,
and cleanup. Build it locally with `./scripts/build-file-writer-binary.sh`,
or deploy it with `sudo ./scripts/deploy-file-writer.sh`. Its command contract,
requirements, and remaining File Service integration work are documented in
`apps/file-writer/README.md`, `apps/file-writer/requirements.md`, and
`apps/file-writer/task.md`.

The planned services will listen on separate loopback-only HTTP(S) ports.
Nginx will expose File Service at `/f/` for content/file operations and `/d/`
for explicit downloads, and Stream Service at `/s/`, while internal service
calls continue to use NATS. For example,
`/f/home/alice/photo.jpg` and `/d/home/alice/photo.jpg` map to
`/home/alice/photo.jpg`, and `/s/media/movie.avi` maps to `/media/movie.avi`;
RBAC checks the resolved host paths. File Service will also provide public
`GET /ls/{absolute-path...}`, separately from the `/f/` file-operation mount.
For example, `/ls/abs/path?recursive&files-only` lists RBAC-authorized files
below host `/abs/path`. Successful responses contain JSON `files` and
`directories` arrays of absolute host paths; `files-only` returns an empty
`directories` array. Errors use a non-success HTTP status and JSON with an
`error` code.

File Service will also support `DELETE /f/{absolute-path...}` for authorized
regular files. It checks an explicit RBAC delete decision and rejects targets
inside protected system paths before changing the filesystem. Installed
protected roots include `/etc` and platform-critical runtime directories; the
`filesystem.protected_paths` setting lets operators add protected trees.
Built-in OS/runtime roots cannot be removed through configuration. Directory
and symbolic-link deletion are outside the initial scope.

File downloads use standard HTTP byte ranges. A regular browser download can
resume or seek when its client uses `Range`; the future file-explorer download
manager can fetch separate ranges concurrently (16 per file by default), then
reassemble them locally. That accelerated mode is implemented in the client;
the service simply answers ordinary `HEAD` and `GET` requests. Strong ETags
keep all segments tied to the same file version if an upload replaces it while
the download is running. `/f/` uses inline headers for recognized PDFs/images
and attachment headers for other content; `/d/` always uses
`application/octet-stream` with attachment. Neither route redirects.

The planned File Service supports ordinary browser/form uploads at `/f/upload`
with `multipart/form-data` and a separate resumable protocol for the future
LiteNAS frontend app under `/f/proprietary/upload`. Resumable chunks default to
1 MiB and can be configured from 1 MiB through 16 MiB; the effective size is
saved per upload.
Concurrency is limited by a 60 MiB chunk-buffer budget in a 64 MiB tmpfs, with
4 MiB reserved for metadata working files. A failed chunk resumes from the last
committed offset; when buffers are full the service returns a retryable on-hold
response and the app probes `/f/proprietary/upload-window`. Both upload modes
create and write a user-owned `<destination>.upload` file in the destination
directory through `lite-nas-file-writer`, invoked as the verified Unix user
through restricted noninteractive sudo. Exclusive creation by the writer is
the Unix permission check; no preliminary `touch` is needed. After size and
SHA-256 verification and a fresh RBAC decision, the writer attempts the
final rename as the verified Unix user. A successful upload atomically replaces
an existing regular file; a failed rename leaves the old file intact and
triggers staged-file cleanup. Uploads and deletes are denied under built-in
protected OS/runtime paths and
operator-added `filesystem.protected_paths`.

The File and Stream Services run non-root. `lite-nas-fs-rd` grants file read
and directory read plus search/traverse in provisioned file-backed storage
roots, with no write permission. Neither service has a general host write
group. System-critical paths are excluded, and existing content and new mounts
need read/search group/ACL provisioning. RBAC still decides each requesting
user's potential access. The user owns the completed file because the writer
creates and writes the stage under that identity; its group, mode, and ACL
become the final file's metadata even on overwrite. The previous file's
permissions and ACL are not copied. The writer has a separate enforce-mode
AppArmor profile, and deployment must verify its confinement across sudo.

The service also exposes its authenticated Swagger UI at `/f/docs` and its
OpenAPI document at `/f/openapi.json`. The `api_docs.enabled` setting can turn
both off. Development and test configurations enable these routes; production
configuration disables them.

In the planned browser flow, both mounts use the existing `lite-nas-at` access
cookie on the same HTTPS origin. The cookie is `Secure`, `HttpOnly`,
`SameSite=Lax`, and scoped to `/`; Nginx forwards it to the matching service,
which validates the token itself.

Install a built package with dependency resolution:

```bash
sudo ./scripts/package/install-lite-nas-deb.sh --package .build/packages/lite-nas_0.2.0_amd64.deb
```

For local testing, prefer `apt-get install <package.deb>` or the helper above
instead of `dpkg -i`. The package already declares dependencies such as
`zfsutils-linux` and `nats-server`, but `dpkg` does not resolve or install
them automatically.

Lint a built package with:

```bash
./scripts/package/lint-lite-nas-deb.sh .build/packages/lite-nas_0.2.0_amd64.deb
```

## Rotate NATS certificates

Create or reuse the NATS root CA, rotate the NATS server certificate, rotate
client certificates, normalize certificate permissions, and restart NATS:

```bash
sudo ./scripts/rotate-nats-certificates.sh
```

By default, client certificates are generated for:

```text
lite-nas-system-metrics
lite-nas-zfs-metrics
lite-nas-zfs-metrics-cli
lite-nas-system-metrics-cli
lite-nas-sys-log-mgr
lite-nas-sec-log-mgr
lite-nas-sys-log-mgr-cli
lite-nas-sec-log-mgr-cli
lite-nas-web-gateway
lite-nas-auth-service
lite-nas-resources-monitor
lite-nas-rbac-service
```

Override the list with repeated `--user` options:

```bash
sudo ./scripts/rotate-nats-certificates.sh --user lite-nas-system-metrics --user lite-nas-other-service
```

The script stores NATS server CA and server certificate material under
`/etc/nats-server/certificates`, publishes the shared transport CA to
`/etc/lite-nas/certificates/transport/root-ca.crt`, and stores service client
certificate material under `/etc/lite-nas/certificates/transport/<service>/`.

## Rotate auth token certificates

Create or reuse the Ed25519 JWT signing key and certificate used by the auth
service and web gateway:

```bash
sudo ./scripts/rotate-auth-token-certificates.sh --if-missing
```

The script stores `token-signing.key` and `token-signing.crt` under
`/etc/lite-nas/certificates/auth/`. The private key is readable only by root;
the certificate is readable by the shared LiteNAS group so gateway services can
verify access tokens.

## Git hooks

Install hooks manually when needed:

```bash
npx lefthook install
```

The pre-commit hook runs only on staged files. It auto-fixes when possible,
re-stages modified files, and fails the commit when issues remain.

## Manual formatting

Run all available formatters and safe autofixes:

```bash
./scripts/format/all.sh
```

Run individual formatters:

```bash
./scripts/format/markdown.sh
./scripts/format/js-ts.sh
./scripts/format/go.sh
./scripts/format/shell.sh
```

Markdown formatting uses `markdownlint-cli2 --fix`. It fixes only rules that
markdownlint can safely autofix; remaining findings still require manual edits.

## Manual linting

Run the full local CI static analysis suite:

```bash
./scripts/run-ci-checks.sh
```

This calls the same analysis scripts used by GitHub Actions. It expects local
developer dependencies to already be installed with
`./scripts/install-dev-dependencies.sh`.

Static analysis jobs validate repository-owned source files, templates, and
configuration files. They must not depend on LiteNAS already being installed on
the developer machine or CI runner. Checks that need real installed paths such
as `/usr/libexec/lite-nas/*`, `/etc/lite-nas/*`, or running services belong in
package install validation, post-install validation, or top-level system tests.

`./scripts/run-ci.sh` remains as a compatibility wrapper around
`./scripts/run-ci-checks.sh`.

Run the local CI build checks:

```bash
./scripts/run-ci-component-build-checks.sh
```

Run the local CI test and coverage checks:

```bash
./scripts/run-ci-test.sh
```

This includes admin-panel unit tests through `scripts/test-admin-panel.sh`.

Run Markdown analysis:

```bash
./scripts/ci/markdown-analysis.sh
```

Run JS/TS/JSON analysis:

```bash
./scripts/ci/js-ts-analysis.sh
```

Build or test the browser app directly:

```bash
./scripts/build-admin-panel.sh
./scripts/test-admin-panel.sh
```

Run Go analysis:

```bash
./scripts/ci/go-analysis.sh
```

Run shell analysis:

```bash
./scripts/ci/shell-analysis.sh
```

Run repo-wide Go duplication analysis:

```bash
./scripts/ci/go-duplication-analysis.sh
```

Run repo-wide shell duplication analysis:

```bash
./scripts/ci/bash-duplication-analysis.sh
```

## CI scripts

Reusable CI scripts live in `scripts/ci/`.

Analysis scripts are shared by local on-demand checks and GitHub Actions:

- `github-actions-analysis.sh`
- `markdown-analysis.sh`
- `shell-analysis.sh`
- `go-duplication-analysis.sh`
- `bash-duplication-analysis.sh`
- `js-ts-analysis.sh`
- `go-analysis.sh`

CI-specific dependency setup scripts are separate from local developer setup:

- `install-node-dependencies.sh`
- `install-github-actions-dependencies.sh`
- `install-shell-dependencies.sh`
- `install-go-dependencies.sh`
- `install-debian-packaging-dependencies.sh`
- `package-analysis.sh`

Shell scripts share logging helpers from `scripts/helpers/logger.sh`. Source
that helper relative to the script file, not the current working directory, so
scripts work from any launch path.

## CI static analysis

GitHub Actions runs separate static analysis jobs for Markdown, shell, Go
duplication, shell duplication, JS/TS, Go, and GitHub Actions workflows. Jobs
explicitly pass when no matching files or Go modules exist.

PR validation and the main package gate also run real JS/TS build and unit test
jobs for the admin-panel application.

CI workflow order:

1. `Static analysis` runs on pull requests and pushes to `main` or `master`.
2. `Main pipeline` runs only after `Static analysis` completes successfully on
   any branch.
3. `Release pipeline` runs only after `Main pipeline` completes successfully
   on `main`.

`Main pipeline` uploads the package-build inputs as short-lived workflow
artifacts, including the built `auth-service`, `rbac-service`,
`system-logging-manager`, `security-logging-manager`, `system-metrics`,
`system-logging-manager-cli`, `security-logging-manager-cli`,
`system-metrics-cli`, `web-gateway`, `resources-monitor`, and `admin-panel`
frontend assets. The downstream package job consumes those artifacts when
assembling the `.deb`.
GitHub Actions only supports whole-day retention values, so the workflow uses
the minimum supported retention of 1 day.

`Release pipeline` downloads those upstream binary artifacts by name, builds
the architecture-specific `lite-nas` package, validates that the package can be
installed in an Ubuntu container for the target architecture, and only then
uploads the `.deb` artifact.

GitHub only evaluates `workflow_run` workflows that already exist on the
default branch. When these workflow files are introduced for the first time in a
pull request, only `Static analysis` may appear until the PR is merged. After
that initial merge, future branches will use the full chained workflow order.

`Main pipeline` currently contains a manual approval gate. Configure the GitHub
environment `main-pipeline-approval` with required reviewers in repository
settings to make GitHub pause the job for approval. Without environment
protection rules, GitHub will run the stub without pausing.

Duplication policy:

- Repo-wide Go duplication is enforced by `jscpd`, which scans the repository
  as a single tree rather than respecting `go.mod` boundaries.
- Repo-wide shell duplication is enforced by `jscpd`.
- Go duplication is also enforced inside individual Go modules by
  `golangci-lint` using `dupl`.
- JS/TS duplication is enforced by `jscpd` in CI only.
- JS/TS duplication is not part of pre-commit hooks.
