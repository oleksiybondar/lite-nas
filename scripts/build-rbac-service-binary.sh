#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

output_path=""

usage() {
	cat <<'MSG'
Usage: scripts/build-rbac-service-binary.sh [options]

Options:
  --output PATH       Output binary path. Defaults to .build/rbac-service/linux-<arch>/rbac-service
  -h, --help          Show this help.
MSG
}

args.parse "$@"
if ! args.assertKnown output help h; then
	log.error "Unknown option: --$(args.unknownKeys output help h | head -n 1)"
	usage >&2
	exit 64
fi
if args.has h || args.has help; then
	usage
	exit 0
fi
if args.has output && ! output_path="$(args.require_arg output)"; then
	log.error "Missing value for --output"
	usage >&2
	exit 64
fi

log.requireCommand "go" "Install Go and retry."

target_arch="$(build.resolveTargetArch)"
output_path="$(build.prepareOutputPath "$output_path" "rbac-service" "rbac-service")"

output_dir="$(dirname "$output_path")"
mkdir -p "$output_dir"

build.prepareGoCache
build.goBinary "rbac-service" "$LITE_NAS_RBAC_SERVICE_MODULE" "$output_path" 0 "$target_arch"
