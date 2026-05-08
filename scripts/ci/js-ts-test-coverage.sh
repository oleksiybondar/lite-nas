#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

log.pushTask "Running admin-panel unit tests with coverage"
"$LITE_NAS_REPO_ROOT/scripts/test-admin-panel.sh" --with-coverage
log.popTask
