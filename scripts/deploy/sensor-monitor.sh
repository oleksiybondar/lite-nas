#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_SENSOR_MONITOR_SERVICE_NAME="${LITE_NAS_SENSOR_MONITOR_SERVICE_NAME:-lite-nas-sensor-monitor}"
readonly LITE_NAS_SENSOR_MONITOR_RUNTIME_USER="${LITE_NAS_SENSOR_MONITOR_RUNTIME_USER:-lite-nas-sensor-monitor}"
readonly LITE_NAS_SENSOR_MONITOR_RUNTIME_GROUP="${LITE_NAS_SENSOR_MONITOR_RUNTIME_GROUP:-$LITE_NAS_SENSOR_MONITOR_RUNTIME_USER}"
readonly LITE_NAS_SENSOR_MONITOR_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_SENSOR_MONITOR_BINARY_TARGET="${LITE_NAS_SENSOR_MONITOR_BINARY_TARGET:-/usr/libexec/lite-nas/sensor-monitor}"
readonly LITE_NAS_SENSOR_MONITOR_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_SENSOR_MONITOR_CONFIG_SOURCE="${LITE_NAS_SENSOR_MONITOR_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/sensor-monitor.conf}"
readonly LITE_NAS_SENSOR_MONITOR_CONFIG_TARGET="${LITE_NAS_SENSOR_MONITOR_CONFIG_TARGET:-$LITE_NAS_SENSOR_MONITOR_CONFIG_DIR/sensor-monitor.conf}"
readonly LITE_NAS_SENSOR_MONITOR_NAMES_SOURCE="${LITE_NAS_SENSOR_MONITOR_NAMES_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/sensor-names.json}"
readonly LITE_NAS_SENSOR_MONITOR_NAMES_TARGET="${LITE_NAS_SENSOR_MONITOR_NAMES_TARGET:-$LITE_NAS_SENSOR_MONITOR_CONFIG_DIR/sensor-names.json}"
readonly LITE_NAS_SENSOR_MONITOR_UNIT_TEMPLATE="${LITE_NAS_SENSOR_MONITOR_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-sensor-monitor.service}"
readonly LITE_NAS_SENSOR_MONITOR_UNIT_TARGET="${LITE_NAS_SENSOR_MONITOR_UNIT_TARGET:-/etc/systemd/system/lite-nas-sensor-monitor.service}"
readonly LITE_NAS_SENSOR_MONITOR_UDEV_RULE_SOURCE="${LITE_NAS_SENSOR_MONITOR_UDEV_RULE_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/udev/rules.d/99-lite-nas-vcio.rules}"
readonly LITE_NAS_SENSOR_MONITOR_UDEV_RULE_TARGET="${LITE_NAS_SENSOR_MONITOR_UDEV_RULE_TARGET:-/etc/udev/rules.d/99-lite-nas-vcio.rules}"
readonly LITE_NAS_SENSOR_MONITOR_LOG_DIR="${LITE_NAS_SENSOR_MONITOR_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_SENSOR_MONITOR_LOG_FILE="${LITE_NAS_SENSOR_MONITOR_LOG_FILE:-$LITE_NAS_SENSOR_MONITOR_LOG_DIR/sensor-monitor.log}"

deploy.sensorMonitor.usage() {
	cat <<'MSG'
Usage: scripts/deploy-sensor-monitor.sh [options]

Options:
  --binary PATH       Install an existing binary instead of building one.
  --no-start          Install files but do not enable or start the service.
  --skip-bootstrap    Install files without running LiteNAS bootstrap first.
  -h, --help          Show this help.
MSG
}

