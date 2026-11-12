#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/go-modules.sh"

build.runGoBinaryScript "scripts/build-operational-safety-automation-binary.sh" "operational-safety-automation" "$LITE_NAS_OPERATIONAL_SAFETY_AUTOMATION_MODULE" "operational-safety-automation" "0" "0" "$@"
