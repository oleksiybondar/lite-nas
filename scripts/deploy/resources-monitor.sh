#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_RESOURCES_MONITOR_SERVICE_NAME="${LITE_NAS_RESOURCES_MONITOR_SERVICE_NAME:-lite-nas-resources-monitor}"
readonly LITE_NAS_RESOURCES_MONITOR_RUNTIME_USER="${LITE_NAS_RESOURCES_MONITOR_RUNTIME_USER:-lite-nas-resources-monitor}"
readonly LITE_NAS_RESOURCES_MONITOR_RUNTIME_GROUP="${LITE_NAS_RESOURCES_MONITOR_RUNTIME_GROUP:-$LITE_NAS_RESOURCES_MONITOR_RUNTIME_USER}"
readonly LITE_NAS_RESOURCES_MONITOR_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_RESOURCES_MONITOR_OPERATOR_GROUP="${LITE_NAS_OPERATOR_GROUP:-lite-nas-operator}"
readonly LITE_NAS_RESOURCES_MONITOR_BINARY_TARGET="${LITE_NAS_RESOURCES_MONITOR_BINARY_TARGET:-/usr/libexec/lite-nas/resources-monitor}"
readonly LITE_NAS_RESOURCES_MONITOR_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_RESOURCES_MONITOR_CONFIG_SOURCE="${LITE_NAS_RESOURCES_MONITOR_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/resources-monitor.conf}"
readonly LITE_NAS_RESOURCES_MONITOR_CONFIG_TARGET="${LITE_NAS_RESOURCES_MONITOR_CONFIG_TARGET:-$LITE_NAS_RESOURCES_MONITOR_CONFIG_DIR/resources-monitor.conf}"
readonly LITE_NAS_RESOURCES_MONITOR_RULES_SOURCE_DIR="${LITE_NAS_RESOURCES_MONITOR_RULES_SOURCE_DIR:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/resources-monitor/rules}"
readonly LITE_NAS_RESOURCES_MONITOR_RULES_TARGET_DIR="${LITE_NAS_RESOURCES_MONITOR_RULES_TARGET_DIR:-$LITE_NAS_RESOURCES_MONITOR_CONFIG_DIR/resources-monitor/rules}"
readonly LITE_NAS_RESOURCES_MONITOR_UNIT_TEMPLATE="${LITE_NAS_RESOURCES_MONITOR_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-resources-monitor.service}"
readonly LITE_NAS_RESOURCES_MONITOR_UNIT_TARGET="${LITE_NAS_RESOURCES_MONITOR_UNIT_TARGET:-/etc/systemd/system/lite-nas-resources-monitor.service}"
readonly LITE_NAS_RESOURCES_MONITOR_LOG_DIR="${LITE_NAS_RESOURCES_MONITOR_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_RESOURCES_MONITOR_LOG_FILE="${LITE_NAS_RESOURCES_MONITOR_LOG_FILE:-$LITE_NAS_RESOURCES_MONITOR_LOG_DIR/resources-monitor.log}"

