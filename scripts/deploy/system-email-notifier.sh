#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_SERVICE_NAME="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_SERVICE_NAME:-lite-nas-system-email-notifier}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER:-lite-nas-sys-email-notifier}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_GROUP="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_GROUP:-$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_ACCESS_GROUP="${LITE_NAS_OPERATOR_GROUP:-lite-nas-operator}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_BINARY_TARGET="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_BINARY_TARGET:-/usr/libexec/lite-nas/system-email-notifier}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_SOURCE="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/system-email-notifier.conf}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_TARGET="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_TARGET:-$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_DIR/system-email-notifier.conf}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_UNIT_TEMPLATE="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-system-email-notifier.service}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_UNIT_TARGET="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_UNIT_TARGET:-/etc/systemd/system/lite-nas-system-email-notifier.service}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_DIR="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_FILE="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_FILE:-$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_DIR/system-email-notifier.log}"

deploy.systemEmailNotifier.requireTools() {
	local tool
	local tools=(getent groupadd install chmod chown realpath systemctl useradd usermod id)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.systemEmailNotifier.ensureGroup() {
	local group_name="$1"
	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi
	groupadd --system "$group_name"
}

deploy.systemEmailNotifier.ensureRuntimeUser() {
	deploy.systemEmailNotifier.ensureGroup "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_GROUP"
	deploy.systemEmailNotifier.ensureGroup "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_GROUP"
	deploy.systemEmailNotifier.ensureGroup "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_ACCESS_GROUP"
	if ! id "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin \
			--gid "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_GROUP" \
			--groups "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_GROUP,$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_ACCESS_GROUP" \
			"$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER"
		return 0
	fi

	usermod \
		--gid "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_GROUP" \
		--append \
		--groups "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_GROUP,$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_ACCESS_GROUP" \
		"$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER"
}

deploy.systemEmailNotifier.installBinary() {
	local source_binary="$1"
	local target_dir
	if [ ! -f "$source_binary" ]; then
		log.error "Missing system-email-notifier binary: $source_binary"
		exit 1
	fi
	target_dir="$(dirname "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_BINARY_TARGET"
		return 0
	fi
	install -m 0755 "$source_binary" "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_BINARY_TARGET"
}

deploy.systemEmailNotifier.installConfig() {
	if [ ! -f "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_SOURCE" ]; then
		log.error "Missing system-email-notifier config source: $LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_SOURCE"
		exit 1
	fi
	install -d -m 0711 -o root -g "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_GROUP" "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_DIR"
	install -m 0640 -o "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER" -g "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_GROUP" \
		"$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_SOURCE" \
		"$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_TARGET"
}

deploy.systemEmailNotifier.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_CONFIG_GROUP" "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_DIR"
	if [ ! -f "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER" -g "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_GROUP" \
			/dev/null "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_FILE"
		return 0
	fi
	chown "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER:$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_GROUP" "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_FILE"
	chmod 0640 "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_LOG_FILE"
}

deploy.systemEmailNotifier.installUnitFile() {
	if [ ! -f "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_UNIT_TEMPLATE" ]; then
		log.error "Missing system-email-notifier systemd unit template: $LITE_NAS_SYSTEM_EMAIL_NOTIFIER_UNIT_TEMPLATE"
		exit 1
	fi
	install -D -m 0644 "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_UNIT_TEMPLATE" "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_UNIT_TARGET"
}

deploy.systemEmailNotifier.enableAndStart() {
	systemctl daemon-reload
	systemctl enable "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_SERVICE_NAME.service"
	systemctl restart "$LITE_NAS_SYSTEM_EMAIL_NOTIFIER_SERVICE_NAME.service"
}

deploy.systemEmailNotifier.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.systemEmailNotifier.ensureRuntimeUser
	deploy.systemEmailNotifier.installBinary "$source_binary"
	deploy.systemEmailNotifier.installConfig
	deploy.systemEmailNotifier.installLogTarget
	deploy.systemEmailNotifier.installUnitFile

	if [ "$should_start" = "1" ]; then
		deploy.systemEmailNotifier.enableAndStart
		return 0
	fi

	log.info "Skipping system-email-notifier enable/start."
}
