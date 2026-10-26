#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-monitor.sh"

deploy.mount-monitor.deploy() {
	deploy.securityMonitor.deploy "mount-monitor" "$@"
}

deploy.entrypoint.run \
	"scripts/deploy-mount-monitor.sh" \
	"1" \
	"mount-monitor" \
	"build-mount-monitor-binary.sh" \
	"deploy.securityMonitor.requireTools" \
	"deploy.mount-monitor.deploy" \
	"Deploying mount-monitor service" \
	"mount-monitor deployment completed." \
	"$@"
