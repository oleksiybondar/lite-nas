#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

output_dir=""

usage() {
	cat <<'MSG'
Usage: scripts/build-admin-panel.sh [options]

Options:
  --output-dir PATH   Output directory. Defaults to .build/admin-panel
  -h, --help          Show this help.
MSG
}

args.parse "$@"
if ! args.assertKnown output-dir help h; then
	log.error "Unknown option: --$(args.unknownKeys output-dir help h | head -n 1)"
	usage >&2
	exit 64
fi
if args.has h || args.has help; then
	usage
	exit 0
fi
if args.has output-dir && ! output_dir="$(args.require_arg output-dir)"; then
	log.error "Missing value for --output-dir"
	usage >&2
	exit 64
fi

log.requireCommand "npm" "Install Node.js/npm and retry."

if [ -z "$output_dir" ]; then
	output_dir="$LITE_NAS_REPO_ROOT/.build/admin-panel"
elif [ "${output_dir#/}" = "$output_dir" ]; then
	output_dir="$LITE_NAS_REPO_ROOT/$output_dir"
fi

log.pushTask "Building admin-panel frontend"
LITE_NAS_ADMIN_PANEL_OUT_DIR="$output_dir" npm --prefix "$LITE_NAS_REPO_ROOT/apps/admin-panel" run build
log.popTask

log.info "Built admin-panel assets in: $output_dir"
