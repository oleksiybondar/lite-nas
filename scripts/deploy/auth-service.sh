#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_AUTH_SERVICE_NAME="${LITE_NAS_AUTH_SERVICE_NAME:-lite-nas-auth}"
readonly LITE_NAS_AUTH_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_AUTH_BINARY_TARGET="${LITE_NAS_AUTH_BINARY_TARGET:-/usr/libexec/lite-nas/auth-service}"
readonly LITE_NAS_AUTH_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_AUTH_CONFIG_SOURCE="${LITE_NAS_AUTH_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/auth.conf}"
readonly LITE_NAS_AUTH_CONFIG_TARGET="${LITE_NAS_AUTH_CONFIG_TARGET:-$LITE_NAS_AUTH_CONFIG_DIR/auth.conf}"
readonly LITE_NAS_AUTH_UNIT_TEMPLATE="${LITE_NAS_AUTH_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-auth.service}"
readonly LITE_NAS_AUTH_UNIT_TARGET="${LITE_NAS_AUTH_UNIT_TARGET:-/etc/systemd/system/lite-nas-auth.service}"
readonly LITE_NAS_AUTH_LOG_DIR="${LITE_NAS_AUTH_LOG_DIR:-/var/lib/lite-nas}"
readonly LITE_NAS_AUTH_LOG_FILE="${LITE_NAS_AUTH_LOG_FILE:-$LITE_NAS_AUTH_LOG_DIR/auth-service.log}"

deploy.authService.usage() {
	cat <<'MSG'
Usage: scripts/deploy-auth-service.sh [options]

Options:
  --binary PATH       Install an existing binary instead of building one.
  --no-start          Install files but do not enable or start the service.
  --skip-bootstrap    Install files without running LiteNAS bootstrap first.
  -h, --help          Show this help.
MSG
}

deploy.authService.requireTools() {
	local tool
	local tools=(
		getent
		groupadd
		install
		chmod
		chown
		realpath
		systemctl
		touch
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.authService.ensureConfigGroup() {
	if getent group "$LITE_NAS_AUTH_CONFIG_GROUP" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating system group: $LITE_NAS_AUTH_CONFIG_GROUP"
	groupadd --system "$LITE_NAS_AUTH_CONFIG_GROUP"
}

deploy.authService.installBinary() {
	local source_binary="$1"
	local target_dir

	if [ ! -f "$source_binary" ]; then
		log.error "Missing auth-service binary: $source_binary"
		exit 1
	fi

	target_dir="$(dirname "$LITE_NAS_AUTH_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"

	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_AUTH_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_AUTH_BINARY_TARGET"
		return 0
	fi

	install -m 0755 "$source_binary" "$LITE_NAS_AUTH_BINARY_TARGET"
}

deploy.authService.installConfig() {
	if [ ! -f "$LITE_NAS_AUTH_CONFIG_SOURCE" ]; then
		log.error "Missing auth-service config source: $LITE_NAS_AUTH_CONFIG_SOURCE"
		exit 1
	fi

	install -d -m 0750 -o root -g "$LITE_NAS_AUTH_CONFIG_GROUP" "$LITE_NAS_AUTH_CONFIG_DIR"
	install -m 0640 -o root -g "$LITE_NAS_AUTH_CONFIG_GROUP" \
		"$LITE_NAS_AUTH_CONFIG_SOURCE" \
		"$LITE_NAS_AUTH_CONFIG_TARGET"
}

deploy.authService.installLogTarget() {
	install -d -m 0750 -o root -g "$LITE_NAS_AUTH_CONFIG_GROUP" "$LITE_NAS_AUTH_LOG_DIR"

	if [ ! -f "$LITE_NAS_AUTH_LOG_FILE" ]; then
		install -m 0640 -o root -g root /dev/null "$LITE_NAS_AUTH_LOG_FILE"
		return 0
	fi

	chown root:root "$LITE_NAS_AUTH_LOG_FILE"
	chmod 0640 "$LITE_NAS_AUTH_LOG_FILE"
}

deploy.authService.installUnitFile() {
	if [ ! -f "$LITE_NAS_AUTH_UNIT_TEMPLATE" ]; then
		log.error "Missing systemd unit template: $LITE_NAS_AUTH_UNIT_TEMPLATE"
		exit 1
	fi

	install -D -m 0644 "$LITE_NAS_AUTH_UNIT_TEMPLATE" "$LITE_NAS_AUTH_UNIT_TARGET"
}

deploy.authService.enableAndStart() {
	systemctl daemon-reload
	systemctl enable --now "$LITE_NAS_AUTH_SERVICE_NAME.service"
}

deploy.authService.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.authService.ensureConfigGroup
	deploy.authService.installBinary "$source_binary"
	deploy.authService.installConfig
	deploy.authService.installLogTarget
	deploy.authService.installUnitFile

	if [ "$should_start" = "1" ]; then
		deploy.authService.enableAndStart
		return 0
	fi

	log.info "Skipping auth-service enable/start."
}
