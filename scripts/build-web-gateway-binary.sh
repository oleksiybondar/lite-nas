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
Usage: scripts/build-web-gateway-binary.sh [options]

Options:
  --arch=amd64|arm64  Target linux architecture. Defaults to current GOARCH.
  --output PATH       Output binary path. Defaults to .build/web-gateway/linux-<arch>/web-gateway
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
