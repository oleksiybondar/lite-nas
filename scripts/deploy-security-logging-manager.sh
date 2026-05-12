#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-logging-manager.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"

deploy.entrypoint.run \
	"scripts/deploy-security-logging-manager.sh" \
	1 \
	"security-logging-manager" \
	"build-security-logging-manager-binary.sh" \
	"deploy.securityLoggingManager.requireTools" \
	"deploy.securityLoggingManager.deploy" \
	"Deploying security-logging-manager service" \
	"security-logging-manager deployment completed." \
	"$@"
