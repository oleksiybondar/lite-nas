#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-monitor.sh"

deploy.log-analysis.deploy() {
	deploy.securityMonitor.deploy "log-analysis" "$@"
}

deploy.entrypoint.run \
	"scripts/deploy-log-analysis.sh" \
	"1" \
	"log-analysis" \
	"build-log-analysis-binary.sh" \
	"deploy.securityMonitor.requireTools" \
	"deploy.log-analysis.deploy" \
	"Deploying log-analysis service" \
	"log-analysis deployment completed." \
	"$@"
