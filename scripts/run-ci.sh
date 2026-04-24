#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"
log.warn "scripts/run-ci.sh is kept as a compatibility wrapper. Use scripts/run-ci-analysis.sh."
log.pushTask "Running local CI static analysis checks"
scripts/run-ci-analysis.sh
log.popTask
