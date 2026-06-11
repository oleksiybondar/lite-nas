#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/repo-files.sh"

declare -a duplicate_files=()
repo.files.collectShellFiles duplicate_files

if [ "${#duplicate_files[@]}" -eq 0 ]; then
	log.info "No shell files found for duplication detection."
	exit 0
fi

log.pushTask "Running shell duplication detection"
log.requireCommand "npx" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-node-dependencies.sh."
npx --no-install jscpd --config dev-configs/jscpd-bash.json .
log.popTask
