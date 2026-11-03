#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/disk-metrics.sh"

deploy.entrypoint.run \
	"scripts/deploy-disk-metrics.sh" \
	"1" \
	"disk-metrics" \
	"build-disk-metrics-binary.sh" \
	"deploy.diskMetrics.requireTools" \
	"deploy.diskMetrics.deploy" \
	"Deploying disk-metrics service" \
	"disk-metrics deployment completed." \
	"$@"
