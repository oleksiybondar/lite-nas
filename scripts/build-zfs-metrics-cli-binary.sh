#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-zfs-metrics-cli-binary.sh" "zfs-metrics-cli" "$LITE_NAS_ZFS_METRICS_CLI_APP_MODULE" "zfs-metrics-cli" "0" "0" "$@"
