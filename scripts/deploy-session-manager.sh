#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/host-manager.sh"

deploy.entrypoint.run \
	"scripts/deploy-session-manager.sh" "1" "session-manager" \
	"build-session-manager-binary.sh" "deploy.hostManager.requireTools" \
	"deploy.hostManager.deploy" "Deploying session-manager service" \
	"session-manager deployment completed." "$@"
