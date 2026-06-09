#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-resources-monitor-binary.sh" "resources-monitor" "$LITE_NAS_RESOURCES_MONITOR_MODULE" "resources-monitor" "0" "0" "$@"
