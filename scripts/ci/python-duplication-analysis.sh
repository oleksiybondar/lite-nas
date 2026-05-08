#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t duplicate_files < <(find tests -type f -name '*.py' \
	-not -path '*/__pycache__/*')

if [ "${#duplicate_files[@]}" -eq 0 ]; then
	log.info "No Python files found for duplication detection."
	exit 0
fi

log.pushTask "Running Python duplication detection"
log.requireCommand "npx" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-node-dependencies.sh."
npx --no-install jscpd --config dev-configs/jscpd-python.json .
log.popTask
