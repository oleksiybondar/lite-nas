#!/usr/bin/env bash
set -eu

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/apparmor.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/postfix.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/auth-service.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/rbac-service.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/system-logging-manager.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/security-logging-manager.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/system-email-notifier.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/security-email-notifier.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/network-metrics.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/process-metrics.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/disk-metrics.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/service-metrics.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/system-metrics.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/zfs-metrics.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/system-logging-manager-cli.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/security-logging-manager-cli.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/system-metrics-cli.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/network-metrics-cli.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/zfs-metrics-cli.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/file-writer.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/file-service.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/stream-service.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/web-gateway.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/moments.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/resources-monitor.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/litenas-integrity.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/security-monitor.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/restart-affected-services.sh"

# Package maintainer execution should not inherit pipefail from sourced helpers.
# Some platform tools intentionally close pipes early, which is benign during install.
set +o pipefail

run_mode="full"
manage_nats_config=1

usage() {
	cat <<'MSG'
Usage: scripts/runtime/deploy-package-runtime.sh [options]

Options:
  --mode MODE         Runtime mode: full (default) or validate.
  --no-nats-config    Keep current NATS configuration unchanged during bootstrap.
  -h, --help          Show this help.
MSG
}

deploy_runtime_files_without_start() {
	local load_writer_profile="$1"
	deploy.authService.deploy "$PACKAGE_ROOT/auth-service" 0
	deploy.rbacService.deploy "$PACKAGE_ROOT/rbac-service" 0
	deploy.systemLoggingManager.deploy "$PACKAGE_ROOT/system-logging-manager" 0
	deploy.securityLoggingManager.deploy "$PACKAGE_ROOT/security-logging-manager" 0
	deploy.systemEmailNotifier.deploy "$PACKAGE_ROOT/system-email-notifier" 0
	deploy.securityEmailNotifier.deploy "$PACKAGE_ROOT/security-email-notifier" 0
	deploy.networkMetrics.deploy "$PACKAGE_ROOT/network-metrics" 0
	deploy.processMetrics.deploy "$PACKAGE_ROOT/process-metrics" 0
	deploy.diskMetrics.deploy "$PACKAGE_ROOT/disk-metrics" 0
	deploy.serviceMetrics.deploy "$PACKAGE_ROOT/service-metrics" 0
	deploy.systemMetrics.deploy "$PACKAGE_ROOT/system-metrics" 0
	deploy.zfsMetrics.deploy "$PACKAGE_ROOT/zfs-metrics" 0
	deploy.systemLoggingManagerCLI.deploy "$PACKAGE_ROOT/system-logging-manager-cli"
	deploy.securityLoggingManagerCLI.deploy "$PACKAGE_ROOT/security-logging-manager-cli"
	deploy.systemMetricsCLI.deploy "$PACKAGE_ROOT/system-metrics-cli"
	deploy.networkMetricsCLI.deploy "$PACKAGE_ROOT/network-metrics-cli"
	deploy.zfsMetricsCLI.deploy "$PACKAGE_ROOT/zfs-metrics-cli"
	deploy.fileWriter.deploy "$PACKAGE_ROOT/lite-nas-file-writer" "$load_writer_profile"
	deploy.fileService.deploy "$PACKAGE_ROOT/file-service" 0
	deploy.streamService.deploy "$PACKAGE_ROOT/stream-service" 0
	deploy.webGateway.deploy "$PACKAGE_ROOT/web-gateway" 0
	deploy.moments.deploy "$PACKAGE_ROOT/moments" "$PACKAGE_ROOT/moments-assets" 0
	deploy.resourcesMonitor.deploy "$PACKAGE_ROOT/resources-monitor" 0
	deploy.liteNASIntegrity.deploy "$PACKAGE_ROOT/litenas-integrity" 0
	deploy.securityMonitor.deploy "mount-monitor" "$PACKAGE_ROOT/mount-monitor" 0
	deploy.securityMonitor.deploy "antivirus" "$PACKAGE_ROOT/antivirus" 0
	deploy.securityMonitor.deploy "system-service-monitor" "$PACKAGE_ROOT/system-service-monitor" 0
	deploy.securityMonitor.deploy "service-watchdog" "$PACKAGE_ROOT/service-watchdog" 0
	deploy.securityMonitor.deploy "log-analysis" "$PACKAGE_ROOT/log-analysis" 0
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--mode)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --mode"
			usage >&2
			exit 2
		fi
		run_mode="$2"
		shift 2
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

