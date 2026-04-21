#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

checks=(
	"markdown-analysis"
	"shell-analysis"
	"js-ts-analysis"
	"go-analysis"
)

for check in "${checks[@]}"; do
	log.pushTask "Running $check"
	"scripts/ci/${check}.sh"
	log.popTask
done

log.info "All local CI checks passed."
