#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_SYSTEM_METRICS_SERVICE_NAME="${LITE_NAS_SYSTEM_METRICS_SERVICE_NAME:-lite-nas-system-metrics}"
readonly LITE_NAS_SYSTEM_METRICS_RUNTIME_USER="${LITE_NAS_SYSTEM_METRICS_RUNTIME_USER:-lite-nas-system-metrics}"
readonly LITE_NAS_SYSTEM_METRICS_RUNTIME_GROUP="${LITE_NAS_SYSTEM_METRICS_RUNTIME_GROUP:-$LITE_NAS_SYSTEM_METRICS_RUNTIME_USER}"
readonly LITE_NAS_SYSTEM_METRICS_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_SYSTEM_METRICS_BINARY_TARGET="${LITE_NAS_SYSTEM_METRICS_BINARY_TARGET:-/usr/libexec/lite-nas/system-metrics}"
readonly LITE_NAS_SYSTEM_METRICS_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/liteNAS}"
readonly LITE_NAS_SYSTEM_METRICS_CONFIG_SOURCE="${LITE_NAS_SYSTEM_METRICS_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/liteNAS/system-metrics.conf}"
readonly LITE_NAS_SYSTEM_METRICS_CONFIG_TARGET="${LITE_NAS_SYSTEM_METRICS_CONFIG_TARGET:-$LITE_NAS_SYSTEM_METRICS_CONFIG_DIR/system-metrics.conf}"
readonly LITE_NAS_SYSTEM_METRICS_UNIT_TEMPLATE="${LITE_NAS_SYSTEM_METRICS_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/systemd/system/lite-nas-system-metrics.service}"
readonly LITE_NAS_SYSTEM_METRICS_UNIT_TARGET="${LITE_NAS_SYSTEM_METRICS_UNIT_TARGET:-/lib/systemd/system/lite-nas-system-metrics.service}"
readonly LITE_NAS_SYSTEM_METRICS_LOG_DIR="${LITE_NAS_SYSTEM_METRICS_LOG_DIR:-/var/log/liteNAS}"
readonly LITE_NAS_SYSTEM_METRICS_LOG_FILE="${LITE_NAS_SYSTEM_METRICS_LOG_FILE:-$LITE_NAS_SYSTEM_METRICS_LOG_DIR/system-metrics.log}"

deploy.systemMetrics.usage() {
	cat <<'MSG'
Usage: scripts/deploy-system-metrics.sh [options]

Options:
  --binary PATH       Install an existing binary instead of building one.
  --arch=amd64|arm64  Build target architecture when --binary is not set.
  --no-start          Install files but do not enable or start the service.
  --skip-bootstrap    Install files without running LiteNAS bootstrap first.
  -h, --help          Show this help.
MSG
}

