#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/system-logging-manager-cli.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"

deploy.entrypoint.run \
	"scripts/deploy-system-logging-manager-cli.sh" \
	0 \
	"system-logging-manager-cli" \
	"build-system-logging-manager-cli-binary.sh" \
	"deploy.systemLoggingManagerCLI.requireTools" \
	"deploy.systemLoggingManagerCLI.deploy" \
	"Deploying system-logging-manager-cli app" \
	"system-logging-manager-cli deployment completed." \
	"$@"
