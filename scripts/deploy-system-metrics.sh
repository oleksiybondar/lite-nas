#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/system-metrics.sh"

binary_path=""
should_start=1
should_bootstrap=1

while [ "$#" -gt 0 ]; do
	case "$1" in
	--binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --binary"
			deploy.systemMetrics.usage >&2
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
		deploy.systemMetrics.usage
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		deploy.systemMetrics.usage >&2
		exit 2
		;;
	esac
done

sudo.guard.requireRoot "scripts/deploy-system-metrics.sh"
deploy.liteNAS.requireTools
deploy.systemMetrics.requireTools

if [ -z "$binary_path" ]; then
	tmp_dir="$(mktemp -d)"
	trap 'rm -rf "$tmp_dir"' EXIT

	binary_path="$tmp_dir/system-metrics"
	build_args=("--output=$binary_path")

	log.pushTask "Building system-metrics deploy artifact"
	"$ENTRYPOINT_DIR/build-system-metrics-binary.sh" "${build_args[@]}"
	log.popTask
fi

if [ "$should_bootstrap" -eq 1 ]; then
	log.pushTask "Bootstrapping LiteNAS prerequisites"
	deploy.liteNAS.bootstrap 1
	log.popTask
else
	log.info "Skipping LiteNAS bootstrap."
fi

log.pushTask "Deploying system-metrics service"
deploy.systemMetrics.deploy "$binary_path" "$should_start"
log.popTask

log.info "system-metrics deployment completed."
