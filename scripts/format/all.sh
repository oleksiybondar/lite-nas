#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"

formatters=(
	"markdown"
	"js-ts"
	"go"
	"shell"
)

for formatter in "${formatters[@]}"; do
	log.pushTask "Formatting $formatter"
	"scripts/format/${formatter}.sh"
	log.popTask
done

log.info "Formatting complete."
