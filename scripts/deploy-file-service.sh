#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/file-service.sh"

source_binary=""
should_start=1
should_bootstrap=1
while [ "$#" -gt 0 ]; do
	case "$1" in
	--binary)
		source_binary="$2"
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
		printf '%s\n' 'Usage: scripts/deploy-file-service.sh [--binary PATH] [--no-start] [--skip-bootstrap]'
		exit 0
		;;
	*)
		printf 'Unknown option: %s\n' "$1" >&2
		exit 2
		;;
	esac
done
sudo.guard.requireRoot "scripts/deploy-file-service.sh"
if [ "$should_bootstrap" = "1" ]; then
	deploy.liteNAS.bootstrap 1
fi
if [ -z "$source_binary" ]; then
	source_binary="$(mktemp)"
	trap 'rm -f "$source_binary"' EXIT
	"$SCRIPT_DIR/build-file-service-binary.sh" --output="$source_binary"
fi
deploy.fileService.deploy "$source_binary" "$should_start"
