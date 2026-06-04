#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_SERVICE_NAME="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_SERVICE_NAME:-lite-nas-security-email-notifier}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER:-lite-nas-sec-email-notifier}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_GROUP="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_GROUP:-$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_ACCESS_GROUP="${LITE_NAS_SECURITY_GROUP:-lite-nas-security}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_BINARY_TARGET="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_BINARY_TARGET:-/usr/libexec/lite-nas/security-email-notifier}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_SOURCE="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/security-email-notifier.conf}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_TARGET="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_TARGET:-$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_DIR/security-email-notifier.conf}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_TEMPLATES_SOURCE_DIR="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_TEMPLATES_SOURCE_DIR:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/security-email-notifier}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_TEMPLATES_TARGET_DIR="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_TEMPLATES_TARGET_DIR:-$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_DIR/security-email-notifier}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_UNIT_TEMPLATE="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-security-email-notifier.service}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_UNIT_TARGET="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_UNIT_TARGET:-/etc/systemd/system/lite-nas-security-email-notifier.service}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_DIR="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_FILE="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_FILE:-$LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_DIR/security-email-notifier.log}"

deploy.securityEmailNotifier.requireTools() {
	local tool
	local tools=(find getent groupadd install chmod chown realpath systemctl useradd usermod id)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.securityEmailNotifier.ensureGroup() {
	local group_name="$1"
	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi
	groupadd --system "$group_name"
}

deploy.securityEmailNotifier.ensureRuntimeUser() {
	deploy.securityEmailNotifier.ensureGroup "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_GROUP"
	deploy.securityEmailNotifier.ensureGroup "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_GROUP"
	deploy.securityEmailNotifier.ensureGroup "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_ACCESS_GROUP"
	if ! id "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin \
			--gid "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_GROUP" \
			--groups "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_GROUP,$LITE_NAS_SECURITY_EMAIL_NOTIFIER_ACCESS_GROUP" \
			"$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER"
		return 0
	fi

	usermod \
		--gid "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_GROUP" \
		--append \
		--groups "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_GROUP,$LITE_NAS_SECURITY_EMAIL_NOTIFIER_ACCESS_GROUP" \
		"$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER"
}

deploy.securityEmailNotifier.installBinary() {
	local source_binary="$1"
	local target_dir
	if [ ! -f "$source_binary" ]; then
		log.error "Missing security-email-notifier binary: $source_binary"
		exit 1
	fi
	target_dir="$(dirname "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_BINARY_TARGET"
		return 0
	fi
	install -m 0755 "$source_binary" "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_BINARY_TARGET"
}

deploy.securityEmailNotifier.installConfig() {
	if [ ! -f "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_SOURCE" ]; then
		log.error "Missing security-email-notifier config source: $LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_SOURCE"
		exit 1
	fi
	install -d -m 0711 -o root -g "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_GROUP" "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_DIR"
	install -m 0640 -o "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER" -g "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_GROUP" \
		"$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_SOURCE" \
		"$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_TARGET"
}

deploy.securityEmailNotifier.installTemplates() {
	if [ ! -d "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_TEMPLATES_SOURCE_DIR" ]; then
		log.error "Missing security-email-notifier templates source: $LITE_NAS_SECURITY_EMAIL_NOTIFIER_TEMPLATES_SOURCE_DIR"
		exit 1
	fi

	install -d -m 0750 -o "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER" -g "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_GROUP" \
		"$LITE_NAS_SECURITY_EMAIL_NOTIFIER_TEMPLATES_TARGET_DIR"

	while IFS= read -r -d '' template_file; do
		install -m 0640 -o "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER" -g "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_GROUP" \
			"$template_file" \
			"$LITE_NAS_SECURITY_EMAIL_NOTIFIER_TEMPLATES_TARGET_DIR/$(basename "$template_file")"
	done < <(find "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_TEMPLATES_SOURCE_DIR" -maxdepth 1 -type f -name '*.html' -print0)
}

deploy.securityEmailNotifier.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_CONFIG_GROUP" "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_DIR"
	if [ ! -f "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER" -g "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_GROUP" \
			/dev/null "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_FILE"
		return 0
	fi
	chown "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER:$LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_GROUP" "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_FILE"
	chmod 0640 "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_LOG_FILE"
}

deploy.securityEmailNotifier.installUnitFile() {
	if [ ! -f "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_UNIT_TEMPLATE" ]; then
		log.error "Missing security-email-notifier systemd unit template: $LITE_NAS_SECURITY_EMAIL_NOTIFIER_UNIT_TEMPLATE"
		exit 1
	fi
	install -D -m 0644 "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_UNIT_TEMPLATE" "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_UNIT_TARGET"
}

deploy.securityEmailNotifier.enableAndStart() {
	deploy.enableAndRefreshService "$LITE_NAS_SECURITY_EMAIL_NOTIFIER_SERVICE_NAME.service"
}

deploy.securityEmailNotifier.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.securityEmailNotifier.ensureRuntimeUser
	deploy.securityEmailNotifier.installBinary "$source_binary"
	deploy.securityEmailNotifier.installConfig
	deploy.securityEmailNotifier.installTemplates
	deploy.securityEmailNotifier.installLogTarget
	deploy.securityEmailNotifier.installUnitFile

	if [ "$should_start" = "1" ]; then
		deploy.securityEmailNotifier.enableAndStart
		return 0
	fi

	log.info "Skipping security-email-notifier enable/start."
}
