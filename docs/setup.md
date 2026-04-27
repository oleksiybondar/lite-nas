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
services. The currently affected service list contains `nats-server`.

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

Usage:

```bash
sudo ./.build/system-metrics-cli/linux-<arch>/system-metrics-cli
sudo ./.build/system-metrics-cli/linux-<arch>/system-metrics-cli --cpu
sudo ./.build/system-metrics-cli/linux-<arch>/system-metrics-cli --ram
sudo ./.build/system-metrics-cli/linux-<arch>/system-metrics-cli --history
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
- log path `/var/lib/lite-nas/system-metrics.log`
- systemd service `lite-nas-system-metrics.service`
- runtime user `lite-nas-system-metrics`
- runtime group `lite-nas-system-metrics`
- supplementary config group `lite-nas`

The installed systemd unit is deployed under `/etc/systemd/system/` with the
repository-managed LiteNAS defaults.

## Build Debian packages

Build the current LiteNAS package set:

```bash
./scripts/package/build-all-debs.sh --version=0.1.0+local
```

Build the native-architecture LiteNAS package directly when needed:

```bash
./scripts/package/build-lite-nas-deb.sh --version=0.1.0+local
./scripts/package/build-lite-nas-deb.sh --system-metrics-binary=/tmp/system-metrics --system-metrics-cli-binary=/tmp/system-metrics-cli
```

The package output currently contains one native-architecture package:

- `lite-nas`: bootstrap/profile package that also bundles the system metrics service and CLI binaries

The `lite-nas` package:

- depends on `libpam0g`, `zfsutils-linux`, `nats-server`, `openssl`, and `systemd`
- recommends future hardening packages such as `aide`, `clamav`, `ufw`, and `usbguard`
- shows a GPLv3 notice during installation
- explains that it applies a managed LiteNAS host profile
- asks whether LiteNAS may replace the local NATS configuration
- rotates or creates NATS certificates only when one or more expected files are missing or stale
- stores the shared CA under `/etc/lite-nas/certificates/root-ca.crt`
- stores service certificates under `/etc/lite-nas/certificates/<service>/`

The `lite-nas` package installs the auth service, system metrics service,
system metrics CLI app, and web gateway under that profile.

Install a built package with dependency resolution:

```bash
sudo ./scripts/package/install-lite-nas-deb.sh --package .build/packages/lite-nas_0.1.0_amd64.deb
```

For local testing, prefer `apt-get install <package.deb>` or the helper above
instead of `dpkg -i`. The package already declares dependencies such as
`zfsutils-linux` and `nats-server`, but `dpkg` does not resolve or install
them automatically.

Lint a built package with:

```bash
./scripts/package/lint-system-metrics-deb.sh .build/packages/lite-nas_0.1.0_amd64.deb
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
lite-nas-system-metrics-cli
lite-nas-web-gateway
lite-nas-auth-service
```

Override the list with repeated `--user` options:

```bash
sudo ./scripts/rotate-nats-certificates.sh --user lite-nas-system-metrics --user lite-nas-other-service
```

The script stores NATS server CA and server certificate material under
`/etc/nats-server/certificates`, the shared LiteNAS CA under
`/etc/lite-nas/certificates`, and service client certificate material under
`/etc/lite-nas/certificates/<service>/`.

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
./scripts/run-ci-analysis.sh
```

This calls the same analysis scripts used by GitHub Actions. It expects local
developer dependencies to already be installed with
`./scripts/install-dev-dependencies.sh`.

`./scripts/run-ci.sh` remains as a compatibility wrapper around
`./scripts/run-ci-analysis.sh`.

Run the local CI build checks:

```bash
./scripts/run-ci-build.sh
```

Run the local CI test and coverage checks:

```bash
./scripts/run-ci-test.sh
```

Run Markdown analysis:

```bash
./scripts/ci/markdown-analysis.sh
```

Run JS/TS/JSON analysis:

```bash
./scripts/ci/js-ts-analysis.sh
```

Run Go analysis:

```bash
./scripts/ci/go-analysis.sh
```

Run shell analysis:

```bash
./scripts/ci/shell-analysis.sh
```

## CI scripts

Reusable CI scripts live in `scripts/ci/`.

Analysis scripts are shared by local on-demand checks and GitHub Actions:

- `github-actions-analysis.sh`
- `markdown-analysis.sh`
- `shell-analysis.sh`
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

GitHub Actions runs separate static analysis jobs for Markdown, shell, JS/TS,
Go, and GitHub Actions workflows. Jobs explicitly pass when no matching files
or Go modules exist.

CI workflow order:

1. `Static analysis` runs on pull requests and pushes to `main` or `master`.
2. `Main pipeline` runs only after `Static analysis` completes successfully on
   any branch.
3. `Release pipeline` runs only after `Main pipeline` completes successfully
   on `main`.

`Main pipeline` uploads the built `system-metrics` binaries as short-lived
workflow artifacts together with the built `system-metrics-cli` binaries.
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

- Go duplication is enforced by `golangci-lint` using `dupl`.
- JS/TS duplication is enforced by `jscpd` in CI only.
- JS/TS duplication is not part of pre-commit hooks.
