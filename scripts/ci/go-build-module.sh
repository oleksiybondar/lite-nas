#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"

if [ "$#" -ne 1 ]; then
	log.error "Usage: scripts/ci/go-build-module.sh <module-dir>"
	exit 64
fi

module_dir="$1"
target_arch="$(go env GOARCH)"

if [ ! -f "${module_dir}/go.mod" ]; then
	log.error "Go module not found: ${module_dir}"
	exit 1
fi

export GOCACHE="${GOCACHE:-$(pwd)/.cache/go-build}"
mkdir -p "$GOCACHE"

log.requireCommand "go" "Install Go and retry."

log.pushTask "Building Go module in ${module_dir} for linux/${target_arch}"
(
	cd "$module_dir"
	GOOS=linux GOARCH="$target_arch" go build ./...
)
log.popTask