deploy.sensorMonitor.requireTools() {
	local tool
	local tools=(
		getent
		groupadd
		install
		cat
		chmod
		chown
		realpath
		systemctl
		useradd
		usermod
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.sensorMonitor.ensureGroup() {
	local group_name="$1"

	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating system group: $group_name"
	groupadd --system "$group_name"
}

deploy.sensorMonitor.ensureRuntimeUser() {
	local supplementary_groups="$LITE_NAS_SENSOR_MONITOR_CONFIG_GROUP"
	if getent group video >/dev/null 2>&1; then
		supplementary_groups+=",video"
	fi
	deploy.sensorMonitor.ensureGroup "$LITE_NAS_SENSOR_MONITOR_CONFIG_GROUP"
	deploy.sensorMonitor.ensureGroup "$LITE_NAS_SENSOR_MONITOR_RUNTIME_GROUP"

	if ! id "$LITE_NAS_SENSOR_MONITOR_RUNTIME_USER" >/dev/null 2>&1; then
		log.info "Creating system user: $LITE_NAS_SENSOR_MONITOR_RUNTIME_USER"
		useradd \
			--system \
			--no-create-home \
			--home-dir /nonexistent \
			--shell /usr/sbin/nologin \
			--gid "$LITE_NAS_SENSOR_MONITOR_RUNTIME_GROUP" \
			--groups "$supplementary_groups" \
			"$LITE_NAS_SENSOR_MONITOR_RUNTIME_USER"
		return 0
	fi

	log.info "Updating system user groups: $LITE_NAS_SENSOR_MONITOR_RUNTIME_USER"
	usermod \
		--gid "$LITE_NAS_SENSOR_MONITOR_RUNTIME_GROUP" \
		--append \
		--groups "$supplementary_groups" \
		"$LITE_NAS_SENSOR_MONITOR_RUNTIME_USER"
}

deploy.sensorMonitor.installBinary() {
	local source_binary="$1"
	local target_dir

	if [ ! -f "$source_binary" ]; then
		log.error "Missing sensor-monitor binary: $source_binary"
		exit 1
	fi

	target_dir="$(dirname "$LITE_NAS_SENSOR_MONITOR_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"

	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_SENSOR_MONITOR_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_SENSOR_MONITOR_BINARY_TARGET"
		return 0
	fi

	install -m 0755 "$source_binary" "$LITE_NAS_SENSOR_MONITOR_BINARY_TARGET"
}

deploy.sensorMonitor.installConfig() {
	if [ ! -f "$LITE_NAS_SENSOR_MONITOR_CONFIG_SOURCE" ]; then
		log.error "Missing sensor-monitor config source: $LITE_NAS_SENSOR_MONITOR_CONFIG_SOURCE"
		exit 1
	fi

	install -d -m 0711 -o root -g "$LITE_NAS_SENSOR_MONITOR_CONFIG_GROUP" "$LITE_NAS_SENSOR_MONITOR_CONFIG_DIR"
	deploy.installMergedConfig "$LITE_NAS_SENSOR_MONITOR_CONFIG_SOURCE" "$LITE_NAS_SENSOR_MONITOR_CONFIG_TARGET" "$LITE_NAS_SENSOR_MONITOR_RUNTIME_USER" "$LITE_NAS_SENSOR_MONITOR_CONFIG_GROUP"
	if [ ! -f "$LITE_NAS_SENSOR_MONITOR_NAMES_TARGET" ]; then
		install -m 0644 -o root -g "$LITE_NAS_SENSOR_MONITOR_CONFIG_GROUP" "$LITE_NAS_SENSOR_MONITOR_NAMES_SOURCE" "$LITE_NAS_SENSOR_MONITOR_NAMES_TARGET"
	else
		python3 "$LITE_NAS_REPO_ROOT/scripts/deploy/merge-json-config.py" \
			"$LITE_NAS_SENSOR_MONITOR_NAMES_TARGET" "$LITE_NAS_SENSOR_MONITOR_NAMES_SOURCE"
	fi
}

deploy.sensorMonitor.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_SENSOR_MONITOR_CONFIG_GROUP" "$LITE_NAS_SENSOR_MONITOR_LOG_DIR"

	if [ ! -f "$LITE_NAS_SENSOR_MONITOR_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_SENSOR_MONITOR_RUNTIME_USER" -g "$LITE_NAS_SENSOR_MONITOR_RUNTIME_GROUP" \
			/dev/null \
			"$LITE_NAS_SENSOR_MONITOR_LOG_FILE"
		return 0
	fi

	chown "$LITE_NAS_SENSOR_MONITOR_RUNTIME_USER:$LITE_NAS_SENSOR_MONITOR_RUNTIME_GROUP" "$LITE_NAS_SENSOR_MONITOR_LOG_FILE"
	chmod 0640 "$LITE_NAS_SENSOR_MONITOR_LOG_FILE"
}

deploy.sensorMonitor.installUnitFile() {
	if [ ! -f "$LITE_NAS_SENSOR_MONITOR_UNIT_TEMPLATE" ]; then
		log.error "Missing systemd unit template: $LITE_NAS_SENSOR_MONITOR_UNIT_TEMPLATE"
		exit 1
	fi

	install -D -m 0644 "$LITE_NAS_SENSOR_MONITOR_UNIT_TEMPLATE" "$LITE_NAS_SENSOR_MONITOR_UNIT_TARGET"
}

deploy.sensorMonitor.installUdevRule() {
	if [ ! -f "$LITE_NAS_SENSOR_MONITOR_UDEV_RULE_SOURCE" ]; then
		log.error "Missing sensor-monitor udev rule: $LITE_NAS_SENSOR_MONITOR_UDEV_RULE_SOURCE"
		exit 1
	fi

	install -D -o root -g root -m 0644 \
		"$LITE_NAS_SENSOR_MONITOR_UDEV_RULE_SOURCE" "$LITE_NAS_SENSOR_MONITOR_UDEV_RULE_TARGET"
}

deploy.sensorMonitor.enableAndStart() {
	if [ -e /dev/vcio ]; then
		log.requireCommand "udevadm" "Install udev tooling and retry."
		udevadm control --reload-rules
		udevadm trigger --action=change --name-match=/dev/vcio --settle
	fi
	deploy.enableAndRefreshService "$LITE_NAS_SENSOR_MONITOR_SERVICE_NAME.service"
}

deploy.sensorMonitor.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.sensorMonitor.ensureRuntimeUser
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing --user "$LITE_NAS_SENSOR_MONITOR_RUNTIME_USER"
	deploy.sensorMonitor.installBinary "$source_binary"
	deploy.sensorMonitor.installConfig
	deploy.sensorMonitor.installLogTarget
	deploy.sensorMonitor.installUnitFile
	deploy.sensorMonitor.installUdevRule

	if [ "$should_start" = "1" ]; then
		deploy.sensorMonitor.enableAndStart
		return 0
	fi

	log.info "Skipping sensor-monitor enable/start."
}
