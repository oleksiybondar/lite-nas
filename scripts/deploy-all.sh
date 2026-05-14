#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/auth-service.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/system-logging-manager.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-logging-manager.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/system-metrics.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/system-logging-manager-cli.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-logging-manager-cli.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/system-metrics-cli.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/web-gateway.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/resources-monitor.sh"

auth_service_binary=""
system_logging_manager_binary=""
security_logging_manager_binary=""
system_metrics_binary=""
system_logging_manager_cli_binary=""
security_logging_manager_cli_binary=""
system_metrics_cli_binary=""
web_gateway_binary=""
resources_monitor_binary=""
should_bootstrap=1
manage_nats_config=1
should_start=1

usage() {
	cat <<'MSG'
Usage: scripts/deploy-all.sh [options]

Options:
  --auth-service-binary PATH      Install an existing auth-service binary.
  --system-logging-manager-binary PATH
                                  Install an existing system-logging-manager binary.
  --security-logging-manager-binary PATH
                                  Install an existing security-logging-manager binary.
  --system-metrics-binary PATH    Install an existing system-metrics binary.
  --system-logging-manager-cli-binary PATH
                                  Install an existing system-logging-manager-cli binary.
  --security-logging-manager-cli-binary PATH
                                  Install an existing security-logging-manager-cli binary.
  --system-metrics-cli-binary PATH
                                  Install an existing system-metrics-cli binary.
  --web-gateway-binary PATH       Install an existing web-gateway binary.
  --resources-monitor-binary PATH Install an existing resources-monitor binary.
  --no-start                      Install files but do not enable or start services.
  --skip-bootstrap                Skip shared config, certificates, and NATS restart.
  --no-nats-config                Keep the current NATS configuration unchanged during bootstrap.
  -h, --help                      Show this help.
MSG
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--auth-service-binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --auth-service-binary"
			usage >&2
			exit 2
		fi
		auth_service_binary="$2"
		shift 2
		;;
	--system-metrics-binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --system-metrics-binary"
			usage >&2
			exit 2
		fi
		system_metrics_binary="$2"
		shift 2
		;;
	--system-logging-manager-binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --system-logging-manager-binary"
			usage >&2
			exit 2
		fi
		system_logging_manager_binary="$2"
		shift 2
		;;
	--security-logging-manager-binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --security-logging-manager-binary"
			usage >&2
			exit 2
		fi
		security_logging_manager_binary="$2"
		shift 2
		;;
	--system-logging-manager-cli-binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --system-logging-manager-cli-binary"
			usage >&2
			exit 2
		fi
		system_logging_manager_cli_binary="$2"
		shift 2
		;;
	--security-logging-manager-cli-binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --security-logging-manager-cli-binary"
			usage >&2
			exit 2
		fi
		security_logging_manager_cli_binary="$2"
		shift 2
		;;
	--system-metrics-cli-binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --system-metrics-cli-binary"
			usage >&2
			exit 2
		fi
		system_metrics_cli_binary="$2"
		shift 2
		;;
	--web-gateway-binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --web-gateway-binary"
			usage >&2
			exit 2
		fi
		web_gateway_binary="$2"
		shift 2
		;;
	--resources-monitor-binary)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --resources-monitor-binary"
			usage >&2
			exit 2
		fi
		resources_monitor_binary="$2"
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

sudo.guard.requireRoot "scripts/deploy-all.sh"
deploy.liteNAS.requireTools
deploy.authService.requireTools
deploy.systemLoggingManager.requireTools
deploy.securityLoggingManager.requireTools
deploy.systemMetrics.requireTools
deploy.systemLoggingManagerCLI.requireTools
deploy.securityLoggingManagerCLI.requireTools
deploy.systemMetricsCLI.requireTools
deploy.webGateway.requireTools
deploy.resourcesMonitor.requireTools

