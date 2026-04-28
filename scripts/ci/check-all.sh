#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"

checks=(
	"github-actions-analysis"
	"markdown-analysis"
	"shell-analysis"
	"go-duplication-analysis"
	"bash-duplication-analysis"
	"js-ts-analysis"
	"go-analysis"
	"go-test-coverage"
	"package-analysis"
)

for check in "${checks[@]}"; do
	log.pushTask "Running $check"
	"scripts/ci/${check}.sh"
	log.popTask
done

log.info "All local CI checks passed."
