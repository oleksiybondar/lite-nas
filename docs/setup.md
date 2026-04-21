# Developer setup

This repository keeps shared developer tooling configuration in `dev-configs/`.
Thin root-level config files are present so tools can discover their configs
without extra command-line flags.

## Required tools

- Node.js and npm
- Go
- Git
- `shellcheck`
- Go-installed tools:
  - `gofumpt`
  - `goimports`
  - `golangci-lint`
  - `shfmt`
- Node-installed tools:
  - Biome
  - Lefthook
  - jscpd
  - markdownlint-cli2

## Install developer dependencies

Run:

```bash
./scripts/install-dev-dependencies.sh
```

The script is safe to re-run. It installs Node dependencies, Go tools, and
Lefthook Git hooks. On Debian/Ubuntu, it also installs missing base packages
with `apt-get`:

```bash
git nodejs npm golang-go shellcheck shfmt
```

Run the script as your normal user when possible. If you run it with
`sudo bash scripts/install-dev-dependencies.sh`, system packages are installed
as root and repo-local installs are run as the original sudo user.

On macOS, install base tools with:

```bash
brew install node go shellcheck shfmt
```

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
./scripts/run-ci.sh
```

This calls the same analysis scripts used by GitHub Actions. It expects local
developer dependencies to already be installed with
`./scripts/install-dev-dependencies.sh`.

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

- `markdown-analysis.sh`
- `shell-analysis.sh`
- `js-ts-analysis.sh`
- `go-analysis.sh`

CI-specific dependency setup scripts are separate from local developer setup:

- `install-node-dependencies.sh`
- `install-shell-dependencies.sh`
- `install-go-dependencies.sh`

Shell scripts share logging helpers from `scripts/helpers/logger.sh`. Source
that helper relative to the script file, not the current working directory, so
scripts work from any launch path.

## CI static analysis

GitHub Actions runs separate static analysis jobs for Markdown, shell, JS/TS,
and Go. Jobs explicitly pass when no matching files or Go modules exist.

CI workflow order:

1. `Static analysis` runs on pull requests and pushes to `main` or `master`.
2. `Main pipeline` runs only after `Static analysis` completes successfully.
3. `Release pipeline` runs only after `Main pipeline` completes successfully
   on `main`.

`Main pipeline` currently contains a manual approval gate. Configure the GitHub
environment `main-pipeline-approval` with required reviewers in repository
settings to make GitHub pause the job for approval. Without environment
protection rules, GitHub will run the stub without pausing.

Duplication policy:

- Go duplication is enforced by `golangci-lint` using `dupl`.
- JS/TS duplication is enforced by `jscpd` in CI only.
- JS/TS duplication is not part of pre-commit hooks.
