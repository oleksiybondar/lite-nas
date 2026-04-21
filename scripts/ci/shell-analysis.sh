#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f \( -name '*.sh' -o -name '*.bash' -o -name '*.zsh' \) \
	-not -path './node_modules/*' \
	-not -path './dist/*' \
	-not -path './build/*')

if [ "${#files[@]}" -eq 0 ]; then
	log.info "No shell files found."
	exit 0
fi

log.pushTask "Running shellcheck"
log.requireCommand "shellcheck" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-shell-dependencies.sh."
shellcheck "${files[@]}"
log.popTask

log.pushTask "Running shfmt"
log.requireCommand "shfmt" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-shell-dependencies.sh."
shfmt -d "${files[@]}"
log.popTask
