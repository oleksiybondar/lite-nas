#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

target_arch=""
output_dir=""

usage() {
	cat <<'MSG'
Usage: scripts/build-all.sh [options]

Options:
  --arch=amd64|arm64  Target linux architecture. Defaults to current GOARCH.
  --output-dir PATH   Output directory. Defaults to .build/all/linux-<arch>
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
	--output-dir=*)
		output_dir="${arg#--output-dir=}"
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

if [ -z "$output_dir" ]; then
	output_dir="$LITE_NAS_REPO_ROOT/.build/all/linux-${target_arch}"
fi

mkdir -p "$output_dir"

log.pushTask "Building deployable binaries for linux/${target_arch}"
"$SCRIPT_DIR/build-system-metrics-binary.sh" --arch="$target_arch" --output="$output_dir/system-metrics"
"$SCRIPT_DIR/build-system-metrics-cli-binary.sh" --arch="$target_arch" --output="$output_dir/system-metrics-cli"
"$SCRIPT_DIR/build-web-gateway-binary.sh" --arch="$target_arch" --output="$output_dir/web-gateway"
log.popTask

log.info "Built binaries in: $output_dir"
