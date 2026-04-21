#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '\n==> %s\n' "$1"
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'Missing required command: %s\n' "$1" >&2
    return 1
  fi
}

log "Checking base tooling"
require_command node
require_command npm
require_command go

log "Installing Node developer dependencies"
npm install

log "Installing Go developer tools"
go install mvdan.cc/gofumpt@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install mvdan.cc/sh/v3/cmd/shfmt@latest

if ! command -v shellcheck >/dev/null 2>&1; then
  cat <<'MSG'

shellcheck is not installed.

Debian/Ubuntu:
  sudo apt-get update && sudo apt-get install -y shellcheck

macOS:
  brew install shellcheck
MSG
else
  log "shellcheck is already installed"
fi

log "Installing Git hooks"
npx lefthook install

cat <<'MSG'

Developer tooling is installed.

Ensure your Go bin directory is on PATH, for example:
  export PATH="$(go env GOPATH)/bin:$PATH"
MSG
