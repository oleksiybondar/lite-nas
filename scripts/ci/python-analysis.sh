#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"

# shellcheck disable=SC1091
source .venv/bin/activate

mapfile -t python_files < <(find tests -type f -name '*.py' \
	-not -path '*/__pycache__/*')

if [ "${#python_files[@]}" -eq 0 ]; then
	log.info "No Python files found; skipping Python analysis."
	exit 0
fi

log.pushTask "Running ruff"
log.requireCommand "ruff" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-python-dependencies.sh."
ruff check --config dev-configs/ruff.toml tests/
log.popTask

log.pushTask "Running black"
log.requireCommand "black" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-python-dependencies.sh."
black --check --line-length 100 --target-version py311 tests/
log.popTask

log.pushTask "Running mypy"
log.requireCommand "mypy" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-python-dependencies.sh."
MYPYPATH=tests mypy --explicit-package-bases --config-file dev-configs/mypy.ini tests/
log.popTask

log.pushTask "Running xenon"
log.requireCommand "xenon" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-python-dependencies.sh."
xenon --max-absolute B --max-modules A --max-average A tests/
log.popTask
