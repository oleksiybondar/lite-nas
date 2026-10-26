#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/tool-paths.sh"

cd "$(git rev-parse --show-toplevel)"
log.pushTask "Running local component build checks"
scripts/build-auth-service.sh
scripts/build-rbac-service.sh
scripts/build-system-logging-manager.sh
scripts/build-security-logging-manager.sh
scripts/build-system-email-notifier.sh
scripts/build-security-email-notifier.sh
scripts/build-resources-monitor.sh
scripts/build-mount-monitor.sh
scripts/build-antivirus.sh
scripts/build-system-service-monitor.sh
scripts/build-service-watchdog.sh
scripts/build-log-analysis.sh
scripts/build-network-metrics.sh
scripts/build-process-metrics.sh
scripts/build-disk-metrics.sh
scripts/build-service-metrics.sh
scripts/build-system-metrics.sh
scripts/build-zfs-metrics.sh
scripts/build-system-logging-manager-cli.sh
scripts/build-security-logging-manager-cli.sh
scripts/build-system-metrics-cli.sh
scripts/build-network-metrics-cli.sh
scripts/build-zfs-metrics-cli.sh
scripts/build-file-writer-binary.sh
scripts/build-shared.sh
scripts/build-web-gateway.sh
scripts/ci/go-build-module.sh ./services/moments
scripts/build-admin-panel.sh
scripts/build-file-explorer.sh
scripts/build-moments.sh
log.popTask
