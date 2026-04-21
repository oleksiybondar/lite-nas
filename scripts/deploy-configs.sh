#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/sudo-guard.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/normalize-etc-permissions.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/restart-affected-services.sh"

source_dir="$REPO_ROOT/configs/etc"
target_dir="${LITE_NAS_ETC_TARGET:-/etc}"
restart_services=1

usage() {
	cat <<'MSG'
Usage: scripts/deploy-configs.sh [options]

Options:
  --source-dir PATH  Source etc config directory. Defaults to configs/etc.
  --target-dir PATH  Target etc directory. Defaults to /etc.
  --no-restart       Deploy files without restarting affected services.
  -h, --help         Show this help.
MSG
}

require_option_value() {
	local option="$1"
	local value="${2:-}"

	if [ -z "$value" ]; then
		log.error "Missing value for $option"
		usage >&2
		exit 2
	fi
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--source-dir)
		require_option_value "$1" "${2:-}"
		source_dir="${2:-}"
		shift 2
		;;
	--target-dir)
		require_option_value "$1" "${2:-}"
		target_dir="${2:-}"
		shift 2
		;;
	--no-restart)
		restart_services=0
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

sudo.guard.requireRoot "scripts/deploy-configs.sh"

require_directory() {
	local path="$1"
	local description="$2"

	if [ ! -d "$path" ]; then
		log.error "Missing $description directory: $path"
		exit 1
	fi
}

require_directory "$source_dir" "source"
require_directory "$target_dir" "target"

log.pushTask "Deploying etc configs"
cp -a "$source_dir/." "$target_dir/"
log.popTask

deploy.normalizeEtcPermissions "$target_dir"

if [ "$restart_services" -eq 1 ]; then
	deploy.restartAffectedServices
else
	log.info "Skipping affected service restart."
fi

log.info "Config deployment completed."
