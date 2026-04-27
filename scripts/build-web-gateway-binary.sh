#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

output_path=""

usage() {
	cat <<'MSG'
Usage: scripts/build-web-gateway-binary.sh [options]

Options:
  --output PATH       Output binary path. Defaults to .build/web-gateway/linux-<arch>/web-gateway
  -h, --help          Show this help.
MSG
}

for arg in "$@"; do
	case "$arg" in
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

target_arch="$(go env GOARCH)"

if [ -z "$output_path" ]; then
	output_path="$LITE_NAS_REPO_ROOT/.build/web-gateway/linux-${target_arch}/web-gateway"
fi

output_dir="$(dirname "$output_path")"
mkdir -p "$output_dir"

export GOCACHE="${GOCACHE:-${TMPDIR:-/tmp}/lite-nas-go-build}"
mkdir -p "$GOCACHE"

log.pushTask "Building web-gateway binary for linux/${target_arch}"
(
	cd "$LITE_NAS_WEB_GATEWAY_MODULE"
	CGO_ENABLED=0 GOOS=linux GOARCH="$target_arch" go build \
		-ldflags="-s -w" \
		-o "$output_path" .
)
log.popTask

log.info "Built binary: $output_path"
