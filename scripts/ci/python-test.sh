#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

log.pushTask "Running Moments caption worker unit tests"
python3 -m unittest discover -s services/moments -p '*_test.py'
log.popTask

log.pushTask "Running Python test suite"
"$LITE_NAS_REPO_ROOT/scripts/test-python.sh" "$@"
log.popTask
