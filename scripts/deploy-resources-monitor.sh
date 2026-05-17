#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/resources-monitor.sh"

binary_path=""
should_start=1
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
	--no-start)
		should_start=0
		shift
		;;
	--skip-bootstrap)
		should_bootstrap=0
		shift
		;;
	-h | --help)
		echo "Usage: scripts/deploy-resources-monitor.sh [--binary PATH] [--no-start] [--skip-bootstrap]"
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		exit 2
		;;
	esac
done

sudo.guard.requireRoot "scripts/deploy-resources-monitor.sh"
deploy.liteNAS.requireTools
deploy.resourcesMonitor.requireTools

if [ -z "$binary_path" ]; then
	tmp_dir="$(mktemp -d)"
	trap 'rm -rf "$tmp_dir"' EXIT
	binary_path="$tmp_dir/resources-monitor"
	"$ENTRYPOINT_DIR/build-resources-monitor-binary.sh" "--output=$binary_path"
fi

if [ "$should_bootstrap" -eq 1 ]; then
	deploy.liteNAS.bootstrap 1
fi

deploy.resourcesMonitor.deploy "$binary_path" "$should_start"
