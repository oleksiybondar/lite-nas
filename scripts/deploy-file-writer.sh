#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/file-writer.sh"

binary_path=""
should_bootstrap=1

while [ "$#" -gt 0 ]; do
	case "$1" in
	--binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --binary"
			exit 2
		fi
		binary_path="$2"
		shift 2
		;;
	--skip-bootstrap)
		should_bootstrap=0
		shift
		;;
	-h | --help)
		cat <<'MSG'
Usage: scripts/deploy-file-writer.sh [--binary PATH] [--skip-bootstrap]
MSG
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		exit 2
		;;
	esac
done

sudo.guard.requireRoot "scripts/deploy-file-writer.sh"
deploy.liteNAS.requireTools
deploy.fileWriter.requireTools

if [ -z "$binary_path" ]; then
	tmp_dir="$(mktemp -d)"
	trap 'rm -rf "$tmp_dir"' EXIT
	binary_path="$tmp_dir/lite-nas-file-writer"
	"$ENTRYPOINT_DIR/build-file-writer-binary.sh" "--output=$binary_path"
fi

if [ "$should_bootstrap" -eq 1 ]; then
	deploy.liteNAS.bootstrap 1
fi

deploy.fileWriter.deploy "$binary_path"
log.info "file-writer deployment completed."
