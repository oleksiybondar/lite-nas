#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/build-binary.sh"

build.runGoBinaryScript "scripts/build-file-writer-binary.sh" "lite-nas-file-writer" "$LITE_NAS_FILE_WRITER_APP_MODULE" "lite-nas-file-writer" "0" "0" "$@"
