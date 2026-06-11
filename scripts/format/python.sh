#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t python_files < <(find tests -type f -name '*.py')

if [ "${#python_files[@]}" -eq 0 ]; then
	log.info "No Python files found."
	exit 0
fi

log.pushTask "Running ruff autofix"
log.requireCommand "ruff" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-python-dependencies.sh."
ruff check --fix --config dev-configs/ruff.toml tests/
log.popTask

log.pushTask "Running black autofix"
log.requireCommand "black" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-python-dependencies.sh."
black --line-length 100 --target-version py311 tests/
log.popTask
