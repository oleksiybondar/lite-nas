#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

readonly MINIMUM_COVERAGE="${GO_TEST_MIN_COVERAGE:-75}"

mapfile -t modules < <(find . -name go.mod -not -path './vendor/*' | sort)

if [ "${#modules[@]}" -eq 0 ]; then
	log.info "No Go modules found."
	exit 0
fi

ran_coverage_check=0

for module in "${modules[@]}"; do
	module_dir="$(dirname "$module")"
	ran_coverage_check=1
	"$SCRIPT_DIR/go-test-module.sh" "$module_dir" --with-coverage "--minimum-coverage=${MINIMUM_COVERAGE}"
done

if [ "$ran_coverage_check" -eq 0 ]; then
	log.info "No Go test files found in any module; skipping coverage check."
	exit 0
fi

log.info "Go test coverage checks passed."
