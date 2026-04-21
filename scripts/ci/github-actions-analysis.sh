#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find .github/workflows -type f \( -name '*.yml' -o -name '*.yaml' \) 2>/dev/null || true)

if [ "${#files[@]}" -eq 0 ]; then
	log.info "No GitHub Actions workflow files found."
	exit 0
fi

log.pushTask "Running actionlint"
log.requireCommand "actionlint" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-github-actions-dependencies.sh."
actionlint "${files[@]}"
log.popTask
