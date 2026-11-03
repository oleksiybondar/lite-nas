#!/usr/bin/env bash

if [ -n "${LITE_NAS_GO_MODULES_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_GO_MODULES_LOADED=1

export LITE_NAS_SHARED_GO_MODULE="./shared/go"
readonly LITE_NAS_SHARED_GO_MODULE

export LITE_NAS_BACKUP_RESTORE_MODULE="./services/backup-restore/cmd/backup-restore"
readonly LITE_NAS_BACKUP_RESTORE_MODULE

export LITE_NAS_BACKUP_RESTORE_CLI_MODULE="./apps/backup-restore-cli"
readonly LITE_NAS_BACKUP_RESTORE_CLI_MODULE

export LITE_NAS_SYSTEM_METRICS_MODULE="./services/system-metrics"
readonly LITE_NAS_SYSTEM_METRICS_MODULE

export LITE_NAS_NETWORK_METRICS_MODULE="./services/network-metrics"
readonly LITE_NAS_NETWORK_METRICS_MODULE

export LITE_NAS_PROCESS_METRICS_MODULE="./services/process-metrics"
readonly LITE_NAS_PROCESS_METRICS_MODULE

export LITE_NAS_DISK_METRICS_MODULE="./services/disk-metrics"
readonly LITE_NAS_DISK_METRICS_MODULE

export LITE_NAS_SERVICE_METRICS_MODULE="./services/service-metrics"
readonly LITE_NAS_SERVICE_METRICS_MODULE

export LITE_NAS_ZFS_METRICS_MODULE="./services/zfs-metrics"
readonly LITE_NAS_ZFS_METRICS_MODULE

export LITE_NAS_ZFS_METRICS_CLI_APP_MODULE="./apps/zfs-metrics-cli"
readonly LITE_NAS_ZFS_METRICS_CLI_APP_MODULE

export LITE_NAS_FILE_WRITER_APP_MODULE="./apps/file-writer"
readonly LITE_NAS_FILE_WRITER_APP_MODULE

export LITE_NAS_FILE_SERVICE_MODULE="./services/file-service"
readonly LITE_NAS_FILE_SERVICE_MODULE

export LITE_NAS_SYSTEM_METRICS_CLI_APP_MODULE="./apps/system-metrics-cli"
readonly LITE_NAS_SYSTEM_METRICS_CLI_APP_MODULE

export LITE_NAS_NETWORK_METRICS_CLI_APP_MODULE="./apps/network-metrics-cli"
readonly LITE_NAS_NETWORK_METRICS_CLI_APP_MODULE

export LITE_NAS_WEB_GATEWAY_MODULE="./services/web-gateway"
readonly LITE_NAS_WEB_GATEWAY_MODULE

export LITE_NAS_MOMENTS_MODULE="./services/moments"
readonly LITE_NAS_MOMENTS_MODULE

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

export LITE_NAS_LITENAS_INTEGRITY_MODULE="./services/litenas-integrity"
readonly LITE_NAS_LITENAS_INTEGRITY_MODULE

export LITE_NAS_RBAC_SERVICE_MODULE="./services/rbac"
readonly LITE_NAS_RBAC_SERVICE_MODULE

export LITE_NAS_MOUNT_MONITOR_MODULE="./services/mount-monitor"
readonly LITE_NAS_MOUNT_MONITOR_MODULE

export LITE_NAS_ANTIVIRUS_MODULE="./services/antivirus"
readonly LITE_NAS_ANTIVIRUS_MODULE

export LITE_NAS_SYSTEM_SERVICE_MONITOR_MODULE="./services/system-service-monitor"
readonly LITE_NAS_SYSTEM_SERVICE_MONITOR_MODULE

export LITE_NAS_SERVICE_WATCHDOG_MODULE="./services/service-watchdog"
readonly LITE_NAS_SERVICE_WATCHDOG_MODULE

export LITE_NAS_SERVICES_MANAGER_MODULE="./services/services-manager"
readonly LITE_NAS_SERVICES_MANAGER_MODULE

export LITE_NAS_SESSION_MANAGER_MODULE="./services/session-manager"
readonly LITE_NAS_SESSION_MANAGER_MODULE

export LITE_NAS_JOURNAL_MANAGER_MODULE="./services/journal-manager"
readonly LITE_NAS_JOURNAL_MANAGER_MODULE

export LITE_NAS_USB_MANAGER_MODULE="./services/usb-manager"
readonly LITE_NAS_USB_MANAGER_MODULE

export LITE_NAS_LOG_ANALYSIS_MODULE="./services/log-analysis"
readonly LITE_NAS_LOG_ANALYSIS_MODULE

export LITE_NAS_NETWORK_SECURITY_MODULE="./services/network-security"
readonly LITE_NAS_NETWORK_SECURITY_MODULE

export LITE_NAS_SECURITY_AUTOMATION_MODULE="./services/security-automation"
readonly LITE_NAS_SECURITY_AUTOMATION_MODULE

export LITE_NAS_SECURITY_AUTOMATION_CLI_MODULE="./apps/security-automation-cli"
readonly LITE_NAS_SECURITY_AUTOMATION_CLI_MODULE
