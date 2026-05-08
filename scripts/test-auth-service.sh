#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/go-modules.sh"

"$SCRIPT_DIR/ci/go-test-module.sh" "$LITE_NAS_AUTH_SERVICE_MODULE" "$@"