case "$run_mode" in
full | validate) ;;
*)
	log.error "Unsupported --mode value: $run_mode"
	usage >&2
	exit 2
	;;
esac

sudo.guard.requireRoot "scripts/runtime/deploy-package-runtime.sh"

deploy.liteNAS.requireTools
deploy.apparmor.requireTools
deploy.authService.requireTools
deploy.rbacService.requireTools
deploy.systemLoggingManager.requireTools
deploy.securityLoggingManager.requireTools
deploy.systemEmailNotifier.requireTools
deploy.securityEmailNotifier.requireTools
deploy.networkMetrics.requireTools
deploy.processMetrics.requireTools
deploy.diskMetrics.requireTools
deploy.systemMetrics.requireTools
deploy.zfsMetrics.requireTools
deploy.systemLoggingManagerCLI.requireTools
deploy.securityLoggingManagerCLI.requireTools
deploy.systemMetricsCLI.requireTools
deploy.networkMetricsCLI.requireTools
deploy.zfsMetricsCLI.requireTools
deploy.fileWriter.requireTools
deploy.fileService.requireTools
deploy.streamService.requireTools
deploy.webGateway.requireTools
deploy.resourcesMonitor.requireTools
deploy.liteNASIntegrity.requireTools
deploy.securityMonitor.requireTools

if [ "$run_mode" = "validate" ]; then
	log.pushTask "Deploying LiteNAS package runtime in validate mode"
	export LITE_NAS_PACKAGE_VALIDATE=1
	export LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE="$PACKAGE_ROOT/admin-panel-assets"
	deploy.apparmor.deploy 0
	deploy.postfix.deploy 0
	deploy_runtime_files_without_start 0
	deploy.normalizeEtcPermissions /etc
	log.popTask
	log.info "LiteNAS package runtime validation deployment completed."
	exit 0
fi

log.pushTask "Bootstrapping LiteNAS runtime prerequisites"
deploy.liteNAS.bootstrap "$manage_nats_config"
log.popTask

export LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE="$PACKAGE_ROOT/admin-panel-assets"

log.pushTask "Deploying LiteNAS runtime files without service start"
deploy_runtime_files_without_start 1
log.popTask

log.pushTask "Normalizing deployed LiteNAS permissions"
deploy.normalizeEtcPermissions /etc
log.popTask

log.pushTask "Restarting dependency services"
deploy.restartAffectedServices
log.popTask

log.pushTask "Starting LiteNAS services"
deploy.authService.enableAndStart
deploy.rbacService.enableAndStart
deploy.systemLoggingManager.enableAndStart
deploy.securityLoggingManager.enableAndStart
deploy.systemEmailNotifier.enableAndStart
deploy.securityEmailNotifier.enableAndStart
deploy.networkMetrics.enableAndStart
deploy.processMetrics.enableAndStart
deploy.diskMetrics.enableAndStart
deploy.systemMetrics.enableAndStart
deploy.zfsMetrics.enableAndStart
deploy.webGateway.enableAndStart
deploy.moments.enableAndStart
deploy.fileService.enableAndStart
deploy.streamService.enableAndStart
deploy.resourcesMonitor.enableAndStart
deploy.liteNASIntegrity.enableAndStart
deploy.enableAndRefreshService "lite-nas-mount-monitor.service"
deploy.enableAndRefreshService "lite-nas-antivirus.service"
deploy.enableAndRefreshService "lite-nas-system-service-monitor.service"
deploy.enableAndRefreshService "lite-nas-service-watchdog.service"
deploy.securityMonitor.refreshLogAnalysis
log.popTask

log.info "LiteNAS package runtime deployment completed."
