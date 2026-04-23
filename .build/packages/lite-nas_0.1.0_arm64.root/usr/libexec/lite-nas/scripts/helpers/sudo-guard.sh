#!/usr/bin/env bash

if [ -n "${LITE_NAS_SUDO_GUARD_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_SUDO_GUARD_LOADED=1

sudo.guard.requireRoot() {
	local script_name="${1:-This script}"

	if [ "$(id -u)" -eq 0 ]; then
		return 0
	fi

	if declare -F log.error >/dev/null 2>&1; then
		log.error "$script_name must be run with sudo or as root."
	else
		printf 'ERROR: %s must be run with sudo or as root.\n' "$script_name" >&2
	fi

	exit 1
}
