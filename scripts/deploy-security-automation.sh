#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-monitor.sh"

deploy.securityAutomation.deploy() {
	deploy.securityMonitor.deploy "security-automation" "$@"
}

deploy.entrypoint.run \
	"scripts/deploy-security-automation.sh" \
	"1" \
	"security-automation" \
	"build-security-automation-binary.sh" \
	"deploy.securityMonitor.requireTools" \
	"deploy.securityAutomation.deploy" \
	"Deploying security-automation service" \
	"security-automation deployment completed." \
	"$@"
