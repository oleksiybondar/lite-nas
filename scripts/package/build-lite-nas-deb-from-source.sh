#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

source_layout="local-build"
source_root=""
package_arch=""
package_version=""
output_dir=""
admin_panel_assets_path=""
file_explorer_assets_path=""

usage() {
	cat <<'MSG'
Usage: scripts/package/build-lite-nas-deb-from-source.sh [options]

Options:
  --source-layout=LAYOUT    Binary source layout: local-build (default), ci-artifacts, or build-all.
  --source-root=PATH        Root directory for the chosen source layout.
                            Defaults to .build for local-build/build-all and .artifacts for ci-artifacts.
  --package-arch=ARCH       Package architecture (amd64 or arm64). Defaults to the current Go architecture.
  --version=VERSION         Debian package version to pass through.
  --output-dir=PATH         Directory where the package and build root will be written.
  --admin-panel-assets=PATH Override the admin-panel asset directory instead of deriving it from the source layout.
  --file-explorer-assets=PATH Override File Explorer assets instead of deriving them from the source layout.
  -h, --help                Show this help.
MSG
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--source-layout=*)
		source_layout="${1#*=}"
		shift
		;;
	--source-root=*)
		source_root="${1#*=}"
		shift
		;;
	--package-arch=*)
		package_arch="${1#*=}"
		shift
		;;
	--version=*)
		package_version="${1#*=}"
		shift
		;;
	--output-dir=*)
		output_dir="${1#*=}"
		shift
		;;
	--admin-panel-assets=*)
		admin_panel_assets_path="${1#*=}"
		shift
		;;
	--file-explorer-assets=*)
		file_explorer_assets_path="${1#*=}"
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		usage >&2
		exit 64
		;;
	esac
done

if [ -z "$package_arch" ]; then
	package_arch="$(build.resolveTargetArch)"
fi

case "$package_arch" in
amd64 | arm64) ;;
*)
	log.error "Unsupported package architecture: $package_arch"
	exit 64
	;;
esac

case "$source_layout" in
local-build)
	source_root="${source_root:-$LITE_NAS_REPO_ROOT/.build}"
	;;
build-all)
	source_root="${source_root:-$LITE_NAS_REPO_ROOT/.build/all}"
	;;
ci-artifacts)
	source_root="${source_root:-$LITE_NAS_REPO_ROOT/.artifacts}"
	;;
*)
	log.error "Unsupported --source-layout value: $source_layout"
	exit 64
	;;
esac

resolve_binary_path() {
	local component_dir="$1"
	local binary_name="$2"

	case "$source_layout" in
	local-build)
		printf '%s/%s/linux-%s/%s\n' "$source_root" "$component_dir" "$package_arch" "$binary_name"
		;;
	build-all)
		printf '%s/linux-%s/%s\n' "$source_root" "$package_arch" "$binary_name"
		;;
	ci-artifacts)
		printf '%s/%s/%s/%s\n' "$source_root" "$package_arch" "$component_dir" "$binary_name"
		;;
	esac
}

resolve_admin_panel_assets_path() {
	if [ -n "$admin_panel_assets_path" ]; then
		printf '%s\n' "$admin_panel_assets_path"
		return 0
	fi

	if [ "$source_layout" = "build-all" ]; then
		printf '%s/linux-%s/admin-panel\n' "$source_root" "$package_arch"
	else
		printf '%s/admin-panel\n' "$source_root"
	fi
}

resolve_file_explorer_assets_path() {
	if [ -n "$file_explorer_assets_path" ]; then
		printf '%s\n' "$file_explorer_assets_path"
	elif [ "$source_layout" = "build-all" ]; then
		printf '%s/linux-%s/file-explorer\n' "$source_root" "$package_arch"
	else
		printf '%s/file-explorer\n' "$source_root"
	fi
}

require_input_path() {
	local description="$1"
	local path="$2"

	if [ -e "$path" ]; then
		return 0
	fi

	log.error "Missing ${description}: $path"
	exit 1
}

auth_service_binary_path="$(resolve_binary_path auth-service auth-service)"
rbac_service_binary_path="$(resolve_binary_path rbac-service rbac-service)"
system_logging_manager_binary_path="$(resolve_binary_path system-logging-manager system-logging-manager)"
security_logging_manager_binary_path="$(resolve_binary_path security-logging-manager security-logging-manager)"
system_email_notifier_binary_path="$(resolve_binary_path system-email-notifier system-email-notifier)"
security_email_notifier_binary_path="$(resolve_binary_path security-email-notifier security-email-notifier)"
network_metrics_binary_path="$(resolve_binary_path network-metrics network-metrics)"
process_metrics_binary_path="$(resolve_binary_path process-metrics process-metrics)"
disk_metrics_binary_path="$(resolve_binary_path disk-metrics disk-metrics)"
service_metrics_binary_path="$(resolve_binary_path service-metrics service-metrics)"
system_metrics_binary_path="$(resolve_binary_path system-metrics system-metrics)"
zfs_metrics_binary_path="$(resolve_binary_path zfs-metrics zfs-metrics)"
system_logging_manager_cli_binary_path="$(resolve_binary_path system-logging-manager-cli system-logging-manager-cli)"
security_logging_manager_cli_binary_path="$(resolve_binary_path security-logging-manager-cli security-logging-manager-cli)"
system_metrics_cli_binary_path="$(resolve_binary_path system-metrics-cli system-metrics-cli)"
network_metrics_cli_binary_path="$(resolve_binary_path network-metrics-cli network-metrics-cli)"
zfs_metrics_cli_binary_path="$(resolve_binary_path zfs-metrics-cli zfs-metrics-cli)"
file_writer_binary_path="$(resolve_binary_path lite-nas-file-writer lite-nas-file-writer)"
file_service_binary_path="$(resolve_binary_path file-service file-service)"
stream_service_binary_path="$(resolve_binary_path stream-service stream-service)"
web_gateway_binary_path="$(resolve_binary_path web-gateway web-gateway)"
resources_monitor_binary_path="$(resolve_binary_path resources-monitor resources-monitor)"
litenas_integrity_binary_path="$(resolve_binary_path litenas-integrity litenas-integrity)"
mount_monitor_binary_path="$(resolve_binary_path mount-monitor mount-monitor)"
antivirus_binary_path="$(resolve_binary_path antivirus antivirus)"
system_service_monitor_binary_path="$(resolve_binary_path system-service-monitor system-service-monitor)"
service_watchdog_binary_path="$(resolve_binary_path service-watchdog service-watchdog)"
log_analysis_binary_path="$(resolve_binary_path log-analysis log-analysis)"
admin_panel_assets_path="$(resolve_admin_panel_assets_path)"
file_explorer_assets_path="$(resolve_file_explorer_assets_path)"

