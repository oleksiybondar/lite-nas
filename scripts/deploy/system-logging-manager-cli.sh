#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_ACCESS_GROUP="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_ACCESS_GROUP:-lite-nas-operator}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_USER="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_USER:-lite-nas-sys-log-mgr-cli}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_GROUP="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_GROUP:-$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_USER}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_BINARY_TARGET="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_BINARY_TARGET:-/usr/libexec/lite-nas/system-logging-manager-cli}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_SYMLINK_TARGET="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_SYMLINK_TARGET:-/usr/bin/system-logging-manager-cli}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_SOURCE="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/system-logging-manager-cli.conf}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_TARGET="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_TARGET:-$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_DIR/system-logging-manager-cli.conf}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_DIR="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_FILE="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_FILE:-$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_DIR/system-logging-manager-cli.log}"

deploy.systemLoggingManagerCLI.requireTools() {
	local tool
	local tools=(getent groupadd install ln chmod chown realpath id useradd usermod)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.systemLoggingManagerCLI.ensureConfigGroup() {
	if getent group "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_GROUP" >/dev/null 2>&1; then
		return
	fi
	groupadd --system "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_GROUP"
}

deploy.systemLoggingManagerCLI.ensureAccessGroup() {
	if getent group "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_ACCESS_GROUP" >/dev/null 2>&1; then
		return
	fi
	groupadd --system "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_ACCESS_GROUP"
}

deploy.systemLoggingManagerCLI.ensureIdentityGroup() {
	if getent group "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_GROUP" >/dev/null 2>&1; then
		return
	fi
	groupadd --system "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_GROUP"
}

deploy.systemLoggingManagerCLI.ensureIdentityUser() {
	deploy.systemLoggingManagerCLI.ensureIdentityGroup

	if ! id "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_USER" >/dev/null 2>&1; then
		useradd \
			--system \
			--no-create-home \
			--home-dir /nonexistent \
			--shell /usr/sbin/nologin \
			--gid "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_GROUP" \
			--groups "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_GROUP,$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_ACCESS_GROUP" \
			"$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_USER"
		return 0
	fi

	usermod \
		--gid "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_GROUP" \
		--append \
		--groups "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_GROUP,$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_ACCESS_GROUP" \
		"$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_USER"
}

deploy.systemLoggingManagerCLI.installBinary() {
	local source_binary="$1"
	local target_dir
	if [ ! -f "$source_binary" ]; then
		log.error "Missing system-logging-manager-cli binary: $source_binary"
		exit 1
	fi
	target_dir="$(dirname "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_BINARY_TARGET"
		return 0
	fi
	install -m 0755 "$source_binary" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_BINARY_TARGET"
}

deploy.systemLoggingManagerCLI.installBinarySymlink() {
	local symlink_dir
	symlink_dir="$(dirname "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_SYMLINK_TARGET")"
	install -d -m 0755 "$symlink_dir"
	ln -sfn "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_BINARY_TARGET" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_SYMLINK_TARGET"
}

deploy.systemLoggingManagerCLI.installConfig() {
	if [ ! -f "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_SOURCE" ]; then
		log.error "Missing system-logging-manager-cli config source: $LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_SOURCE"
		exit 1
	fi
	install -d -m 0711 -o root -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_GROUP" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_DIR"
	install -m 0640 -o "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_USER" -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_ACCESS_GROUP" \
		"$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_SOURCE" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_TARGET"
}

deploy.systemLoggingManagerCLI.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CONFIG_GROUP" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_DIR"
	if [ ! -f "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_FILE" ]; then
		install -m 0666 -o root -g root /dev/null "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_FILE"
		return 0
	fi
	chown root:root "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_FILE"
	chmod 0666 "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_LOG_FILE"
}

deploy.systemLoggingManagerCLI.deploy() {
	local source_binary="$1"
	deploy.systemLoggingManagerCLI.ensureConfigGroup
	deploy.systemLoggingManagerCLI.ensureAccessGroup
	deploy.systemLoggingManagerCLI.ensureIdentityUser
	deploy.systemLoggingManagerCLI.installBinary "$source_binary"
	deploy.systemLoggingManagerCLI.installBinarySymlink
	deploy.systemLoggingManagerCLI.installConfig
	deploy.systemLoggingManagerCLI.installLogTarget
}
