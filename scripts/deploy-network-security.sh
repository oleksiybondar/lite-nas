#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-monitor.sh"

deploy.network-security.deploy() {
	deploy.securityMonitor.deploy "network-security" "$@"
}

deploy.entrypoint.run \
	"scripts/deploy-network-security.sh" \
	"1" \
	"network-security" \
	"build-network-security-binary.sh" \
	"deploy.securityMonitor.requireTools" \
	"deploy.network-security.deploy" \
	"Deploying network-security service" \
	"network-security deployment completed." \
	"$@"
