#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/repo-files.sh"

declare -a files=()
repo.files.collectShellFiles files

if [ "${#files[@]}" -eq 0 ]; then
	log.info "No shell files found."
	exit 0
fi

log.pushTask "Running shfmt autofix"
log.requireCommand "shfmt" "Run ./scripts/install-dev-dependencies.sh to install shfmt."
shfmt -w "${files[@]}"
log.popTask
