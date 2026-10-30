#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-backup-restore-cli-binary.sh" "backup-restore-cli" "$LITE_NAS_BACKUP_RESTORE_CLI_MODULE" "backup-restore-cli" "0" "0" "$@"
