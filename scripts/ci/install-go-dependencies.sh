#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t modules < <(find . -name go.mod -not -path './vendor/*')

if [ "${#modules[@]}" -eq 0 ]; then
  log.info "No Go modules found; skipping CI Go dependency installation."
  exit 0
fi

log.pushTask "Installing CI Go analysis dependencies"
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
log.popTask
