#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/zfs-metrics-cli.sh"

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
Usage: scripts/deploy-zfs-metrics-cli.sh [options]

Options:
  --binary PATH       Install an existing binary instead of building one.
  --skip-bootstrap    Install files without running LiteNAS bootstrap first.
  -h, --help          Show this help.
MSG
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		exit 2
		;;
	esac
done

sudo.guard.requireRoot "scripts/deploy-zfs-metrics-cli.sh"
deploy.liteNAS.requireTools
deploy.zfsMetricsCLI.requireTools

if [ -z "$binary_path" ]; then
	tmp_dir="$(mktemp -d)"
	trap 'rm -rf "$tmp_dir"' EXIT

	binary_path="$tmp_dir/zfs-metrics-cli"
	log.pushTask "Building zfs-metrics-cli deploy artifact"
	"$ENTRYPOINT_DIR/build-zfs-metrics-cli-binary.sh" "--output=$binary_path"
	log.popTask
fi

if [ "$should_bootstrap" -eq 1 ]; then
	log.pushTask "Bootstrapping LiteNAS prerequisites"
	deploy.liteNAS.bootstrap 1
	log.popTask
else
	log.info "Skipping LiteNAS bootstrap."
fi

log.pushTask "Deploying zfs-metrics-cli app"
deploy.zfsMetricsCLI.deploy "$binary_path"
log.popTask

log.info "zfs-metrics-cli deployment completed."
