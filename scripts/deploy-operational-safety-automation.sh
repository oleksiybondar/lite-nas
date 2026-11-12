#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-monitor.sh"

deploy.operationalSafetyAutomation.deploy() {
	deploy.securityMonitor.deploy "operational-safety-automation" "$@"
}

deploy.entrypoint.run \
	"scripts/deploy-operational-safety-automation.sh" \
	"1" \
	"operational-safety-automation" \
	"build-operational-safety-automation-binary.sh" \
	"deploy.securityMonitor.requireTools" \
	"deploy.operationalSafetyAutomation.deploy" \
	"Deploying operational-safety-automation service" \
	"operational-safety-automation deployment completed." \
	"$@"
