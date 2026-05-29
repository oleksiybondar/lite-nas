#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_POSTFIX_SERVICE_NAME="${LITE_NAS_POSTFIX_SERVICE_NAME:-postfix}"
readonly LITE_NAS_POSTFIX_CONFIG_SOURCE_DIR="${LITE_NAS_POSTFIX_CONFIG_SOURCE_DIR:-$LITE_NAS_REPO_ROOT/configs/etc/postfix}"
readonly LITE_NAS_POSTFIX_CONFIG_TARGET_DIR="${LITE_NAS_POSTFIX_CONFIG_TARGET_DIR:-/etc/postfix}"
readonly LITE_NAS_POSTFIX_MAIN_CF_SOURCE="${LITE_NAS_POSTFIX_MAIN_CF_SOURCE:-$LITE_NAS_POSTFIX_CONFIG_SOURCE_DIR/main.cf}"
readonly LITE_NAS_POSTFIX_MASTER_CF_SOURCE="${LITE_NAS_POSTFIX_MASTER_CF_SOURCE:-$LITE_NAS_POSTFIX_CONFIG_SOURCE_DIR/master.cf}"
readonly LITE_NAS_POSTFIX_MAIN_CF_TARGET="${LITE_NAS_POSTFIX_MAIN_CF_TARGET:-$LITE_NAS_POSTFIX_CONFIG_TARGET_DIR/main.cf}"
readonly LITE_NAS_POSTFIX_MASTER_CF_TARGET="${LITE_NAS_POSTFIX_MASTER_CF_TARGET:-$LITE_NAS_POSTFIX_CONFIG_TARGET_DIR/master.cf}"

deploy.postfix.requireTools() {
	local tool
	local tools=(install mkdir postfix postconf systemctl)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required Postfix/AppArmor tooling and retry."
	done
}

deploy.postfix.installConfig() {
	if [ ! -f "$LITE_NAS_POSTFIX_MAIN_CF_SOURCE" ]; then
		log.error "Missing Postfix main.cf source: $LITE_NAS_POSTFIX_MAIN_CF_SOURCE"
		exit 1
	fi
	if [ ! -f "$LITE_NAS_POSTFIX_MASTER_CF_SOURCE" ]; then
		log.error "Missing Postfix master.cf source: $LITE_NAS_POSTFIX_MASTER_CF_SOURCE"
		exit 1
	fi

	install -d -m 0755 "$LITE_NAS_POSTFIX_CONFIG_TARGET_DIR"
	install -m 0644 "$LITE_NAS_POSTFIX_MAIN_CF_SOURCE" "$LITE_NAS_POSTFIX_MAIN_CF_TARGET"
	install -m 0644 "$LITE_NAS_POSTFIX_MASTER_CF_SOURCE" "$LITE_NAS_POSTFIX_MASTER_CF_TARGET"
}

deploy.postfix.validateConfig() {
	postfix check
	postconf inet_interfaces >/dev/null
}

deploy.postfix.enableAndStart() {
	systemctl daemon-reload
	systemctl enable "$LITE_NAS_POSTFIX_SERVICE_NAME.service"
	systemctl restart "$LITE_NAS_POSTFIX_SERVICE_NAME.service"
}

deploy.postfix.deploy() {
	local should_start="${1:-1}"
	deploy.postfix.installConfig
	deploy.postfix.validateConfig
	if [ "$should_start" = "1" ]; then
		deploy.postfix.enableAndStart
		return 0
	fi

	log.info "Skipping Postfix enable/start."
}
