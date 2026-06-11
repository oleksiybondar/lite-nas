#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-security-email-notifier-binary.sh" "security-email-notifier" "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_MODULE" "security-email-notifier" "0" "0" "$@"
