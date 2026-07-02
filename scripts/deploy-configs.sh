#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/normalize-etc-permissions.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/restart-affected-services.sh"

source_dir="$LITE_NAS_REPO_ROOT/configs/etc"
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

validate_sudoers_drop_ins() {
	local sudoers_dir="$1"
	local validated_any=0
	local sudoers_file

	if [ ! -d "$sudoers_dir" ]; then
		return 0
	fi

	log.requireCommand "visudo" "Install visudo and retry."

	while IFS= read -r -d '' sudoers_file; do
		deploy.validateSudoersDropIn "$sudoers_file"
		validated_any=1
	done < <(find "$sudoers_dir" -maxdepth 1 -type f -name 'lite-nas-*' -print0)

	if [ "$validated_any" -eq 1 ]; then
		log.info "Validated LiteNAS sudoers drop-ins in $sudoers_dir"
	fi
}

log.pushTask "Deploying etc configs"
cp -a "$source_dir/." "$target_dir/"
log.popTask

validate_sudoers_drop_ins "$target_dir/sudoers.d"

deploy.normalizeEtcPermissions "$target_dir"

if [ "$restart_services" -eq 1 ]; then
	deploy.restartAffectedServices
else
	log.info "Skipping affected service restart."
fi

log.info "Config deployment completed."
