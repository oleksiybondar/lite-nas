#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-security-logging-manager-cli-binary.sh" "security-logging-manager-cli" "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_APP_MODULE" "security-logging-manager-cli" "0" "0" "$@"
