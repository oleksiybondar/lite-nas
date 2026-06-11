#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-system-logging-manager-binary.sh" "system-logging-manager" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_MODULE" "system-logging-manager" "0" "0" "$@"
