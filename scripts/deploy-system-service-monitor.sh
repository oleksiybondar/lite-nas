#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-monitor.sh"

deploy.system-service-monitor.deploy() {
	deploy.securityMonitor.deploy "system-service-monitor" "$@"
}

deploy.entrypoint.run \
	"scripts/deploy-system-service-monitor.sh" \
	"1" \
	"system-service-monitor" \
	"build-system-service-monitor-binary.sh" \
	"deploy.securityMonitor.requireTools" \
	"deploy.system-service-monitor.deploy" \
	"Deploying system-service-monitor service" \
	"system-service-monitor deployment completed." \
	"$@"
