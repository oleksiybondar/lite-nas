#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

pytest_bin="pytest"
if [ -x "$LITE_NAS_REPO_ROOT/.venv/bin/pytest" ]; then
	pytest_bin="$LITE_NAS_REPO_ROOT/.venv/bin/pytest"
fi

log.requireCommand "$pytest_bin" "Run ./scripts/install-dev-dependencies.sh or scripts/ci/install-python-dependencies.sh."

cd "$LITE_NAS_REPO_ROOT/tests"
export PYTHONPATH="$LITE_NAS_REPO_ROOT/tests${PYTHONPATH:+:$PYTHONPATH}"

log.pushTask "Preparing Python test log directory"
rm -rf logs
mkdir -p logs
log.popTask

run_marker_tests() {
	local marker="$1"
	local status=0
	shift

	log.pushTask "Running Python ${marker} tests"
	set +e
	"$pytest_bin" -c pytest.ini . -m "$marker" "$@"
	status=$?
	set -e

	if [ "$status" -eq 5 ]; then
		log.info "No Python ${marker} tests found; skipping."
		log.popTask
		return 0
	fi

	log.popTask
	return "$status"
}

run_marker_tests infra "$@"
run_marker_tests cli "$@"
run_marker_tests api "$@"
run_marker_tests ui "$@"
