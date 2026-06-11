#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

test_login="${LITENAS_API_LOGIN:-testuser}"
operator_login="${LITENAS_OPERATOR_LOGIN:-testoperator}"
security_login="${LITENAS_SECURITY_LOGIN:-testsecurity}"

default_password="${LITENAS_TEST_PASSWORD:-testpassword}"
test_password="${LITENAS_API_PASSWORD:-$default_password}"
operator_password="${LITENAS_OPERATOR_PASSWORD:-$default_password}"
security_password="${LITENAS_SECURITY_PASSWORD:-$default_password}"

test_optional_groups="${LITENAS_TESTUSER_OPTIONAL_GROUPS:-lite-nas}"
operator_optional_groups="${LITENAS_TESTOPERATOR_OPTIONAL_GROUPS:-lite-nas-operator}"
security_optional_groups="${LITENAS_TESTSECURITYUSER_OPTIONAL_GROUPS:-lite-nas-security}"

require_safe_login() {
	local login="$1"
	local name="$2"

	if [ -z "$login" ]; then
		log.error "$name must not be empty."
		exit 64
	fi

	case "$login" in
	*:*)
		log.error "$name must not contain ':'."
		exit 64
		;;
	esac

	case "$login" in
	*'
'*)
		log.error "$name must not contain newlines."
		exit 64
		;;
	esac
}

require_safe_password() {
	local password="$1"
	local name="$2"

	if [ -z "$password" ]; then
		log.error "$name must not be empty."
		exit 64
	fi

	case "$password" in
	*'
'*)
		log.error "$name must not contain newlines."
		exit 64
		;;
	esac
}

run.asRoot() {
	if [ "${EUID:-$(id -u)}" -eq 0 ]; then
		"$@"
		return
	fi

	sudo "$@"
}

ensure_user() {
	local login="$1"
	local password="$2"

	if id "$login" >/dev/null 2>&1; then
		log.info "System-test user already exists: $login"
	else
		run.asRoot useradd --create-home --shell /bin/bash "$login"
		log.info "Created system-test user: $login"
	fi

	printf '%s:%s\n' "$login" "$password" | run.asRoot chpasswd
	log.info "Configured password for system-test user: $login"
}

ensure_user_in_group_if_group_exists() {
	local login="$1"
	local group="$2"

	if [ -z "$group" ]; then
		return 0
	fi

	if ! getent group "$group" >/dev/null 2>&1; then
		log.info "Skipping optional group membership for $login; group does not exist: $group"
		return 0
	fi

	if id -nG "$login" | tr ' ' '\n' | grep -Fx "$group" >/dev/null 2>&1; then
		log.info "System-test user already in group: $login -> $group"
		return 0
	fi

	run.asRoot usermod -a -G "$group" "$login"
	log.info "Added system-test user to group: $login -> $group"
}

attach_optional_groups() {
	local login="$1"
	local groups="$2"
	local group=""

	for group in $groups; do
		ensure_user_in_group_if_group_exists "$login" "$group"
	done
}

log.requireCommand "id"
log.requireCommand "getent"
log.requireCommand "useradd"
log.requireCommand "usermod"
log.requireCommand "chpasswd"
log.requireCommand "grep"
log.requireCommand "tr"

if [ "${EUID:-$(id -u)}" -ne 0 ]; then
	log.requireCommand "sudo"
fi

require_safe_login "$test_login" "LITENAS_API_LOGIN"
require_safe_login "$operator_login" "LITENAS_OPERATOR_LOGIN"
require_safe_login "$security_login" "LITENAS_SECURITY_LOGIN"
require_safe_password "$test_password" "LITENAS_API_PASSWORD/LITENAS_TEST_PASSWORD"
require_safe_password "$operator_password" "LITENAS_OPERATOR_PASSWORD/LITENAS_TEST_PASSWORD"
require_safe_password "$security_password" "LITENAS_SECURITY_PASSWORD/LITENAS_TEST_PASSWORD"

log.pushTask "Preparing LiteNAS system-test users"
ensure_user "$test_login" "$test_password"
ensure_user "$operator_login" "$operator_password"
ensure_user "$security_login" "$security_password"

attach_optional_groups "$test_login" "$test_optional_groups"
attach_optional_groups "$operator_login" "$operator_optional_groups"
attach_optional_groups "$security_login" "$security_optional_groups"
log.popTask
