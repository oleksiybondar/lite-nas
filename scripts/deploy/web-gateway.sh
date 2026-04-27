#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/nginx.sh"

readonly LITE_NAS_WEB_GATEWAY_SERVICE_NAME="${LITE_NAS_WEB_GATEWAY_SERVICE_NAME:-lite-nas-web-gateway}"
readonly LITE_NAS_WEB_GATEWAY_RUNTIME_USER="${LITE_NAS_WEB_GATEWAY_RUNTIME_USER:-lite-nas-web-gateway}"
readonly LITE_NAS_WEB_GATEWAY_RUNTIME_GROUP="${LITE_NAS_WEB_GATEWAY_RUNTIME_GROUP:-$LITE_NAS_WEB_GATEWAY_RUNTIME_USER}"
readonly LITE_NAS_WEB_GATEWAY_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_WEB_GATEWAY_BINARY_TARGET="${LITE_NAS_WEB_GATEWAY_BINARY_TARGET:-/usr/libexec/lite-nas/web-gateway}"
readonly LITE_NAS_WEB_GATEWAY_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/liteNAS}"
readonly LITE_NAS_WEB_GATEWAY_CONFIG_SOURCE="${LITE_NAS_WEB_GATEWAY_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/liteNAS/web-gateway.conf}"
readonly LITE_NAS_WEB_GATEWAY_CONFIG_TARGET="${LITE_NAS_WEB_GATEWAY_CONFIG_TARGET:-$LITE_NAS_WEB_GATEWAY_CONFIG_DIR/web-gateway.conf}"
readonly LITE_NAS_WEB_GATEWAY_UNIT_TEMPLATE="${LITE_NAS_WEB_GATEWAY_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/systemd/system/lite-nas-web-gateway.service}"
readonly LITE_NAS_WEB_GATEWAY_UNIT_TARGET="${LITE_NAS_WEB_GATEWAY_UNIT_TARGET:-/lib/systemd/system/lite-nas-web-gateway.service}"
readonly LITE_NAS_WEB_GATEWAY_LOG_DIR="${LITE_NAS_WEB_GATEWAY_LOG_DIR:-/var/log/liteNAS}"
readonly LITE_NAS_WEB_GATEWAY_LOG_FILE="${LITE_NAS_WEB_GATEWAY_LOG_FILE:-$LITE_NAS_WEB_GATEWAY_LOG_DIR/web-gateway.log}"
readonly LITE_NAS_WEB_GATEWAY_SHARE_ROOT="${LITE_NAS_WEB_GATEWAY_SHARE_ROOT:-/usr/share/lite-nas/web-gateway}"
readonly LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE="${LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE:-$LITE_NAS_REPO_ROOT/services/web-gateway/assets}"

deploy.webGateway.usage() {
	cat <<'MSG'
Usage: scripts/deploy-web-gateway.sh [options]

Options:
  --binary PATH       Install an existing binary instead of building one.
  --arch=amd64|arm64  Build target architecture when --binary is not set.
  --no-start          Install files but do not enable or start the service.
  --skip-bootstrap    Install files without running LiteNAS bootstrap first.
  -h, --help          Show this help.
MSG
}

