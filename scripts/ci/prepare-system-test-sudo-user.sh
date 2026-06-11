#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

testsudo_login="${LITENAS_TESTSUDO_LOGIN:-testsudouser}"
default_password="${LITENAS_TEST_PASSWORD:-testpassword}"
testsudo_password="${LITENAS_TESTSUDO_PASSWORD:-$default_password}"
sudoers_file="${LITENAS_TESTSUDO_SUDOERS_FILE:-/etc/sudoers.d/lite-nas-tests-aa-status}"

aa_status_path="${LITENAS_TESTSUDO_AA_STATUS_PATH:-/usr/sbin/aa-status}"
system_logging_manager_cli_path="${LITENAS_SYSTEM_LOGGING_MANAGER_CLI_BINARY:-/usr/bin/system-logging-manager-cli}"
security_logging_manager_cli_path="${LITENAS_SECURITY_LOGGING_MANAGER_CLI_BINARY:-/usr/bin/security-logging-manager-cli}"

require_safe_login() {
	local login="$1"
	local name="$2"

	if [ -z "$login" ]; then
		log.error "$name must not be empty."
		exit 64
	fi
	case "$login" in
	*:* | *$'\n'*)
		log.error "$name contains unsupported characters."
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
	*$'\n'*)
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
		log.info "System-test sudo user already exists: $login"
	else
		run.asRoot useradd --create-home --shell /bin/bash "$login"
		log.info "Created system-test sudo user: $login"
	fi

	printf '%s:%s\n' "$login" "$password" | run.asRoot chpasswd
	log.info "Configured password for system-test sudo user: $login"
}

install_sudoers_rule() {
	local tmp_file
	tmp_file="$(mktemp)"

	cat >"$tmp_file" <<EOF
# Managed by scripts/ci/prepare-system-test-sudo-user.sh for CI system tests.
${testsudo_login} ALL=(root) NOPASSWD: ${aa_status_path}, ${system_logging_manager_cli_path}, ${security_logging_manager_cli_path}
EOF

	run.asRoot install -m 0440 -o root -g root "$tmp_file" "$sudoers_file"
	rm -f "$tmp_file"
	log.info "Installed sudoers rule: $sudoers_file"
}

log.requireCommand "id"
log.requireCommand "useradd"
log.requireCommand "chpasswd"
log.requireCommand "install"
log.requireCommand "mktemp"

if [ "${EUID:-$(id -u)}" -ne 0 ]; then
	log.requireCommand "sudo"
fi

require_safe_login "$testsudo_login" "LITENAS_TESTSUDO_LOGIN"
require_safe_password "$testsudo_password" "LITENAS_TESTSUDO_PASSWORD/LITENAS_TEST_PASSWORD"

log.pushTask "Preparing restricted sudo system-test user"
ensure_user "$testsudo_login" "$testsudo_password"
install_sudoers_rule
log.popTask
