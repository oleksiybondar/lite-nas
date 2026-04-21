#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"
log.pushTask "Running local CI checks"
scripts/ci/check-all.sh
log.popTask
