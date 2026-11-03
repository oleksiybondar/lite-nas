#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/zfs-metrics.sh"

deploy.entrypoint.run \
	"scripts/deploy-zfs-metrics.sh" \
	"1" \
	"zfs-metrics" \
	"build-zfs-metrics-binary.sh" \
	"deploy.zfsMetrics.requireTools" \
	"deploy.zfsMetrics.deploy" \
	"Deploying zfs-metrics service" \
	"zfs-metrics deployment completed." \
	"$@"
