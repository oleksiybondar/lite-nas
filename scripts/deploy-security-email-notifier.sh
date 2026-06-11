#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/security-email-notifier.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"

deploy.entrypoint.run \
	"scripts/deploy-security-email-notifier.sh" \
	1 \
	"security-email-notifier" \
	"build-security-email-notifier-binary.sh" \
	"deploy.securityEmailNotifier.requireTools" \
	"deploy.securityEmailNotifier.deploy" \
	"Deploying security-email-notifier service" \
	"security-email-notifier deployment completed." \
	"$@"
