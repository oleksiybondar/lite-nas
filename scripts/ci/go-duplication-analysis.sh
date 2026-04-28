#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t duplicate_files < <(find . -type f -name '*.go' \
	-not -path './node_modules/*' \
	-not -path './dist/*' \
	-not -path './build/*' \
	-not -path './.cache/*' \
	-not -path './.build/*' \
	-not -path './.artifacts/*')

if [ "${#duplicate_files[@]}" -eq 0 ]; then
	log.info "No Go files found for duplication detection."
	exit 0
fi

log.pushTask "Running Go duplication detection"
log.requireCommand "npx" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-node-dependencies.sh."
npx --no-install jscpd --config dev-configs/jscpd-go.json .
log.popTask
