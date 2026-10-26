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
manage_nats_config=1

usage() {
	cat <<'MSG'
Usage: scripts/deploy-configs.sh [options]

Options:
  --source-dir PATH  Source etc config directory. Defaults to configs/etc.
  --target-dir PATH  Target etc directory. Defaults to /etc.
  --no-nats-config   Keep the current NATS configuration unchanged.
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
	--no-nats-config)
		manage_nats_config=0
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

kill_helper_source="$LITE_NAS_REPO_ROOT/scripts/helpers/lite-nas-kill-pid.sh"
kill_helper_target="${LITE_NAS_KILL_HELPER_TARGET:-/usr/libexec/lite-nas/lite-nas-kill-pid}"
if [ ! -f "$kill_helper_source" ]; then
	log.error "Missing service-admin kill helper: $kill_helper_source"
	exit 1
fi
install -D -o root -g root -m 0755 "$kill_helper_source" "$kill_helper_target"

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
security_monitor_saved="$(mktemp -d)"
trap 'rm -rf "$security_monitor_saved"' EXIT
for config_name in mount-monitor.conf antivirus.conf system-service-monitor.conf service-watchdog.conf log-analysis.conf; do
	if [ -f "$target_dir/lite-nas/$config_name" ]; then
		cp -p "$target_dir/lite-nas/$config_name" "$security_monitor_saved/$config_name"
	fi
done
if [ -d "$target_dir/lite-nas/log-analysis" ]; then
	cp -a "$target_dir/lite-nas/log-analysis" "$security_monitor_saved/log-analysis"
fi
old_moments_link=/usr/share/lite-nas/nginx/moments.conf
if [ "$target_dir" = "/etc" ] && [ -L "$old_moments_link" ] && [ "$(readlink "$old_moments_link")" = "/usr/share/lite-nas/moments/nginx/moments.conf" ]; then
	rm -f "$old_moments_link"
fi
config_entries=("$source_dir"/*)
for config_entry in "${config_entries[@]}"; do
	[ -e "$config_entry" ] || [ -L "$config_entry" ] || continue
	config_name="${config_entry##*/}"
	if [ "$manage_nats_config" -eq 0 ] && { [ "$config_name" = "nats-server.conf" ] || [ "$config_name" = "nats-server" ]; }; then
		continue
	fi
	cp -a "$config_entry" "$target_dir/"
done
for config_name in mount-monitor.conf antivirus.conf system-service-monitor.conf service-watchdog.conf log-analysis.conf; do
	if [ -f "$security_monitor_saved/$config_name" ]; then
		cp -p "$security_monitor_saved/$config_name" "$target_dir/lite-nas/$config_name"
	fi
done
if [ -d "$security_monitor_saved/log-analysis" ]; then
	cp -a "$security_monitor_saved/log-analysis" "$target_dir/lite-nas/"
fi
log.popTask

# visudo checks the live sudoers include directory ownership during validation.
deploy.normalizeEtcPermissions "$target_dir"

validate_sudoers_drop_ins "$target_dir/sudoers.d"

if [ "$restart_services" -eq 1 ]; then
	deploy.restartAffectedServices
else
	log.info "Skipping affected service restart."
fi

log.info "Config deployment completed."
