#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

output_dir=""

usage() {
	cat <<'MSG'
Usage: scripts/build-all.sh [options]

Options:
  --output-dir PATH   Output directory. Defaults to .build/all/linux-<arch>
  -h, --help          Show this help.
MSG
}

for arg in "$@"; do
	case "$arg" in
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

target_arch="$(go env GOARCH)"

if [ -z "$output_dir" ]; then
	output_dir="$LITE_NAS_REPO_ROOT/.build/all/linux-${target_arch}"
fi

mkdir -p "$output_dir"

log.pushTask "Building deployable binaries for linux/${target_arch}"
"$SCRIPT_DIR/build-auth-service-binary.sh" --output="$output_dir/auth-service"
"$SCRIPT_DIR/build-system-metrics-binary.sh" --output="$output_dir/system-metrics"
"$SCRIPT_DIR/build-system-metrics-cli-binary.sh" --output="$output_dir/system-metrics-cli"
"$SCRIPT_DIR/build-web-gateway-binary.sh" --output="$output_dir/web-gateway"
log.popTask

log.info "Built binaries in: $output_dir"
