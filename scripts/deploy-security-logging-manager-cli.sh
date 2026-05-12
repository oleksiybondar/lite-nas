#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-logging-manager-cli.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"

deploy.entrypoint.run \
	"scripts/deploy-security-logging-manager-cli.sh" \
	0 \
	"security-logging-manager-cli" \
	"build-security-logging-manager-cli-binary.sh" \
	"deploy.securityLoggingManagerCLI.requireTools" \
	"deploy.securityLoggingManagerCLI.deploy" \
	"Deploying security-logging-manager-cli app" \
	"security-logging-manager-cli deployment completed." \
	"$@"
