#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-litenas-integrity-binary.sh" "litenas-integrity" "$LITE_NAS_LITENAS_INTEGRITY_MODULE" "litenas-integrity" "0" "0" "$@"
