#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-system-metrics-cli-binary.sh" "system-metrics-cli" "$LITE_NAS_SYSTEM_METRICS_CLI_APP_MODULE" "system-metrics-cli" "0" "0" "$@"
