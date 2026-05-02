#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

log.requireCommand "npm" "Install Node.js/npm and retry."

log.pushTask "Testing admin-panel frontend"
npm --prefix "$LITE_NAS_REPO_ROOT/apps/admin-panel" run test:unit
log.popTask
