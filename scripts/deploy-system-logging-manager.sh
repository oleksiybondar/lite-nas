#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/system-logging-manager.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"

deploy.entrypoint.run \
	"scripts/deploy-system-logging-manager.sh" \
	1 \
	"system-logging-manager" \
	"build-system-logging-manager-binary.sh" \
	"deploy.systemLoggingManager.requireTools" \
	"deploy.systemLoggingManager.deploy" \
	"Deploying system-logging-manager service" \
	"system-logging-manager deployment completed." \
	"$@"
