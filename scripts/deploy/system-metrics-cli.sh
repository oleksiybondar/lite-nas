#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_SYSTEM_METRICS_CLI_ACCESS_GROUP="${LITE_NAS_SYSTEM_METRICS_CLI_ACCESS_GROUP:-users}"
readonly LITE_NAS_SYSTEM_METRICS_CLI_BINARY_TARGET="${LITE_NAS_SYSTEM_METRICS_CLI_BINARY_TARGET:-/usr/libexec/lite-nas/system-metrics-cli}"
readonly LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_SOURCE="${LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/system-metrics-cli.conf}"
readonly LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_TARGET="${LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_TARGET:-$LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_DIR/system-metrics-cli.conf}"

deploy.systemMetricsCLI.usage() {
	cat <<'MSG'
Usage: scripts/deploy-system-metrics-cli.sh [options]

Options:
  --binary PATH       Install an existing binary instead of building one.
  --skip-bootstrap    Install files without running LiteNAS bootstrap first.
  -h, --help          Show this help.
MSG
}

deploy.systemMetricsCLI.requireTools() {
	local tool
	local tools=(
		getent
		groupadd
		install
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.systemMetricsCLI.ensureConfigGroup() {
	if getent group "$LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_GROUP" >/dev/null 2>&1; then
		return
	fi

	log.info "Creating system group: $LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_GROUP"
	groupadd --system "$LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_GROUP"
}

deploy.systemMetricsCLI.ensureAccessGroup() {
	if getent group "$LITE_NAS_SYSTEM_METRICS_CLI_ACCESS_GROUP" >/dev/null 2>&1; then
		return
	fi

	log.info "Creating system group: $LITE_NAS_SYSTEM_METRICS_CLI_ACCESS_GROUP"
	groupadd --system "$LITE_NAS_SYSTEM_METRICS_CLI_ACCESS_GROUP"
}

deploy.systemMetricsCLI.installBinary() {
	local source_binary="$1"
	local target_dir

	if [ ! -f "$source_binary" ]; then
		log.error "Missing system-metrics-cli binary: $source_binary"
		exit 1
	fi

	target_dir="$(dirname "$LITE_NAS_SYSTEM_METRICS_CLI_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"

	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_SYSTEM_METRICS_CLI_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_SYSTEM_METRICS_CLI_BINARY_TARGET"
		return 0
	fi

	install -m 0755 "$source_binary" "$LITE_NAS_SYSTEM_METRICS_CLI_BINARY_TARGET"
}

deploy.systemMetricsCLI.installConfig() {
	if [ ! -f "$LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_SOURCE" ]; then
		log.error "Missing system-metrics-cli config source: $LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_SOURCE"
		exit 1
	fi

	install -d -m 0711 -o root -g "$LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_GROUP" "$LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_DIR"
	install -m 0640 -o root -g "$LITE_NAS_SYSTEM_METRICS_CLI_ACCESS_GROUP" \
		"$LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_SOURCE" \
		"$LITE_NAS_SYSTEM_METRICS_CLI_CONFIG_TARGET"
}

deploy.systemMetricsCLI.deploy() {
	local source_binary="$1"

	deploy.systemMetricsCLI.ensureConfigGroup
	deploy.systemMetricsCLI.ensureAccessGroup
	deploy.systemMetricsCLI.installBinary "$source_binary"
	deploy.systemMetricsCLI.installConfig
}
