#!/usr/bin/env bash
set -euo pipefail

ENTRYPOINT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/lite-nas.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/rbac-service.sh"
# shellcheck disable=SC1091
source "$ENTRYPOINT_DIR/deploy/entrypoint.sh"

deploy.entrypoint.run \
	"scripts/deploy-rbac-service.sh" \
	"1" \
	"rbac-service" \
	"build-rbac-service-binary.sh" \
	"deploy.rbacService.requireTools" \
	"deploy.rbacService.deploy" \
	"Deploying rbac-service" \
	"rbac-service deployment completed." \
	"$@"
