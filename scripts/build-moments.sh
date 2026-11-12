#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

output_dir="$LITE_NAS_REPO_ROOT/.build/moments"
for arg in "$@"; do
	case "$arg" in
	--output-dir=*) output_dir="${arg#*=}" ;;
	-h | --help)
		printf 'Usage: scripts/build-moments.sh [--output-dir=PATH]\n'
		exit 0
		;;
	*)
		log.error "Unknown option: $arg"
		exit 64
		;;
	esac
done

if [ -z "$output_dir" ]; then
	log.error "Output directory cannot be empty."
	exit 64
fi
if [ "${output_dir#/}" = "$output_dir" ]; then
	output_dir="$LITE_NAS_REPO_ROOT/$output_dir"
fi

log.requireCommand npm "Install Node.js/npm and retry."
LITE_NAS_MOMENTS_OUT_DIR="$output_dir" npm --prefix "$LITE_NAS_REPO_ROOT/apps/moments" run build
