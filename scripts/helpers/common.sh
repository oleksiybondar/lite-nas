#!/usr/bin/env bash

if [ -n "${LITE_NAS_COMMON_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_COMMON_LOADED=1

COMMON_SOURCE="${BASH_SOURCE[0]}"
if command -v realpath >/dev/null 2>&1; then
	COMMON_SOURCE="$(realpath "$COMMON_SOURCE")"
fi

COMMON_DIR="$(cd "$(dirname "$COMMON_SOURCE")" && pwd)"
export LITE_NAS_REPO_ROOT="${LITE_NAS_REPO_ROOT:-$(cd "$COMMON_DIR/../.." && pwd)}"

# shellcheck disable=SC1091
source "$COMMON_DIR/logger.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/sudo-guard.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/tool-paths.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/go-modules.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/args.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/build.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/packaging.sh"

deploy.enableAndRefreshService() {
	local service="$1"

	if command -v systemctl >/dev/null 2>&1; then
		systemctl daemon-reload
		systemctl enable "$service"

		if systemctl is-active --quiet "$service"; then
			systemctl restart "$service"
			return 0
		fi

		systemctl start "$service"
		return 0
	fi

	if command -v service >/dev/null 2>&1; then
		service "$service" restart
		return 0
	fi

	log.error "Missing service manager; cannot enable/start $service."
	exit 1
}
