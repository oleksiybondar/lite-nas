#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/moments.sh"

binary_path=""
assets_source=""
should_start=1
should_bootstrap=1

usage() {
	cat <<'MSG'
Usage: scripts/deploy-moments.sh [--binary PATH] [--assets-source PATH] [--no-start] [--skip-bootstrap]
MSG
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--binary | --assets-source)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for $1"
			exit 2
		fi
		if [ "$1" = "--binary" ]; then binary_path="$2"; else assets_source="$2"; fi
		shift 2
		;;
	--no-start)
		should_start=0
		shift
		;;
	--skip-bootstrap)
		should_bootstrap=0
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		usage >&2
		exit 2
		;;
	esac
done

sudo.guard.requireRoot "scripts/deploy-moments.sh"

if [ -z "$binary_path" ]; then
	# Vite replaces its asset output directory, so keep the Go binary outside it.
	binary_path="$LITE_NAS_REPO_ROOT/.build/moments-binary/moments"
	"$SCRIPT_DIR/build-moments-binary.sh" --output="$binary_path"
fi
if [ -z "$assets_source" ]; then
	assets_source="$LITE_NAS_REPO_ROOT/.build/moments"
	"$SCRIPT_DIR/build-moments.sh" --output-dir="$assets_source"
fi
if [ ! -f "$binary_path" ] || [ ! -f "$assets_source/index.html" ] || [ ! -d "$assets_source/assets" ]; then
	log.error "Missing Moments binary or frontend assets."
	exit 1
fi

if [ "$should_bootstrap" -eq 1 ]; then
	deploy.liteNAS.bootstrap 1
fi

deploy.moments.deploy "$binary_path" "$assets_source" "$should_start"
log.info "Moments deployment completed."
