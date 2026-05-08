#!/usr/bin/env bash

if [ -n "${LITE_NAS_GO_MODULES_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_GO_MODULES_LOADED=1

export LITE_NAS_SHARED_GO_MODULE="./shared/go"
readonly LITE_NAS_SHARED_GO_MODULE

export LITE_NAS_SYSTEM_METRICS_MODULE="./services/system-metrics"
readonly LITE_NAS_SYSTEM_METRICS_MODULE

export LITE_NAS_SYSTEM_METRICS_CLI_APP_MODULE="./apps/system-metrics-cli"
readonly LITE_NAS_SYSTEM_METRICS_CLI_APP_MODULE

export LITE_NAS_WEB_GATEWAY_MODULE="./services/web-gateway"
readonly LITE_NAS_WEB_GATEWAY_MODULE

export LITE_NAS_AUTH_SERVICE_MODULE="./services/auth"
readonly LITE_NAS_AUTH_SERVICE_MODULE
