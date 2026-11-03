#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/service-metrics.sh"

deploy.entrypoint.run \
	"scripts/deploy-service-metrics.sh" \
	"1" \
	"service-metrics" \
	"build-service-metrics-binary.sh" \
	"deploy.serviceMetrics.requireTools" \
	"deploy.serviceMetrics.deploy" \
	"Deploying service-metrics service" \
	"service-metrics deployment completed." \
	"$@"
