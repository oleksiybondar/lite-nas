#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/go-modules.sh"

build.runGoBinaryScript "scripts/build-security-automation-cli-binary.sh" "security-automation-cli" "$LITE_NAS_SECURITY_AUTOMATION_CLI_MODULE" "security-automation-cli" "0" "0" "$@"