deploy.resourcesMonitor.requireTools() {
	local tool
	local tools=(getent groupadd install chmod chown cp find id useradd usermod systemctl realpath)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.resourcesMonitor.ensureGroup() {
	local group_name="$1"
	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi
	groupadd --system "$group_name"
}

deploy.resourcesMonitor.ensureRuntimeUser() {
	deploy.resourcesMonitor.ensureGroup "$LITE_NAS_RESOURCES_MONITOR_CONFIG_GROUP"
	deploy.resourcesMonitor.ensureGroup "$LITE_NAS_RESOURCES_MONITOR_RUNTIME_GROUP"
	deploy.resourcesMonitor.ensureGroup "$LITE_NAS_RESOURCES_MONITOR_OPERATOR_GROUP"
	if ! id "$LITE_NAS_RESOURCES_MONITOR_RUNTIME_USER" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin \
			--gid "$LITE_NAS_RESOURCES_MONITOR_RUNTIME_GROUP" \
			--groups "$LITE_NAS_RESOURCES_MONITOR_CONFIG_GROUP,$LITE_NAS_RESOURCES_MONITOR_OPERATOR_GROUP" \
			"$LITE_NAS_RESOURCES_MONITOR_RUNTIME_USER"
		return 0
	fi

	usermod \
		--gid "$LITE_NAS_RESOURCES_MONITOR_RUNTIME_GROUP" \
		--append \
		--groups "$LITE_NAS_RESOURCES_MONITOR_CONFIG_GROUP,$LITE_NAS_RESOURCES_MONITOR_OPERATOR_GROUP" \
		"$LITE_NAS_RESOURCES_MONITOR_RUNTIME_USER"
}

deploy.resourcesMonitor.installBinary() {
	local source_binary="$1"
	local target_dir
	if [ ! -f "$source_binary" ]; then
		log.error "Missing resources-monitor binary: $source_binary"
		exit 1
	fi
	target_dir="$(dirname "$LITE_NAS_RESOURCES_MONITOR_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_RESOURCES_MONITOR_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_RESOURCES_MONITOR_BINARY_TARGET"
		return 0
	fi
	install -m 0755 "$source_binary" "$LITE_NAS_RESOURCES_MONITOR_BINARY_TARGET"
}

deploy.resourcesMonitor.installConfig() {
	if [ ! -f "$LITE_NAS_RESOURCES_MONITOR_CONFIG_SOURCE" ]; then
		log.error "Missing resources-monitor config source: $LITE_NAS_RESOURCES_MONITOR_CONFIG_SOURCE"
		exit 1
	fi

	install -d -m 0711 -o root -g "$LITE_NAS_RESOURCES_MONITOR_CONFIG_GROUP" "$LITE_NAS_RESOURCES_MONITOR_CONFIG_DIR"
	install -m 0640 -o root -g "$LITE_NAS_RESOURCES_MONITOR_CONFIG_GROUP" \
		"$LITE_NAS_RESOURCES_MONITOR_CONFIG_SOURCE" \
		"$LITE_NAS_RESOURCES_MONITOR_CONFIG_TARGET"
}

deploy.resourcesMonitor.installRules() {
	local rules_file

	if [ ! -d "$LITE_NAS_RESOURCES_MONITOR_RULES_SOURCE_DIR" ]; then
		log.error "Missing resources-monitor rules source directory: $LITE_NAS_RESOURCES_MONITOR_RULES_SOURCE_DIR"
		exit 1
	fi

	install -d -m 0750 -o root -g "$LITE_NAS_RESOURCES_MONITOR_CONFIG_GROUP" "$LITE_NAS_RESOURCES_MONITOR_RULES_TARGET_DIR"

	while IFS= read -r -d '' rules_file; do
		install -m 0640 -o root -g "$LITE_NAS_RESOURCES_MONITOR_CONFIG_GROUP" \
			"$rules_file" \
			"$LITE_NAS_RESOURCES_MONITOR_RULES_TARGET_DIR/$(basename "$rules_file")"
	done < <(find "$LITE_NAS_RESOURCES_MONITOR_RULES_SOURCE_DIR" -maxdepth 1 -type f -name '*.json' -print0)
}

deploy.resourcesMonitor.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_RESOURCES_MONITOR_CONFIG_GROUP" "$LITE_NAS_RESOURCES_MONITOR_LOG_DIR"
	if [ ! -f "$LITE_NAS_RESOURCES_MONITOR_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_RESOURCES_MONITOR_RUNTIME_USER" -g "$LITE_NAS_RESOURCES_MONITOR_RUNTIME_GROUP" \
			/dev/null "$LITE_NAS_RESOURCES_MONITOR_LOG_FILE"
		return 0
	fi
	chown "$LITE_NAS_RESOURCES_MONITOR_RUNTIME_USER:$LITE_NAS_RESOURCES_MONITOR_RUNTIME_GROUP" "$LITE_NAS_RESOURCES_MONITOR_LOG_FILE"
	chmod 0640 "$LITE_NAS_RESOURCES_MONITOR_LOG_FILE"
}

deploy.resourcesMonitor.installUnitFile() {
	if [ ! -f "$LITE_NAS_RESOURCES_MONITOR_UNIT_TEMPLATE" ]; then
		log.error "Missing resources-monitor systemd unit template: $LITE_NAS_RESOURCES_MONITOR_UNIT_TEMPLATE"
		exit 1
	fi
	install -D -m 0644 "$LITE_NAS_RESOURCES_MONITOR_UNIT_TEMPLATE" "$LITE_NAS_RESOURCES_MONITOR_UNIT_TARGET"
}

deploy.resourcesMonitor.enableAndStart() {
	systemctl daemon-reload
	systemctl enable "$LITE_NAS_RESOURCES_MONITOR_SERVICE_NAME.service"
	systemctl restart "$LITE_NAS_RESOURCES_MONITOR_SERVICE_NAME.service"
}

deploy.resourcesMonitor.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.resourcesMonitor.ensureRuntimeUser
	deploy.resourcesMonitor.installBinary "$source_binary"
	deploy.resourcesMonitor.installConfig
	deploy.resourcesMonitor.installRules
	deploy.resourcesMonitor.installLogTarget
	deploy.resourcesMonitor.installUnitFile

	if [ "$should_start" = "1" ]; then
		deploy.resourcesMonitor.enableAndStart
		return 0
	fi

	log.info "Skipping resources-monitor enable/start."
}