deploy.systemMetrics.requireTools() {
	local tool
	local tools=(
		getent
		groupadd
		install
		chmod
		chown
		realpath
		sed
		systemctl
		touch
		useradd
		usermod
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.systemMetrics.ensureGroup() {
	local group_name="$1"

	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating system group: $group_name"
	groupadd --system "$group_name"
}

deploy.systemMetrics.ensureRuntimeUser() {
	deploy.systemMetrics.ensureGroup "$LITE_NAS_SYSTEM_METRICS_CONFIG_GROUP"
	deploy.systemMetrics.ensureGroup "$LITE_NAS_SYSTEM_METRICS_RUNTIME_GROUP"

	if ! id "$LITE_NAS_SYSTEM_METRICS_RUNTIME_USER" >/dev/null 2>&1; then
		log.info "Creating system user: $LITE_NAS_SYSTEM_METRICS_RUNTIME_USER"
		useradd \
			--system \
			--no-create-home \
			--home-dir /nonexistent \
			--shell /usr/sbin/nologin \
			--gid "$LITE_NAS_SYSTEM_METRICS_RUNTIME_GROUP" \
			--groups "$LITE_NAS_SYSTEM_METRICS_CONFIG_GROUP" \
			"$LITE_NAS_SYSTEM_METRICS_RUNTIME_USER"
		return 0
	fi

	log.info "Updating system user groups: $LITE_NAS_SYSTEM_METRICS_RUNTIME_USER"
	usermod \
		--gid "$LITE_NAS_SYSTEM_METRICS_RUNTIME_GROUP" \
		--append \
		--groups "$LITE_NAS_SYSTEM_METRICS_CONFIG_GROUP" \
		"$LITE_NAS_SYSTEM_METRICS_RUNTIME_USER"
}

deploy.systemMetrics.installBinary() {
	local source_binary="$1"
	local target_dir

	if [ ! -f "$source_binary" ]; then
		log.error "Missing system-metrics binary: $source_binary"
		exit 1
	fi

	target_dir="$(dirname "$LITE_NAS_SYSTEM_METRICS_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"

	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_SYSTEM_METRICS_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_SYSTEM_METRICS_BINARY_TARGET"
		return 0
	fi

	install -m 0755 "$source_binary" "$LITE_NAS_SYSTEM_METRICS_BINARY_TARGET"
}

deploy.systemMetrics.installConfig() {
	if [ ! -f "$LITE_NAS_SYSTEM_METRICS_CONFIG_SOURCE" ]; then
		log.error "Missing system-metrics config source: $LITE_NAS_SYSTEM_METRICS_CONFIG_SOURCE"
		exit 1
	fi

	install -d -m 0750 -o root -g "$LITE_NAS_SYSTEM_METRICS_CONFIG_GROUP" "$LITE_NAS_SYSTEM_METRICS_CONFIG_DIR"
	install -m 0640 -o root -g "$LITE_NAS_SYSTEM_METRICS_CONFIG_GROUP" \
		"$LITE_NAS_SYSTEM_METRICS_CONFIG_SOURCE" \
		"$LITE_NAS_SYSTEM_METRICS_CONFIG_TARGET"
}

deploy.systemMetrics.installLogTarget() {
	install -d -m 0750 -o root -g "$LITE_NAS_SYSTEM_METRICS_RUNTIME_GROUP" "$LITE_NAS_SYSTEM_METRICS_LOG_DIR"

	if [ ! -f "$LITE_NAS_SYSTEM_METRICS_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_SYSTEM_METRICS_RUNTIME_USER" -g "$LITE_NAS_SYSTEM_METRICS_RUNTIME_GROUP" \
			/dev/null \
			"$LITE_NAS_SYSTEM_METRICS_LOG_FILE"
		return 0
	fi

	chown "$LITE_NAS_SYSTEM_METRICS_RUNTIME_USER:$LITE_NAS_SYSTEM_METRICS_RUNTIME_GROUP" "$LITE_NAS_SYSTEM_METRICS_LOG_FILE"
	chmod 0640 "$LITE_NAS_SYSTEM_METRICS_LOG_FILE"
}

deploy.systemMetrics.escapeSedReplacement() {
	printf '%s' "$1" | sed -e 's/[&|]/\\&/g'
}

deploy.systemMetrics.installUnitFile() {
	local binary_target
	local config_dir
	local config_group
	local log_file
	local runtime_group
	local runtime_user
	local rendered_unit

	if [ ! -f "$LITE_NAS_SYSTEM_METRICS_UNIT_TEMPLATE" ]; then
		log.error "Missing systemd unit template: $LITE_NAS_SYSTEM_METRICS_UNIT_TEMPLATE"
		exit 1
	fi

	binary_target="$(deploy.systemMetrics.escapeSedReplacement "$LITE_NAS_SYSTEM_METRICS_BINARY_TARGET")"
	config_dir="$(deploy.systemMetrics.escapeSedReplacement "$LITE_NAS_SYSTEM_METRICS_CONFIG_DIR")"
	config_group="$(deploy.systemMetrics.escapeSedReplacement "$LITE_NAS_SYSTEM_METRICS_CONFIG_GROUP")"
	log_file="$(deploy.systemMetrics.escapeSedReplacement "$LITE_NAS_SYSTEM_METRICS_LOG_FILE")"
	runtime_group="$(deploy.systemMetrics.escapeSedReplacement "$LITE_NAS_SYSTEM_METRICS_RUNTIME_GROUP")"
	runtime_user="$(deploy.systemMetrics.escapeSedReplacement "$LITE_NAS_SYSTEM_METRICS_RUNTIME_USER")"

	rendered_unit="$(mktemp)"
	sed \
		-e "s|@SYSTEM_METRICS_BINARY@|$binary_target|g" \
		-e "s|@SYSTEM_METRICS_CONFIG_DIR@|$config_dir|g" \
		-e "s|@SYSTEM_METRICS_CONFIG_GROUP@|$config_group|g" \
		-e "s|@SYSTEM_METRICS_LOG_FILE@|$log_file|g" \
		-e "s|@SYSTEM_METRICS_RUNTIME_GROUP@|$runtime_group|g" \
		-e "s|@SYSTEM_METRICS_RUNTIME_USER@|$runtime_user|g" \
		"$LITE_NAS_SYSTEM_METRICS_UNIT_TEMPLATE" >"$rendered_unit"

	install -D -m 0644 "$rendered_unit" "$LITE_NAS_SYSTEM_METRICS_UNIT_TARGET"
	rm -f "$rendered_unit"
}

deploy.systemMetrics.enableAndStart() {
	systemctl daemon-reload
	systemctl enable --now "$LITE_NAS_SYSTEM_METRICS_SERVICE_NAME.service"
}

deploy.systemMetrics.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.systemMetrics.ensureRuntimeUser
	deploy.systemMetrics.installBinary "$source_binary"
	deploy.systemMetrics.installConfig
	deploy.systemMetrics.installLogTarget
	deploy.systemMetrics.installUnitFile

	if [ "$should_start" = "1" ]; then
		deploy.systemMetrics.enableAndStart
		return 0
	fi

	log.info "Skipping system-metrics enable/start."
}
