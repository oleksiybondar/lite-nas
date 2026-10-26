#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-system-service-monitor-binary.sh" "system-service-monitor" "${LITE_NAS_SYSTEM_SERVICE_MONITOR_MODULE}" "system-service-monitor" "0" "0" "$@"
