#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"
log.pushTask "Running local CI build checks"
scripts/build-auth-service.sh
scripts/build-system-metrics.sh
scripts/build-system-metrics-cli.sh
scripts/build-shared.sh
scripts/build-web-gateway.sh
log.popTask
