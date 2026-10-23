#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

test_login="${LITENAS_API_LOGIN:-testuser}"
operator_login="${LITENAS_OPERATOR_LOGIN:-testoperator}"
security_login="${LITENAS_SECURITY_LOGIN:-testsecurityuser}"
admin_login="${LITENAS_FILE_SERVICE_ADMIN_LOGIN:-testadmin}"

default_password="${LITENAS_TEST_PASSWORD:-testpassword}"
test_password="${LITENAS_API_PASSWORD:-$default_password}"
operator_password="${LITENAS_OPERATOR_PASSWORD:-$default_password}"
security_password="${LITENAS_SECURITY_PASSWORD:-$default_password}"
admin_password="${LITENAS_FILE_SERVICE_ADMIN_PASSWORD:-$default_password}"

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
log.requireCommand "groupadd"
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
require_safe_login "$admin_login" "LITENAS_FILE_SERVICE_ADMIN_LOGIN"
case "$admin_login" in
*[!a-zA-Z0-9_.-]* | [-.]*)
	log.error "LITENAS_FILE_SERVICE_ADMIN_LOGIN must be a Unix account name."
	exit 64
	;;
esac
if [ "$admin_login" = "$test_login" ] || [ "$admin_login" = "$operator_login" ] || [ "$admin_login" = "$security_login" ] || [ "$admin_login" = "${LITENAS_TESTSUDO_LOGIN:-testsudouser}" ]; then
	log.error "The File Service admin test account must be separate from the other system-test users."
	exit 64
fi
if [ "$admin_login" = root ] || { id -u "$admin_login" >/dev/null 2>&1 && [ "$(id -u "$admin_login")" -eq 0 ]; }; then
	log.error "The File Service admin test account must be nonroot."
	exit 64
fi
require_safe_password "$test_password" "LITENAS_API_PASSWORD/LITENAS_TEST_PASSWORD"
require_safe_password "$operator_password" "LITENAS_OPERATOR_PASSWORD/LITENAS_TEST_PASSWORD"
require_safe_password "$security_password" "LITENAS_SECURITY_PASSWORD/LITENAS_TEST_PASSWORD"
require_safe_password "$admin_password" "LITENAS_FILE_SERVICE_ADMIN_PASSWORD/LITENAS_TEST_PASSWORD"

log.pushTask "Preparing LiteNAS system-test users"
if ! getent group admin >/dev/null 2>&1; then
	run.asRoot groupadd --system admin
fi
ensure_user "$test_login" "$test_password"
ensure_user "$operator_login" "$operator_password"
ensure_user "$security_login" "$security_password"
ensure_user "$admin_login" "$admin_password"

attach_optional_groups "$test_login" "$test_optional_groups"
attach_optional_groups "$operator_login" "$operator_optional_groups"
attach_optional_groups "$security_login" "$security_optional_groups"
ensure_user_in_group_if_group_exists "$admin_login" admin
log.popTask
