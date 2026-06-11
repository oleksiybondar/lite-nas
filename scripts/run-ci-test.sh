#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"
log.pushTask "Running local CI test checks"
scripts/test-auth-service.sh
scripts/test-system-metrics.sh --with-coverage
scripts/test-system-metrics-cli.sh
scripts/test-zfs-metrics-cli.sh
scripts/test-shared.sh --with-coverage
scripts/test-web-gateway.sh
scripts/test-admin-panel.sh --with-coverage
scripts/test-python.sh
log.popTask
