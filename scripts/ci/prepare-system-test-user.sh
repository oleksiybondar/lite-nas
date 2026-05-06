#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

test_login="${LITENAS_API_LOGIN:-testuser}"
test_password="${LITENAS_API_PASSWORD:-testpassword}"

if [ -z "$test_login" ]; then
	log.error "LITENAS_API_LOGIN must not be empty."
	exit 64
fi

if [ -z "$test_password" ]; then
	log.error "LITENAS_API_PASSWORD must not be empty."
	exit 64
fi

case "$test_login" in
*:*)
	log.error "LITENAS_API_LOGIN must not contain ':'."
	exit 64
	;;
esac

case "$test_login" in
*'
'*)
	log.error "LITENAS_API_LOGIN must not contain newlines."
	exit 64
	;;
esac

case "$test_password" in
*'
'*)
	log.error "LITENAS_API_PASSWORD must not contain newlines."
	exit 64
	;;
esac

log.requireCommand "id"
log.requireCommand "useradd"
log.requireCommand "chpasswd"

if [ "${EUID:-$(id -u)}" -ne 0 ]; then
	log.requireCommand "sudo"
fi

run.asRoot() {
	if [ "${EUID:-$(id -u)}" -eq 0 ]; then
		"$@"
		return
	fi

	sudo "$@"
}

log.pushTask "Preparing LiteNAS system-test login user"
if id "$test_login" >/dev/null 2>&1; then
	log.info "System-test user already exists: $test_login"
else
	run.asRoot useradd --create-home --shell /bin/bash "$test_login"
	log.info "Created system-test user: $test_login"
fi

printf '%s:%s\n' "$test_login" "$test_password" | run.asRoot chpasswd
log.info "Configured password for system-test user: $test_login"
log.popTask