require_input_path "auth-service binary" "$auth_service_binary_path"
require_input_path "rbac-service binary" "$rbac_service_binary_path"
require_input_path "system-logging-manager binary" "$system_logging_manager_binary_path"
require_input_path "security-logging-manager binary" "$security_logging_manager_binary_path"
require_input_path "system-email-notifier binary" "$system_email_notifier_binary_path"
require_input_path "security-email-notifier binary" "$security_email_notifier_binary_path"
require_input_path "network-metrics binary" "$network_metrics_binary_path"
require_input_path "process-metrics binary" "$process_metrics_binary_path"
require_input_path "disk-metrics binary" "$disk_metrics_binary_path"
require_input_path "service-metrics binary" "$service_metrics_binary_path"
require_input_path "system-metrics binary" "$system_metrics_binary_path"
require_input_path "zfs-metrics binary" "$zfs_metrics_binary_path"
require_input_path "system-logging-manager-cli binary" "$system_logging_manager_cli_binary_path"
require_input_path "security-logging-manager-cli binary" "$security_logging_manager_cli_binary_path"
require_input_path "system-metrics-cli binary" "$system_metrics_cli_binary_path"
require_input_path "network-metrics-cli binary" "$network_metrics_cli_binary_path"
require_input_path "zfs-metrics-cli binary" "$zfs_metrics_cli_binary_path"
require_input_path "file-writer binary" "$file_writer_binary_path"
require_input_path "file-service binary" "$file_service_binary_path"
require_input_path "stream-service binary" "$stream_service_binary_path"
require_input_path "web-gateway binary" "$web_gateway_binary_path"
require_input_path "resources-monitor binary" "$resources_monitor_binary_path"
require_input_path "litenas-integrity binary" "$litenas_integrity_binary_path"
require_input_path "mount-monitor binary" "$mount_monitor_binary_path"
require_input_path "antivirus binary" "$antivirus_binary_path"
require_input_path "system-service-monitor binary" "$system_service_monitor_binary_path"
require_input_path "service-watchdog binary" "$service_watchdog_binary_path"
require_input_path "log-analysis binary" "$log_analysis_binary_path"
require_input_path "admin-panel assets" "$admin_panel_assets_path"
require_input_path "File Explorer assets" "$file_explorer_assets_path"

package_args=(
	--auth-service-binary="$auth_service_binary_path"
	--rbac-service-binary="$rbac_service_binary_path"
	--system-logging-manager-binary="$system_logging_manager_binary_path"
	--security-logging-manager-binary="$security_logging_manager_binary_path"
	--system-email-notifier-binary="$system_email_notifier_binary_path"
	--security-email-notifier-binary="$security_email_notifier_binary_path"
	--network-metrics-binary="$network_metrics_binary_path"
	--process-metrics-binary="$process_metrics_binary_path"
	--disk-metrics-binary="$disk_metrics_binary_path"
	--service-metrics-binary="$service_metrics_binary_path"
	--system-metrics-binary="$system_metrics_binary_path"
	--zfs-metrics-binary="$zfs_metrics_binary_path"
	--system-logging-manager-cli-binary="$system_logging_manager_cli_binary_path"
	--security-logging-manager-cli-binary="$security_logging_manager_cli_binary_path"
	--system-metrics-cli-binary="$system_metrics_cli_binary_path"
	--network-metrics-cli-binary="$network_metrics_cli_binary_path"
	--zfs-metrics-cli-binary="$zfs_metrics_cli_binary_path"
	--file-writer-binary="$file_writer_binary_path"
	--file-service-binary="$file_service_binary_path"
	--stream-service-binary="$stream_service_binary_path"
	--web-gateway-binary="$web_gateway_binary_path"
	--resources-monitor-binary="$resources_monitor_binary_path"
	--litenas-integrity-binary="$litenas_integrity_binary_path"
	--mount-monitor-binary="$mount_monitor_binary_path"
	--antivirus-binary="$antivirus_binary_path"
	--system-service-monitor-binary="$system_service_monitor_binary_path"
	--service-watchdog-binary="$service_watchdog_binary_path"
	--log-analysis-binary="$log_analysis_binary_path"
	--admin-panel-assets="$admin_panel_assets_path"
	--file-explorer-assets="$file_explorer_assets_path"
)

if [ -n "$package_version" ]; then
	package_args+=(--version="$package_version")
fi

if [ -n "$output_dir" ]; then
	package_args+=(--output-dir="$output_dir")
fi

"$SCRIPT_DIR/build-lite-nas-deb.sh" "${package_args[@]}"
