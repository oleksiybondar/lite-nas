#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/system-email-notifier.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"

deploy.entrypoint.run \
	"scripts/deploy-system-email-notifier.sh" \
	1 \
	"system-email-notifier" \
	"build-system-email-notifier-binary.sh" \
	"deploy.systemEmailNotifier.requireTools" \
	"deploy.systemEmailNotifier.deploy" \
	"Deploying system-email-notifier service" \
	"system-email-notifier deployment completed." \
	"$@"
