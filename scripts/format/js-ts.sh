#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f \( \
	-name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' -o -name '*.json' -o -name '*.jsonc' \) \
	-not -path './node_modules/*' \
	-not -path './dist/*' \
	-not -path './build/*')

if [ "${#files[@]}" -eq 0 ]; then
	log.info "No JS/TS/JSON files found."
	exit 0
fi

log.pushTask "Running Biome autofix"
log.requireCommand "npx" "Run ./scripts/install-dev-dependencies.sh to install Node developer dependencies."
npx --no-install biome check --write .
log.popTask
