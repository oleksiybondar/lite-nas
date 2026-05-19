#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_ZFS_METRICS_CLI_USER="${LITE_NAS_ZFS_METRICS_CLI_USER:-lite-nas-zfs-metrics-cli}"
readonly LITE_NAS_ZFS_METRICS_CLI_GROUP="${LITE_NAS_ZFS_METRICS_CLI_GROUP:-$LITE_NAS_ZFS_METRICS_CLI_USER}"
readonly LITE_NAS_ZFS_METRICS_CLI_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_ZFS_METRICS_CLI_BINARY_TARGET="${LITE_NAS_ZFS_METRICS_CLI_BINARY_TARGET:-/usr/libexec/lite-nas/zfs-metrics-cli}"
readonly LITE_NAS_ZFS_METRICS_CLI_SYMLINK_TARGET="${LITE_NAS_ZFS_METRICS_CLI_SYMLINK_TARGET:-/usr/bin/zfs-metrics-cli}"
readonly LITE_NAS_ZFS_METRICS_CLI_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_ZFS_METRICS_CLI_CONFIG_SOURCE="${LITE_NAS_ZFS_METRICS_CLI_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/zfs-metrics-cli.conf}"
readonly LITE_NAS_ZFS_METRICS_CLI_CONFIG_TARGET="${LITE_NAS_ZFS_METRICS_CLI_CONFIG_TARGET:-$LITE_NAS_ZFS_METRICS_CLI_CONFIG_DIR/zfs-metrics-cli.conf}"
readonly LITE_NAS_ZFS_METRICS_CLI_LOG_DIR="${LITE_NAS_ZFS_METRICS_CLI_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_ZFS_METRICS_CLI_LOG_FILE="${LITE_NAS_ZFS_METRICS_CLI_LOG_FILE:-$LITE_NAS_ZFS_METRICS_CLI_LOG_DIR/zfs-metrics-cli.log}"
readonly LITE_NAS_ZFS_METRICS_CLI_CERT_DIR="${LITE_NAS_ZFS_METRICS_CLI_CERT_DIR:-/etc/lite-nas/certificates/transport/lite-nas-zfs-metrics-cli}"

deploy.zfsMetricsCLI.requireTools() {
	local tool
	local tools=(getent groupadd install ln chmod chown id useradd usermod)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.zfsMetricsCLI.ensureGroup() {
	local group_name="$1"
	if getent group "$group_name" >/dev/null 2>&1; then
		return
	fi
	groupadd --system "$group_name"
}

deploy.zfsMetricsCLI.ensureUser() {
	deploy.zfsMetricsCLI.ensureGroup "$LITE_NAS_ZFS_METRICS_CLI_CONFIG_GROUP"
	deploy.zfsMetricsCLI.ensureGroup "$LITE_NAS_ZFS_METRICS_CLI_GROUP"
	if ! id "$LITE_NAS_ZFS_METRICS_CLI_USER" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin \
			--gid "$LITE_NAS_ZFS_METRICS_CLI_GROUP" \
			--groups "$LITE_NAS_ZFS_METRICS_CLI_CONFIG_GROUP" \
			"$LITE_NAS_ZFS_METRICS_CLI_USER"
		return
	fi
	usermod --gid "$LITE_NAS_ZFS_METRICS_CLI_GROUP" --append --groups "$LITE_NAS_ZFS_METRICS_CLI_CONFIG_GROUP" "$LITE_NAS_ZFS_METRICS_CLI_USER"
}

deploy.zfsMetricsCLI.ensureMessagingCertificates() {
	if [ -f "$LITE_NAS_ZFS_METRICS_CLI_CERT_DIR/client.crt" ] && [ -f "$LITE_NAS_ZFS_METRICS_CLI_CERT_DIR/client.key" ]; then
		return
	fi
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing --user "$LITE_NAS_ZFS_METRICS_CLI_USER"
}

deploy.zfsMetricsCLI.installBinary() {
	local source_binary="$1"
	local target_dir
	if [ ! -f "$source_binary" ]; then
		log.error "Missing zfs-metrics-cli binary: $source_binary"
		exit 1
	fi
	target_dir="$(dirname "$LITE_NAS_ZFS_METRICS_CLI_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_ZFS_METRICS_CLI_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_ZFS_METRICS_CLI_BINARY_TARGET"
		return 0
	fi
	install -m 0755 "$source_binary" "$LITE_NAS_ZFS_METRICS_CLI_BINARY_TARGET"
}

deploy.zfsMetricsCLI.installBinarySymlink() {
	local symlink_dir
	symlink_dir="$(dirname "$LITE_NAS_ZFS_METRICS_CLI_SYMLINK_TARGET")"
	install -d -m 0755 "$symlink_dir"
	ln -sfn "$LITE_NAS_ZFS_METRICS_CLI_BINARY_TARGET" "$LITE_NAS_ZFS_METRICS_CLI_SYMLINK_TARGET"
}

deploy.zfsMetricsCLI.installConfig() {
	if [ ! -f "$LITE_NAS_ZFS_METRICS_CLI_CONFIG_SOURCE" ]; then
		log.error "Missing zfs-metrics-cli config source: $LITE_NAS_ZFS_METRICS_CLI_CONFIG_SOURCE"
		exit 1
	fi
	install -d -m 0711 -o root -g "$LITE_NAS_ZFS_METRICS_CLI_CONFIG_GROUP" "$LITE_NAS_ZFS_METRICS_CLI_CONFIG_DIR"
	install -m 0644 -o root -g root "$LITE_NAS_ZFS_METRICS_CLI_CONFIG_SOURCE" "$LITE_NAS_ZFS_METRICS_CLI_CONFIG_TARGET"
}

deploy.zfsMetricsCLI.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_ZFS_METRICS_CLI_CONFIG_GROUP" "$LITE_NAS_ZFS_METRICS_CLI_LOG_DIR"
	if [ ! -f "$LITE_NAS_ZFS_METRICS_CLI_LOG_FILE" ]; then
		install -m 0666 -o root -g root /dev/null "$LITE_NAS_ZFS_METRICS_CLI_LOG_FILE"
	fi
	chown root:root "$LITE_NAS_ZFS_METRICS_CLI_LOG_FILE"
	chmod 0666 "$LITE_NAS_ZFS_METRICS_CLI_LOG_FILE"
}

deploy.zfsMetricsCLI.deploy() {
	local source_binary="$1"
	deploy.zfsMetricsCLI.ensureUser
	deploy.zfsMetricsCLI.ensureMessagingCertificates
	deploy.zfsMetricsCLI.installBinary "$source_binary"
	deploy.zfsMetricsCLI.installBinarySymlink
	deploy.zfsMetricsCLI.installConfig
	deploy.zfsMetricsCLI.installLogTarget
}
