#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_ACCESS_GROUP="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_ACCESS_GROUP:-lite-nas-security}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_USER="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_USER:-lite-nas-sec-log-mgr-cli}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_GROUP="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_GROUP:-$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_USER}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_BINARY_TARGET="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_BINARY_TARGET:-/usr/libexec/lite-nas/security-logging-manager-cli}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_SYMLINK_TARGET="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_SYMLINK_TARGET:-/usr/bin/security-logging-manager-cli}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_SOURCE="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/security-logging-manager-cli.conf}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_TARGET="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_TARGET:-$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_DIR/security-logging-manager-cli.conf}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_DIR="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_FILE="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_FILE:-$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_DIR/security-logging-manager-cli.log}"

deploy.securityLoggingManagerCLI.requireTools() {
	local tool
	local tools=(getent groupadd install ln chmod chown realpath id useradd usermod)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.securityLoggingManagerCLI.ensureConfigGroup() {
	if getent group "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_GROUP" >/dev/null 2>&1; then
		return
	fi
	groupadd --system "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_GROUP"
}

deploy.securityLoggingManagerCLI.ensureAccessGroup() {
	if getent group "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_ACCESS_GROUP" >/dev/null 2>&1; then
		return
	fi
	groupadd --system "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_ACCESS_GROUP"
}

deploy.securityLoggingManagerCLI.ensureIdentityGroup() {
	if getent group "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_GROUP" >/dev/null 2>&1; then
		return
	fi
	groupadd --system "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_GROUP"
}

deploy.securityLoggingManagerCLI.ensureIdentityUser() {
	deploy.securityLoggingManagerCLI.ensureIdentityGroup

	if ! id "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_USER" >/dev/null 2>&1; then
		useradd \
			--system \
			--no-create-home \
			--home-dir /nonexistent \
			--shell /usr/sbin/nologin \
			--gid "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_GROUP" \
			--groups "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_GROUP,$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_ACCESS_GROUP" \
			"$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_USER"
		return 0
	fi

	usermod \
		--gid "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_GROUP" \
		--append \
		--groups "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_GROUP,$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_ACCESS_GROUP" \
		"$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_USER"
}

deploy.securityLoggingManagerCLI.installBinary() {
	local source_binary="$1"
	local target_dir
	if [ ! -f "$source_binary" ]; then
		log.error "Missing security-logging-manager-cli binary: $source_binary"
		exit 1
	fi
	target_dir="$(dirname "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_BINARY_TARGET"
		return 0
	fi
	install -m 0755 "$source_binary" "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_BINARY_TARGET"
}

deploy.securityLoggingManagerCLI.installBinarySymlink() {
	local symlink_dir
	symlink_dir="$(dirname "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_SYMLINK_TARGET")"
	install -d -m 0755 "$symlink_dir"
	ln -sfn "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_BINARY_TARGET" "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_SYMLINK_TARGET"
}

deploy.securityLoggingManagerCLI.installConfig() {
	if [ ! -f "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_SOURCE" ]; then
		log.error "Missing security-logging-manager-cli config source: $LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_SOURCE"
		exit 1
	fi
	install -d -m 0711 -o root -g "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_GROUP" "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_DIR"
	install -m 0640 -o "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_USER" -g "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_ACCESS_GROUP" \
		"$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_SOURCE" "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_TARGET"
}

deploy.securityLoggingManagerCLI.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CONFIG_GROUP" "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_DIR"
	if [ ! -f "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_FILE" ]; then
		install -m 0666 -o root -g root /dev/null "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_FILE"
		return 0
	fi
	chown root:root "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_FILE"
	chmod 0666 "$LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_LOG_FILE"
}

deploy.securityLoggingManagerCLI.deploy() {
	local source_binary="$1"
	deploy.securityLoggingManagerCLI.ensureConfigGroup
	deploy.securityLoggingManagerCLI.ensureAccessGroup
	deploy.securityLoggingManagerCLI.ensureIdentityUser
	deploy.securityLoggingManagerCLI.installBinary "$source_binary"
	deploy.securityLoggingManagerCLI.installBinarySymlink
	deploy.securityLoggingManagerCLI.installConfig
	deploy.securityLoggingManagerCLI.installLogTarget
}
