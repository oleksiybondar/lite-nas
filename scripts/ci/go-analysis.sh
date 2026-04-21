#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t modules < <(find . -name go.mod -not -path './vendor/*')

if [ "${#modules[@]}" -eq 0 ]; then
  log.info "No Go modules found."
  exit 0
fi

for module in "${modules[@]}"; do
  dir="$(dirname "$module")"
  log.pushTask "Running golangci-lint in ${dir}"
  (cd "$dir" && golangci-lint run)
  log.popTask
done
