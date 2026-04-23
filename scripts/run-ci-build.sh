#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"
log.pushTask "Running local CI build checks"
scripts/build-system-metrics.sh --arch=amd64
scripts/build-system-metrics.sh --arch=arm64
scripts/build-shared.sh --arch=amd64
scripts/build-shared.sh --arch=arm64
log.popTask
