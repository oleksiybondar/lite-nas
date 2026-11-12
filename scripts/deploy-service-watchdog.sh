#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-monitor.sh"

deploy.service-watchdog.deploy() {
	deploy.securityMonitor.deploy "service-watchdog" "$@"
}

deploy.entrypoint.run \
	"scripts/deploy-service-watchdog.sh" \
	"1" \
	"service-watchdog" \
	"build-service-watchdog-binary.sh" \
	"deploy.securityMonitor.requireTools" \
	"deploy.service-watchdog.deploy" \
	"Deploying service-watchdog service" \
	"service-watchdog deployment completed." \
	"$@"
