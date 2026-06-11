#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

log.pushTask "Generating parser artifacts"
"$LITE_NAS_REPO_ROOT/scripts/antlr4/generate-zpool-status-parser.sh"
"$LITE_NAS_REPO_ROOT/scripts/antlr4/generate-getfacl-parser.sh"
log.popTask

log.info "All parser generation tasks completed."
