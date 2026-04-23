#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"

if [ "$#" -lt 1 ] || [ "$#" -gt 3 ]; then
	log.error "Usage: scripts/ci/go-test-module.sh <module-dir> [--with-coverage] [--minimum-coverage=<percent>]"
	exit 64
fi

module_dir="$1"
with_coverage=0
minimum_coverage="${GO_TEST_MIN_COVERAGE:-75}"

shift
for option in "$@"; do
	case "$option" in
	--with-coverage)
		with_coverage=1
		;;
	--minimum-coverage=*)
		minimum_coverage="${option#--minimum-coverage=}"
		;;
	*)
		log.error "Unsupported test option: $option"
		exit 64
		;;
	esac
done

if [ ! -f "${module_dir}/go.mod" ]; then
	log.error "Go module not found: ${module_dir}"
	exit 1
fi

export GOCACHE="${GOCACHE:-$(pwd)/.cache/go-build}"
mkdir -p "$GOCACHE"

log.requireCommand "go" "Install Go and retry."

if ! find "$module_dir" -name '*_test.go' -not -path '*/vendor/*' -print -quit | grep -q .; then
	log.warn "No Go tests found in ${module_dir}; skipping test run."
	exit 0
fi

if [ "$with_coverage" -eq 1 ]; then
	coverage_output_dir="${GO_TEST_COVERAGE_DIR:-$(pwd)/.cache/go-coverage}"
	mkdir -p "$coverage_output_dir"

	module_name="${module_dir#./}"
	profile_name="${module_name//\//-}.cover.out"
	profile_path="${coverage_output_dir}/${profile_name}"

	log.pushTask "Running Go tests with coverage in ${module_dir}"
	(
		cd "$module_dir"
		go test -covermode=atomic -coverpkg=./... -coverprofile="$profile_path" ./...
	)

	coverage_pct="$(
		(
			cd "$module_dir"
			go tool cover -func="$profile_path"
		) | awk '/^total:/ { gsub(/%/, "", $3); print $3 }'
	)"

	if [ -z "$coverage_pct" ]; then
		log.error "Failed to read total coverage for ${module_dir}."
		exit 1
	fi

	log.info "Total coverage for ${module_dir}: ${coverage_pct}%"

	if ! awk -v actual="$coverage_pct" -v minimum="$minimum_coverage" 'BEGIN { exit !(actual + 0 >= minimum + 0) }'; then
		log.error "Coverage for ${module_dir} is below the required ${minimum_coverage}%."
		exit 1
	fi
else
	log.pushTask "Running Go tests in ${module_dir}"
	(
		cd "$module_dir"
		go test ./...
	)
fi

log.popTask
