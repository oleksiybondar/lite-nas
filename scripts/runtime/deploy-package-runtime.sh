#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/auth-service.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/system-metrics.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/system-metrics-cli.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/web-gateway.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/restart-affected-services.sh"

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
deploy.authService.requireTools
deploy.systemMetrics.requireTools
deploy.systemMetricsCLI.requireTools
deploy.webGateway.requireTools

if [ "$run_mode" = "validate" ]; then
	log.pushTask "Deploying LiteNAS package runtime in validate mode"
	export LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE="$PACKAGE_ROOT/admin-panel-assets"
	deploy.authService.deploy "$PACKAGE_ROOT/auth-service" 0
	deploy.systemMetrics.deploy "$PACKAGE_ROOT/system-metrics" 0
	deploy.systemMetricsCLI.deploy "$PACKAGE_ROOT/system-metrics-cli"
	deploy.webGateway.deploy "$PACKAGE_ROOT/web-gateway" 0
	log.popTask
	log.info "LiteNAS package runtime validation deployment completed."
	exit 0
fi

log.pushTask "Bootstrapping LiteNAS runtime prerequisites"
deploy.liteNAS.bootstrap "$manage_nats_config"
log.popTask

export LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE="$PACKAGE_ROOT/admin-panel-assets"

log.pushTask "Deploying LiteNAS runtime files without service start"
deploy.authService.deploy "$PACKAGE_ROOT/auth-service" 0
deploy.systemMetrics.deploy "$PACKAGE_ROOT/system-metrics" 0
deploy.systemMetricsCLI.deploy "$PACKAGE_ROOT/system-metrics-cli"
deploy.webGateway.deploy "$PACKAGE_ROOT/web-gateway" 0
log.popTask

log.pushTask "Restarting dependency services"
deploy.restartAffectedServices
log.popTask

log.pushTask "Starting LiteNAS services"
deploy.authService.enableAndStart
deploy.systemMetrics.enableAndStart
deploy.webGateway.enableAndStart
log.popTask

log.info "LiteNAS package runtime deployment completed."
