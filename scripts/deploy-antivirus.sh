#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-monitor.sh"

deploy.antivirus.deploy() {
	deploy.securityMonitor.deploy "antivirus" "$@"
}

deploy.entrypoint.run \
	"scripts/deploy-antivirus.sh" \
	"1" \
	"antivirus" \
	"build-antivirus-binary.sh" \
	"deploy.securityMonitor.requireTools" \
	"deploy.antivirus.deploy" \
	"Deploying antivirus service" \
	"antivirus deployment completed." \
	"$@"
