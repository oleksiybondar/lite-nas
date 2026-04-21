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
git nodejs npm golang-go shellcheck
```

Run the script as your normal user when possible. If you run it with
`sudo bash scripts/install-dev-dependencies.sh`, system packages are installed
as root and repo-local installs are run as the original sudo user.

On macOS, install `shellcheck` with:

```bash
brew install node go shellcheck
```

## Git hooks

Install hooks manually when needed:

```bash
npx lefthook install
```

The pre-commit hook runs only on staged files. It auto-fixes when possible,
re-stages modified files, and fails the commit when issues remain.

## Manual linting

Run Markdown analysis:

```bash
npx markdownlint-cli2
```

Run JS/TS/JSON analysis:

```bash
npx biome ci .
```

Run JS/TS duplication detection:

```bash
npx jscpd .
```

Run Go analysis from a Go module directory:

```bash
golangci-lint run
```

Run shell analysis:

```bash
shellcheck path/to/script.sh
shfmt -d path/to/script.sh
```

## CI static analysis

GitHub Actions runs separate static analysis jobs for Markdown, shell, JS/TS,
and Go. Jobs explicitly pass when no matching files or Go modules exist.

Duplication policy:

- Go duplication is enforced by `golangci-lint` using `dupl`.
- JS/TS duplication is enforced by `jscpd` in CI only.
- JS/TS duplication is not part of pre-commit hooks.
