#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_ZFS_METRICS_SERVICE_NAME="${LITE_NAS_ZFS_METRICS_SERVICE_NAME:-lite-nas-zfs-metrics}"
readonly LITE_NAS_ZFS_METRICS_RUNTIME_USER="${LITE_NAS_ZFS_METRICS_RUNTIME_USER:-lite-nas-zfs-metrics}"
readonly LITE_NAS_ZFS_METRICS_RUNTIME_GROUP="${LITE_NAS_ZFS_METRICS_RUNTIME_GROUP:-$LITE_NAS_ZFS_METRICS_RUNTIME_USER}"
readonly LITE_NAS_ZFS_METRICS_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_ZFS_METRICS_OPERATOR_GROUP="${LITE_NAS_OPERATOR_GROUP:-lite-nas-operator}"
readonly LITE_NAS_ZFS_METRICS_BINARY_TARGET="${LITE_NAS_ZFS_METRICS_BINARY_TARGET:-/usr/libexec/lite-nas/zfs-metrics}"
readonly LITE_NAS_ZFS_METRICS_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_ZFS_METRICS_CONFIG_SOURCE="${LITE_NAS_ZFS_METRICS_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/zfs-metrics.conf}"
readonly LITE_NAS_ZFS_METRICS_CONFIG_TARGET="${LITE_NAS_ZFS_METRICS_CONFIG_TARGET:-$LITE_NAS_ZFS_METRICS_CONFIG_DIR/zfs-metrics.conf}"
readonly LITE_NAS_ZFS_METRICS_UNIT_TEMPLATE="${LITE_NAS_ZFS_METRICS_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-zfs-metrics.service}"
readonly LITE_NAS_ZFS_METRICS_UNIT_TARGET="${LITE_NAS_ZFS_METRICS_UNIT_TARGET:-/etc/systemd/system/lite-nas-zfs-metrics.service}"
readonly LITE_NAS_ZFS_METRICS_LOG_DIR="${LITE_NAS_ZFS_METRICS_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_ZFS_METRICS_LOG_FILE="${LITE_NAS_ZFS_METRICS_LOG_FILE:-$LITE_NAS_ZFS_METRICS_LOG_DIR/zfs-metrics.log}"
readonly LITE_NAS_ZFS_METRICS_CERT_DIR="${LITE_NAS_ZFS_METRICS_CERT_DIR:-/etc/lite-nas/certificates/transport/lite-nas-zfs-metrics}"

deploy.zfsMetrics.requireTools() {
	local tool
	local tools=(getent groupadd install chmod chown id useradd usermod systemctl realpath)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
	log.requireCommand "bash" "Install bash and retry."
}

deploy.zfsMetrics.ensureGroup() {
	local group_name="$1"
	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi
	groupadd --system "$group_name"
}

deploy.zfsMetrics.ensureRuntimeUser() {
	deploy.zfsMetrics.ensureGroup "$LITE_NAS_ZFS_METRICS_CONFIG_GROUP"
	deploy.zfsMetrics.ensureGroup "$LITE_NAS_ZFS_METRICS_RUNTIME_GROUP"
	deploy.zfsMetrics.ensureGroup "$LITE_NAS_ZFS_METRICS_OPERATOR_GROUP"
	if ! id "$LITE_NAS_ZFS_METRICS_RUNTIME_USER" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin \
			--gid "$LITE_NAS_ZFS_METRICS_RUNTIME_GROUP" \
			--groups "$LITE_NAS_ZFS_METRICS_CONFIG_GROUP,$LITE_NAS_ZFS_METRICS_OPERATOR_GROUP" \
			"$LITE_NAS_ZFS_METRICS_RUNTIME_USER"
		return 0
	fi

	usermod \
		--gid "$LITE_NAS_ZFS_METRICS_RUNTIME_GROUP" \
		--append \
		--groups "$LITE_NAS_ZFS_METRICS_CONFIG_GROUP,$LITE_NAS_ZFS_METRICS_OPERATOR_GROUP" \
		"$LITE_NAS_ZFS_METRICS_RUNTIME_USER"
}