tmp_dir=""
if [ -z "$auth_service_binary" ] || [ -z "$system_logging_manager_binary" ] || [ -z "$security_logging_manager_binary" ] || [ -z "$system_metrics_binary" ] || [ -z "$system_logging_manager_cli_binary" ] || [ -z "$security_logging_manager_cli_binary" ] || [ -z "$system_metrics_cli_binary" ] || [ -z "$web_gateway_binary" ]; then
	tmp_dir="$(mktemp -d)"
	trap 'rm -rf "$tmp_dir"' EXIT

	if [ -z "$auth_service_binary" ]; then
		build_args=("--output=$tmp_dir/auth-service")
		"$ENTRYPOINT_DIR/build-auth-service-binary.sh" "${build_args[@]}"
		auth_service_binary="$tmp_dir/auth-service"
	fi

	if [ -z "$system_logging_manager_binary" ]; then
		build_args=("--output=$tmp_dir/system-logging-manager")
		"$ENTRYPOINT_DIR/build-system-logging-manager-binary.sh" "${build_args[@]}"
		system_logging_manager_binary="$tmp_dir/system-logging-manager"
	fi

	if [ -z "$security_logging_manager_binary" ]; then
		build_args=("--output=$tmp_dir/security-logging-manager")
		"$ENTRYPOINT_DIR/build-security-logging-manager-binary.sh" "${build_args[@]}"
		security_logging_manager_binary="$tmp_dir/security-logging-manager"
	fi

	if [ -z "$system_metrics_binary" ]; then
		build_args=("--output=$tmp_dir/system-metrics")
		"$ENTRYPOINT_DIR/build-system-metrics-binary.sh" "${build_args[@]}"
		system_metrics_binary="$tmp_dir/system-metrics"
	fi

	if [ -z "$system_logging_manager_cli_binary" ]; then
		build_args=("--output=$tmp_dir/system-logging-manager-cli")
		"$ENTRYPOINT_DIR/build-system-logging-manager-cli-binary.sh" "${build_args[@]}"
		system_logging_manager_cli_binary="$tmp_dir/system-logging-manager-cli"
	fi

	if [ -z "$security_logging_manager_cli_binary" ]; then
		build_args=("--output=$tmp_dir/security-logging-manager-cli")
		"$ENTRYPOINT_DIR/build-security-logging-manager-cli-binary.sh" "${build_args[@]}"
		security_logging_manager_cli_binary="$tmp_dir/security-logging-manager-cli"
	fi

	if [ -z "$system_metrics_cli_binary" ]; then
		build_args=("--output=$tmp_dir/system-metrics-cli")
		"$ENTRYPOINT_DIR/build-system-metrics-cli-binary.sh" "${build_args[@]}"
		system_metrics_cli_binary="$tmp_dir/system-metrics-cli"
	fi

	if [ -z "$web_gateway_binary" ]; then
		build_args=("--output=$tmp_dir/web-gateway")
		"$ENTRYPOINT_DIR/build-web-gateway-binary.sh" "${build_args[@]}"
		web_gateway_binary="$tmp_dir/web-gateway"
	fi
fi

if [ "$should_bootstrap" -eq 1 ]; then
	log.pushTask "Preparing shared LiteNAS infrastructure"
	deploy.liteNAS.bootstrap "$manage_nats_config"
	log.popTask
else
	log.info "Skipping shared LiteNAS bootstrap."
fi

log.pushTask "Deploying auth-service"
deploy.authService.deploy "$auth_service_binary" "$should_start"
log.popTask

log.pushTask "Deploying system-logging-manager service"
deploy.systemLoggingManager.deploy "$system_logging_manager_binary" "$should_start"
log.popTask

log.pushTask "Deploying security-logging-manager service"
deploy.securityLoggingManager.deploy "$security_logging_manager_binary" "$should_start"
log.popTask

log.pushTask "Deploying system-metrics service"
deploy.systemMetrics.deploy "$system_metrics_binary" "$should_start"
log.popTask

log.pushTask "Deploying web-gateway service"
deploy.webGateway.deploy "$web_gateway_binary" "$should_start"
log.popTask

log.pushTask "Deploying system-logging-manager-cli app"
deploy.systemLoggingManagerCLI.deploy "$system_logging_manager_cli_binary"
log.popTask

log.pushTask "Deploying security-logging-manager-cli app"
deploy.securityLoggingManagerCLI.deploy "$security_logging_manager_cli_binary"
log.popTask

log.pushTask "Deploying system-metrics-cli app"
deploy.systemMetricsCLI.deploy "$system_metrics_cli_binary"
log.popTask

if [ -n "$resources_monitor_binary" ]; then
	log.pushTask "Deploying resources-monitor service"
	deploy.resourcesMonitor.deploy "$resources_monitor_binary" "$should_start"
	log.popTask
else
	log.info "Skipping resources-monitor deploy (no binary provided)."
fi

log.info "LiteNAS local deployment completed."
