#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f -name '*.go' -not -path './vendor/*')

if [ "${#files[@]}" -eq 0 ]; then
  log.info "No Go files found."
  exit 0
fi

log.pushTask "Running gofumpt"
gofumpt -w "${files[@]}"
log.popTask

log.pushTask "Running goimports"
goimports -w "${files[@]}"
log.popTask
