#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

output_path=""
host_arch=""

usage() {
	cat <<'MSG'
Usage: scripts/build-auth-service-binary.sh [options]

Options:
  --output PATH       Output binary path. Defaults to .build/auth-service/linux-<arch>/auth-service
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
log.requireCommand "gcc" "Install a C compiler and retry."

host_arch="$(go env GOARCH)"
target_arch="$host_arch"

if [ -z "$output_path" ]; then
	output_path="$LITE_NAS_REPO_ROOT/.build/auth-service/linux-${target_arch}/auth-service"
fi

output_dir="$(dirname "$output_path")"
mkdir -p "$output_dir"

export GOCACHE="${GOCACHE:-${TMPDIR:-/tmp}/lite-nas-go-build}"
mkdir -p "$GOCACHE"

log.pushTask "Building auth-service binary for linux/${target_arch}"
(
	cd "$LITE_NAS_AUTH_SERVICE_MODULE"
	CGO_ENABLED=1 GOOS=linux GOARCH="$target_arch" go build \
		-tags pam \
		-ldflags="-s -w" \
		-o "$output_path" .
)
log.popTask

log.info "Built binary: $output_path"
