#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-automation-cli.sh"

deploy.securityAutomationCLI.requireTools() {
	local tool
	for tool in getent groupadd id useradd usermod install ln chmod realpath; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.entrypoint.run \
	"scripts/deploy-security-automation-cli.sh" \
	"0" \
	"security-automation-cli" \
	"build-security-automation-cli-binary.sh" \
	"deploy.securityAutomationCLI.requireTools" \
	"deploy.securityAutomationCLI.deploy" \
	"Deploying security-automation-cli" \
	"security-automation-cli deployment completed." \
	"$@"
