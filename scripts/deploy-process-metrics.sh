#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/process-metrics.sh"

deploy.entrypoint.run \
	"scripts/deploy-process-metrics.sh" \
	"1" \
	"process-metrics" \
	"build-process-metrics-binary.sh" \
	"deploy.processMetrics.requireTools" \
	"deploy.processMetrics.deploy" \
	"Deploying process-metrics service" \
	"process-metrics deployment completed." \
	"$@"