deploy.webGateway.requireTools() {
	local tool
	local tools=(
		cp
		getent
		groupadd
		install
		chmod
		chown
		realpath
		rm
		sed
		systemctl
		touch
		useradd
		usermod
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done

	deploy.nginx.requireTools
}

deploy.webGateway.ensureGroup() {
	local group_name="$1"

	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating system group: $group_name"
	groupadd --system "$group_name"
}

deploy.webGateway.ensureRuntimeUser() {
	deploy.webGateway.ensureGroup "$LITE_NAS_WEB_GATEWAY_CONFIG_GROUP"
	deploy.webGateway.ensureGroup "$LITE_NAS_WEB_GATEWAY_RUNTIME_GROUP"

	if ! id "$LITE_NAS_WEB_GATEWAY_RUNTIME_USER" >/dev/null 2>&1; then
		log.info "Creating system user: $LITE_NAS_WEB_GATEWAY_RUNTIME_USER"
		useradd \
			--system \
			--no-create-home \
			--home-dir /nonexistent \
			--shell /usr/sbin/nologin \
			--gid "$LITE_NAS_WEB_GATEWAY_RUNTIME_GROUP" \
			--groups "$LITE_NAS_WEB_GATEWAY_CONFIG_GROUP" \
			"$LITE_NAS_WEB_GATEWAY_RUNTIME_USER"
		return 0
	fi

	log.info "Updating system user groups: $LITE_NAS_WEB_GATEWAY_RUNTIME_USER"
	usermod \
		--gid "$LITE_NAS_WEB_GATEWAY_RUNTIME_GROUP" \
		--append \
		--groups "$LITE_NAS_WEB_GATEWAY_CONFIG_GROUP" \
		"$LITE_NAS_WEB_GATEWAY_RUNTIME_USER"
}

deploy.webGateway.installBinary() {
	local source_binary="$1"
	local target_dir

	if [ ! -f "$source_binary" ]; then
		log.error "Missing web-gateway binary: $source_binary"
		exit 1
	fi

	target_dir="$(dirname "$LITE_NAS_WEB_GATEWAY_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"

	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_WEB_GATEWAY_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_WEB_GATEWAY_BINARY_TARGET"
		return 0
	fi

	install -m 0755 "$source_binary" "$LITE_NAS_WEB_GATEWAY_BINARY_TARGET"
}

deploy.webGateway.installConfig() {
	if [ ! -f "$LITE_NAS_WEB_GATEWAY_CONFIG_SOURCE" ]; then
		log.error "Missing web-gateway config source: $LITE_NAS_WEB_GATEWAY_CONFIG_SOURCE"
		exit 1
	fi

	install -d -m 0750 -o root -g "$LITE_NAS_WEB_GATEWAY_CONFIG_GROUP" "$LITE_NAS_WEB_GATEWAY_CONFIG_DIR"
	install -m 0640 -o root -g "$LITE_NAS_WEB_GATEWAY_CONFIG_GROUP" \
		"$LITE_NAS_WEB_GATEWAY_CONFIG_SOURCE" \
		"$LITE_NAS_WEB_GATEWAY_CONFIG_TARGET"
}

deploy.webGateway.installSharedAssets() {
	if [ ! -d "$LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE" ]; then
		log.error "Missing web-gateway assets source: $LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE"
		exit 1
	fi

	deploy.webGateway.validateAssetsSource

	install -d -m 0755 "$LITE_NAS_WEB_GATEWAY_SHARE_ROOT"
	rm -rf "$LITE_NAS_WEB_GATEWAY_SHARE_ROOT/assets"
	cp -a "$LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE" "$LITE_NAS_WEB_GATEWAY_SHARE_ROOT/assets"
	chown -R root:root "$LITE_NAS_WEB_GATEWAY_SHARE_ROOT"
	find "$LITE_NAS_WEB_GATEWAY_SHARE_ROOT" -type d -exec chmod 0755 {} +
	find "$LITE_NAS_WEB_GATEWAY_SHARE_ROOT" -type f -exec chmod 0644 {} +
}

deploy.webGateway.validateAssetsSource() {
	local required_file
	local required_files=(
		index.html
		index.css
		index.js
	)

	for required_file in "${required_files[@]}"; do
		if [ -f "$LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE/$required_file" ]; then
			continue
		fi

		log.error "Missing web-gateway asset: $LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE/$required_file"
		exit 1
	done

	if [ ! -f "$LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE/favicon.ico" ]; then
		log.warn "favicon.ico is not present in $LITE_NAS_WEB_GATEWAY_ASSETS_SOURCE; /favicon.ico will return 404 until it is added."
	fi
}

deploy.webGateway.installLogTarget() {
	install -d -m 0750 -o root -g "$LITE_NAS_WEB_GATEWAY_RUNTIME_GROUP" "$LITE_NAS_WEB_GATEWAY_LOG_DIR"

	if [ ! -f "$LITE_NAS_WEB_GATEWAY_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_WEB_GATEWAY_RUNTIME_USER" -g "$LITE_NAS_WEB_GATEWAY_RUNTIME_GROUP" \
			/dev/null \
			"$LITE_NAS_WEB_GATEWAY_LOG_FILE"
		return 0
	fi

	chown "$LITE_NAS_WEB_GATEWAY_RUNTIME_USER:$LITE_NAS_WEB_GATEWAY_RUNTIME_GROUP" "$LITE_NAS_WEB_GATEWAY_LOG_FILE"
	chmod 0640 "$LITE_NAS_WEB_GATEWAY_LOG_FILE"
}

deploy.webGateway.escapeSedReplacement() {
	printf '%s' "$1" | sed -e 's/[&|]/\\&/g'
}

deploy.webGateway.installUnitFile() {
	local binary_target
	local config_dir
	local config_group
	local log_file
	local runtime_group
	local runtime_user
	local share_dir
	local rendered_unit

	if [ ! -f "$LITE_NAS_WEB_GATEWAY_UNIT_TEMPLATE" ]; then
		log.error "Missing systemd unit template: $LITE_NAS_WEB_GATEWAY_UNIT_TEMPLATE"
		exit 1
	fi

	binary_target="$(deploy.webGateway.escapeSedReplacement "$LITE_NAS_WEB_GATEWAY_BINARY_TARGET")"
	config_dir="$(deploy.webGateway.escapeSedReplacement "$LITE_NAS_WEB_GATEWAY_CONFIG_DIR")"
	config_group="$(deploy.webGateway.escapeSedReplacement "$LITE_NAS_WEB_GATEWAY_CONFIG_GROUP")"
	log_file="$(deploy.webGateway.escapeSedReplacement "$LITE_NAS_WEB_GATEWAY_LOG_FILE")"
	runtime_group="$(deploy.webGateway.escapeSedReplacement "$LITE_NAS_WEB_GATEWAY_RUNTIME_GROUP")"
	runtime_user="$(deploy.webGateway.escapeSedReplacement "$LITE_NAS_WEB_GATEWAY_RUNTIME_USER")"
	share_dir="$(deploy.webGateway.escapeSedReplacement "$LITE_NAS_WEB_GATEWAY_SHARE_ROOT")"

	rendered_unit="$(mktemp)"
	sed \
		-e "s|@WEB_GATEWAY_BINARY@|$binary_target|g" \
		-e "s|@WEB_GATEWAY_CONFIG_DIR@|$config_dir|g" \
		-e "s|@WEB_GATEWAY_CONFIG_GROUP@|$config_group|g" \
		-e "s|@WEB_GATEWAY_LOG_FILE@|$log_file|g" \
		-e "s|@WEB_GATEWAY_RUNTIME_GROUP@|$runtime_group|g" \
		-e "s|@WEB_GATEWAY_RUNTIME_USER@|$runtime_user|g" \
		-e "s|@WEB_GATEWAY_SHARE_DIR@|$share_dir|g" \
		"$LITE_NAS_WEB_GATEWAY_UNIT_TEMPLATE" >"$rendered_unit"

	install -D -m 0644 "$rendered_unit" "$LITE_NAS_WEB_GATEWAY_UNIT_TARGET"
	rm -f "$rendered_unit"
}

deploy.webGateway.enableAndStart() {
	systemctl daemon-reload
	systemctl enable --now "$LITE_NAS_WEB_GATEWAY_SERVICE_NAME.service"
}

deploy.webGateway.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.webGateway.ensureRuntimeUser
	deploy.webGateway.installBinary "$source_binary"
	deploy.webGateway.installConfig
	deploy.webGateway.installSharedAssets
	deploy.webGateway.installLogTarget
	deploy.webGateway.installUnitFile
	deploy.nginx.deploy 0

	if [ "$should_start" = "1" ]; then
		deploy.webGateway.enableAndStart
		deploy.nginx.enableAndStart
		return 0
	fi

	log.info "Skipping web-gateway enable/start."
}
