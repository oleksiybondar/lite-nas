#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/system-metrics-cli.sh"

binary_path=""
target_arch=""
should_bootstrap=1

while [ "$#" -gt 0 ]; do
	case "$1" in
	--binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --binary"
			deploy.systemMetricsCLI.usage >&2
			exit 2
		fi
		binary_path="$2"
		shift 2
		;;
	--arch=amd64 | --arch=arm64)
		target_arch="${1#--arch=}"
		shift
		;;
	--skip-bootstrap)
		should_bootstrap=0
		shift
		;;
	-h | --help)
		deploy.systemMetricsCLI.usage
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		deploy.systemMetricsCLI.usage >&2
		exit 2
		;;
	esac
done

sudo.guard.requireRoot "scripts/deploy-system-metrics-cli.sh"
deploy.liteNAS.requireTools
deploy.systemMetricsCLI.requireTools

if [ -z "$binary_path" ]; then
	tmp_dir="$(mktemp -d)"
	trap 'rm -rf "$tmp_dir"' EXIT

	binary_path="$tmp_dir/system-metrics-cli"
	build_args=("--output=$binary_path")
	if [ -n "$target_arch" ]; then
		build_args+=("--arch=$target_arch")
	fi

	log.pushTask "Building system-metrics-cli deploy artifact"
	"$ENTRYPOINT_DIR/build-system-metrics-cli-binary.sh" "${build_args[@]}"
	log.popTask
fi

if [ "$should_bootstrap" -eq 1 ]; then
	log.pushTask "Bootstrapping LiteNAS prerequisites"
	deploy.liteNAS.bootstrap 1
	log.popTask
else
	log.info "Skipping LiteNAS bootstrap."
fi

log.pushTask "Installing system-metrics-cli binary"
deploy.systemMetricsCLI.installBinary "$binary_path"
log.popTask

log.pushTask "Installing system-metrics-cli config"
deploy.systemMetricsCLI.installConfig
log.popTask

log.info "system-metrics-cli deployment completed."