deploy.zfsMetrics.ensureMessagingCertificates() {
	if [ -f "$LITE_NAS_ZFS_METRICS_CERT_DIR/client.crt" ] && [ -f "$LITE_NAS_ZFS_METRICS_CERT_DIR/client.key" ]; then
		return 0
	fi

	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing --user "$LITE_NAS_ZFS_METRICS_RUNTIME_USER"
}

deploy.zfsMetrics.installBinary() {
	local source_binary="$1"
	local target_dir
	if [ ! -f "$source_binary" ]; then
		log.error "Missing zfs-metrics binary: $source_binary"
		exit 1
	fi
	target_dir="$(dirname "$LITE_NAS_ZFS_METRICS_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_ZFS_METRICS_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_ZFS_METRICS_BINARY_TARGET"
		return 0
	fi
	install -m 0755 "$source_binary" "$LITE_NAS_ZFS_METRICS_BINARY_TARGET"
}

deploy.zfsMetrics.installConfig() {
	if [ ! -f "$LITE_NAS_ZFS_METRICS_CONFIG_SOURCE" ]; then
		log.error "Missing zfs-metrics config source: $LITE_NAS_ZFS_METRICS_CONFIG_SOURCE"
		exit 1
	fi

	install -d -m 0711 -o root -g "$LITE_NAS_ZFS_METRICS_CONFIG_GROUP" "$LITE_NAS_ZFS_METRICS_CONFIG_DIR"
	install -m 0640 -o "$LITE_NAS_ZFS_METRICS_RUNTIME_USER" -g "$LITE_NAS_ZFS_METRICS_CONFIG_GROUP" \
		"$LITE_NAS_ZFS_METRICS_CONFIG_SOURCE" \
		"$LITE_NAS_ZFS_METRICS_CONFIG_TARGET"
}

deploy.zfsMetrics.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_ZFS_METRICS_CONFIG_GROUP" "$LITE_NAS_ZFS_METRICS_LOG_DIR"
	if [ ! -f "$LITE_NAS_ZFS_METRICS_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_ZFS_METRICS_RUNTIME_USER" -g "$LITE_NAS_ZFS_METRICS_RUNTIME_GROUP" \
			/dev/null "$LITE_NAS_ZFS_METRICS_LOG_FILE"
		return 0
	fi
	chown "$LITE_NAS_ZFS_METRICS_RUNTIME_USER:$LITE_NAS_ZFS_METRICS_RUNTIME_GROUP" "$LITE_NAS_ZFS_METRICS_LOG_FILE"
	chmod 0640 "$LITE_NAS_ZFS_METRICS_LOG_FILE"
}

deploy.zfsMetrics.installUnitFile() {
	if [ ! -f "$LITE_NAS_ZFS_METRICS_UNIT_TEMPLATE" ]; then
		log.error "Missing zfs-metrics systemd unit template: $LITE_NAS_ZFS_METRICS_UNIT_TEMPLATE"
		exit 1
	fi
	install -D -m 0644 "$LITE_NAS_ZFS_METRICS_UNIT_TEMPLATE" "$LITE_NAS_ZFS_METRICS_UNIT_TARGET"
}

deploy.zfsMetrics.enableAndStart() {
	systemctl daemon-reload
	systemctl enable "$LITE_NAS_ZFS_METRICS_SERVICE_NAME.service"
	systemctl restart "$LITE_NAS_ZFS_METRICS_SERVICE_NAME.service"
}

deploy.zfsMetrics.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.zfsMetrics.ensureRuntimeUser
	deploy.zfsMetrics.ensureMessagingCertificates
	deploy.zfsMetrics.installBinary "$source_binary"
	deploy.zfsMetrics.installConfig
	deploy.zfsMetrics.installLogTarget
	deploy.zfsMetrics.installUnitFile

	if [ "$should_start" = "1" ]; then
		deploy.zfsMetrics.enableAndStart
		return 0
	fi

	log.info "Skipping zfs-metrics enable/start."
}
