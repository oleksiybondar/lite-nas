#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_SERVICE_NAME="${LITE_NAS_SYSTEM_LOGGING_MANAGER_SERVICE_NAME:-lite-nas-system-logging-manager}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER="${LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER:-lite-nas-sys-log-mgr}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP="${LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP:-$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_ACCESS_GROUP="${LITE_NAS_SYSTEM_LOGGING_MANAGER_ACCESS_GROUP:-lite-nas-operator}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_BINARY_TARGET="${LITE_NAS_SYSTEM_LOGGING_MANAGER_BINARY_TARGET:-/usr/libexec/lite-nas/system-logging-manager}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_SOURCE="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/system-logging-manager.conf}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_TARGET="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_TARGET:-$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_DIR/system-logging-manager.conf}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_UNIT_TEMPLATE="${LITE_NAS_SYSTEM_LOGGING_MANAGER_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-system-logging-manager.service}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_UNIT_TARGET="${LITE_NAS_SYSTEM_LOGGING_MANAGER_UNIT_TARGET:-/etc/systemd/system/lite-nas-system-logging-manager.service}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_DIR="${LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_FILE="${LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_FILE:-$LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_DIR/system-logging-manager.log}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_LEGACY_LOG_FILE="/var/lib/lite-nas/system-logging-manager.log"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_DIR="${LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_DIR:-/var/lib/lite-nas/system-logging-manager}"
readonly LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_FILE="${LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_FILE:-$LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_DIR/log.db}"

deploy.systemLoggingManager.requireTools() {
	local tool
	local tools=(getent groupadd install cat chmod chown realpath rm systemctl touch useradd usermod)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.systemLoggingManager.ensureGroup() {
	local group_name="$1"
	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi
	groupadd --system "$group_name"
}

deploy.systemLoggingManager.ensureRuntimeUser() {
	deploy.systemLoggingManager.ensureGroup "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_GROUP"
	deploy.systemLoggingManager.ensureGroup "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP"
	deploy.systemLoggingManager.ensureGroup "$LITE_NAS_SYSTEM_LOGGING_MANAGER_ACCESS_GROUP"

	if ! id "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin \
			--gid "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP" \
			--groups "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_GROUP,$LITE_NAS_SYSTEM_LOGGING_MANAGER_ACCESS_GROUP" \
			"$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER"
		return 0
	fi

	usermod --gid "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP" --append \
		--groups "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_GROUP,$LITE_NAS_SYSTEM_LOGGING_MANAGER_ACCESS_GROUP" \
		"$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER"
}

deploy.systemLoggingManager.installBinary() {
	local source_binary="$1"
	local target_dir
	if [ ! -f "$source_binary" ]; then
		log.error "Missing system-logging-manager binary: $source_binary"
		exit 1
	fi
	target_dir="$(dirname "$LITE_NAS_SYSTEM_LOGGING_MANAGER_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_SYSTEM_LOGGING_MANAGER_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_SYSTEM_LOGGING_MANAGER_BINARY_TARGET"
		return 0
	fi
	install -m 0755 "$source_binary" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_BINARY_TARGET"
}

deploy.systemLoggingManager.installConfig() {
	if [ ! -f "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_SOURCE" ]; then
		log.error "Missing system-logging-manager config source: $LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_SOURCE"
		exit 1
	fi
	install -d -m 0711 -o root -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_GROUP" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_DIR"
	install -m 0640 -o root -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_GROUP" \
		"$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_SOURCE" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_TARGET"
}

deploy.systemLoggingManager.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_CONFIG_GROUP" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_DIR"
	if [ -f "$LITE_NAS_SYSTEM_LOGGING_MANAGER_LEGACY_LOG_FILE" ]; then
		if [ ! -f "$LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_FILE" ]; then
			install -m 0640 -o "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER" -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP" \
				"$LITE_NAS_SYSTEM_LOGGING_MANAGER_LEGACY_LOG_FILE" \
				"$LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_FILE"
		else
			cat "$LITE_NAS_SYSTEM_LOGGING_MANAGER_LEGACY_LOG_FILE" >>"$LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_FILE"
		fi
		rm -f "$LITE_NAS_SYSTEM_LOGGING_MANAGER_LEGACY_LOG_FILE"
	fi
	if [ ! -f "$LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER" -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP" \
			/dev/null "$LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_FILE"
		return 0
	fi
	chown "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER:$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_FILE"
	chmod 0640 "$LITE_NAS_SYSTEM_LOGGING_MANAGER_LOG_FILE"
}

deploy.systemLoggingManager.installDBTarget() {
	install -d -m 0700 -o "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER" -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP" \
		"$LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_DIR"
	if [ ! -f "$LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_FILE" ]; then
		install -m 0600 -o "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER" -g "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP" \
			/dev/null "$LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_FILE"
		return 0
	fi
	chown "$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER:$LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_GROUP" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_FILE"
	chmod 0600 "$LITE_NAS_SYSTEM_LOGGING_MANAGER_DB_FILE"
}

deploy.systemLoggingManager.installUnitFile() {
	if [ ! -f "$LITE_NAS_SYSTEM_LOGGING_MANAGER_UNIT_TEMPLATE" ]; then
		log.error "Missing systemd unit template: $LITE_NAS_SYSTEM_LOGGING_MANAGER_UNIT_TEMPLATE"
		exit 1
	fi
	install -D -m 0644 "$LITE_NAS_SYSTEM_LOGGING_MANAGER_UNIT_TEMPLATE" "$LITE_NAS_SYSTEM_LOGGING_MANAGER_UNIT_TARGET"
}

deploy.systemLoggingManager.enableAndStart() {
	systemctl daemon-reload
	systemctl enable "$LITE_NAS_SYSTEM_LOGGING_MANAGER_SERVICE_NAME.service"
	systemctl restart "$LITE_NAS_SYSTEM_LOGGING_MANAGER_SERVICE_NAME.service"
}

deploy.systemLoggingManager.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"
	deploy.systemLoggingManager.ensureRuntimeUser
	deploy.systemLoggingManager.installBinary "$source_binary"
	deploy.systemLoggingManager.installConfig
	deploy.systemLoggingManager.installLogTarget
	deploy.systemLoggingManager.installDBTarget
	deploy.systemLoggingManager.installUnitFile
	if [ "$should_start" = "1" ]; then
		deploy.systemLoggingManager.enableAndStart
		return 0
	fi
	log.info "Skipping system-logging-manager enable/start."
}
