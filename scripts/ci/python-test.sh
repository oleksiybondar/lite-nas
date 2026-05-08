#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

log.pushTask "Running Python test suite"
"$LITE_NAS_REPO_ROOT/scripts/test-python.sh" "$@"
log.popTask
