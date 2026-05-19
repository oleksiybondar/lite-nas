#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/resources-monitor.sh"

deploy.entrypoint.run \
	"scripts/deploy-resources-monitor.sh" \
	"1" \
	"resources-monitor" \
	"build-resources-monitor-binary.sh" \
	"deploy.resourcesMonitor.requireTools" \
	"deploy.resourcesMonitor.deploy" \
	"Deploying resources-monitor service" \
	"resources-monitor deployment completed." \
	"$@"
