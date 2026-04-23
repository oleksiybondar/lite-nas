#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

target_arch=""
output_path=""

usage() {
	cat <<'MSG'
Usage: scripts/build-system-metrics-binary.sh [options]

Options:
  --arch=amd64|arm64  Target linux architecture. Defaults to current GOARCH.
  --output PATH       Output binary path. Defaults to .build/system-metrics/linux-<arch>/system-metrics
  -h, --help          Show this help.
MSG
}

for arg in "$@"; do
	case "$arg" in
	--arch=amd64)
		target_arch="amd64"
		;;
	--arch=arm64)
		target_arch="arm64"
		;;
	--output=*)
		output_path="${arg#--output=}"
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		log.error "Unknown option: $arg"
		usage >&2
		exit 64
		;;
	esac
done

log.requireCommand "go" "Install Go and retry."

if [ -z "$target_arch" ]; then
	target_arch="$(go env GOARCH)"
fi

if [ -z "$output_path" ]; then
	output_path="$LITE_NAS_REPO_ROOT/.build/system-metrics/linux-${target_arch}/system-metrics"
fi

output_dir="$(dirname "$output_path")"
mkdir -p "$output_dir"

export GOCACHE="${GOCACHE:-${TMPDIR:-/tmp}/lite-nas-go-build}"
mkdir -p "$GOCACHE"

log.pushTask "Building system-metrics binary for linux/${target_arch}"
(
	cd "$LITE_NAS_SYSTEM_METRICS_MODULE"
	GOOS=linux GOARCH="$target_arch" go build -o "$output_path" .
)
log.popTask

log.info "Built binary: $output_path"
