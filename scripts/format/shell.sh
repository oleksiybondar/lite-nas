#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f \( -name '*.sh' -o -name '*.bash' -o -name '*.zsh' \) \
	-not -path './node_modules/*' \
	-not -path './dist/*' \
	-not -path './build/*')

if [ "${#files[@]}" -eq 0 ]; then
	log.info "No shell files found."
	exit 0
fi

log.pushTask "Running shfmt autofix"
log.requireCommand "shfmt" "Run ./scripts/install-dev-dependencies.sh to install shfmt."
shfmt -w "${files[@]}"
log.popTask
