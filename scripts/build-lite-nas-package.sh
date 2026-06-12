#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"$SCRIPT_DIR/build-auth-service-binary.sh"
"$SCRIPT_DIR/build-rbac-service-binary.sh"
"$SCRIPT_DIR/build-system-logging-manager-binary.sh"
"$SCRIPT_DIR/build-security-logging-manager-binary.sh"
"$SCRIPT_DIR/build-system-email-notifier-binary.sh"
"$SCRIPT_DIR/build-security-email-notifier-binary.sh"
"$SCRIPT_DIR/build-resources-monitor-binary.sh"
"$SCRIPT_DIR/build-network-metrics-binary.sh"
"$SCRIPT_DIR/build-system-metrics-binary.sh"
"$SCRIPT_DIR/build-zfs-metrics-binary.sh"
"$SCRIPT_DIR/build-system-logging-manager-cli-binary.sh"
"$SCRIPT_DIR/build-security-logging-manager-cli-binary.sh"
"$SCRIPT_DIR/build-system-metrics-cli-binary.sh"
"$SCRIPT_DIR/build-network-metrics-cli-binary.sh"
"$SCRIPT_DIR/build-zfs-metrics-cli-binary.sh"
"$SCRIPT_DIR/build-web-gateway-binary.sh"
"$SCRIPT_DIR/build-admin-panel.sh"
"$SCRIPT_DIR/package/build-lite-nas-deb-from-source.sh" \
	--source-layout=local-build \
	--source-root="$SCRIPT_DIR/../.build" \
	"$@"
