#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-security-logging-manager-binary.sh" "security-logging-manager" "$LITE_NAS_SECURITY_LOGGING_MANAGER_MODULE" "security-logging-manager" "0" "0" "$@"
