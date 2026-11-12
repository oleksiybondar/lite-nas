#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/litenas-integrity.sh"

deploy.entrypoint.run \
	"scripts/deploy-litenas-integrity.sh" \
	"1" \
	"litenas-integrity" \
	"build-litenas-integrity-binary.sh" \
	"deploy.liteNASIntegrity.requireTools" \
	"deploy.liteNASIntegrity.deploy" \
	"Deploying litenas-integrity service" \
	"litenas-integrity deployment completed." \
	"$@"
