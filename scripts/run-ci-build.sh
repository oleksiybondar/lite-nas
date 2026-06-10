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
scripts/build-rbac-service.sh
scripts/build-system-logging-manager.sh
scripts/build-security-logging-manager.sh
scripts/build-system-email-notifier.sh
scripts/build-security-email-notifier.sh
scripts/build-resources-monitor.sh
scripts/build-system-metrics.sh
scripts/build-zfs-metrics.sh
scripts/build-system-logging-manager-cli.sh
scripts/build-security-logging-manager-cli.sh
scripts/build-system-metrics-cli.sh
scripts/build-zfs-metrics-cli.sh
scripts/build-shared.sh
scripts/build-web-gateway.sh
scripts/build-admin-panel.sh
log.popTask
