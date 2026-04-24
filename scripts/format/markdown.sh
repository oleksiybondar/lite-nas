#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f -name '*.md' \
	-not -path './node_modules/*' \
	-not -path './dist/*' \
	-not -path './build/*')

if [ "${#files[@]}" -eq 0 ]; then
	log.info "No Markdown files found."
	exit 0
fi

log.pushTask "Running markdownlint autofix"
log.requireCommand "npx" "Run ./scripts/install-dev-dependencies.sh to install Node developer dependencies."
npx --no-install markdownlint-cli2 --fix "**/*.md"
log.popTask
