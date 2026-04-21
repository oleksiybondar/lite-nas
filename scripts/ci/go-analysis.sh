#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"

export GOLANGCI_LINT_CACHE="${GOLANGCI_LINT_CACHE:-$(pwd)/.cache/golangci-lint}"
export GOCACHE="${GOCACHE:-$(pwd)/.cache/go-build}"
mkdir -p "$GOLANGCI_LINT_CACHE"
mkdir -p "$GOCACHE"

mapfile -t modules < <(find . -name go.mod -not -path './vendor/*')

if [ "${#modules[@]}" -eq 0 ]; then
	log.info "No Go modules found."
	exit 0
fi

for module in "${modules[@]}"; do
	dir="$(dirname "$module")"
	log.pushTask "Running golangci-lint in ${dir}"
	log.requireCommand "golangci-lint" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-go-dependencies.sh."
	(cd "$dir" && golangci-lint run)
	log.popTask
done
