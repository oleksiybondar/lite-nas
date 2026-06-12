#!/usr/bin/env bash

if [ -n "${LITE_NAS_GO_MODULES_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_GO_MODULES_LOADED=1

export LITE_NAS_SHARED_GO_MODULE="./shared/go"
readonly LITE_NAS_SHARED_GO_MODULE

export LITE_NAS_SYSTEM_METRICS_MODULE="./services/system-metrics"
readonly LITE_NAS_SYSTEM_METRICS_MODULE

export LITE_NAS_NETWORK_METRICS_MODULE="./services/network-metrics"
readonly LITE_NAS_NETWORK_METRICS_MODULE

export LITE_NAS_ZFS_METRICS_MODULE="./services/zfs-metrics"
readonly LITE_NAS_ZFS_METRICS_MODULE

export LITE_NAS_ZFS_METRICS_CLI_APP_MODULE="./apps/zfs-metrics-cli"
readonly LITE_NAS_ZFS_METRICS_CLI_APP_MODULE

export LITE_NAS_SYSTEM_METRICS_CLI_APP_MODULE="./apps/system-metrics-cli"
readonly LITE_NAS_SYSTEM_METRICS_CLI_APP_MODULE

export LITE_NAS_NETWORK_METRICS_CLI_APP_MODULE="./apps/network-metrics-cli"
readonly LITE_NAS_NETWORK_METRICS_CLI_APP_MODULE

export LITE_NAS_WEB_GATEWAY_MODULE="./services/web-gateway"
readonly LITE_NAS_WEB_GATEWAY_MODULE

export LITE_NAS_AUTH_SERVICE_MODULE="./services/auth"
readonly LITE_NAS_AUTH_SERVICE_MODULE

export LITE_NAS_SYSTEM_LOGGING_MANAGER_MODULE="./services/system-logging-manager"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_MODULE

export LITE_NAS_SECURITY_LOGGING_MANAGER_MODULE="./services/security-logging-manager"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_MODULE

export LITE_NAS_SYSTEM_EMAIL_NOTIFIER_MODULE="./services/system-email-notifier"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_MODULE

export LITE_NAS_SECURITY_EMAIL_NOTIFIER_MODULE="./services/security-email-notifier"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_MODULE

export LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_APP_MODULE="./apps/system-logging-manager-cli"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_APP_MODULE

export LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_APP_MODULE="./apps/security-logging-manager-cli"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_APP_MODULE

export LITE_NAS_RESOURCES_MONITOR_MODULE="./services/resources-monitor"
readonly LITE_NAS_RESOURCES_MONITOR_MODULE

export LITE_NAS_RBAC_SERVICE_MODULE="./services/rbac"
readonly LITE_NAS_RBAC_SERVICE_MODULE
