#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_INTEGRITY_SERVICE_NAME="${LITE_NAS_INTEGRITY_SERVICE_NAME:-lite-nas-integrity}"
readonly LITE_NAS_INTEGRITY_RUNTIME_USER="${LITE_NAS_INTEGRITY_RUNTIME_USER:-lite-nas-integrity}"
readonly LITE_NAS_INTEGRITY_RUNTIME_GROUP="${LITE_NAS_INTEGRITY_RUNTIME_GROUP:-$LITE_NAS_INTEGRITY_RUNTIME_USER}"
readonly LITE_NAS_INTEGRITY_TRANSPORT_CERT_USER="${LITE_NAS_INTEGRITY_TRANSPORT_CERT_USER:-lite-nas-integrity}"
readonly LITE_NAS_INTEGRITY_IDENTITY_CERT_USER="${LITE_NAS_INTEGRITY_IDENTITY_CERT_USER:-lite-nas-integrity}"
readonly LITE_NAS_INTEGRITY_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_INTEGRITY_SECURITY_GROUP="${LITE_NAS_SECURITY_GROUP:-lite-nas-security}"
readonly LITE_NAS_INTEGRITY_BINARY_TARGET="${LITE_NAS_INTEGRITY_BINARY_TARGET:-/usr/libexec/lite-nas/litenas-integrity}"
readonly LITE_NAS_INTEGRITY_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_INTEGRITY_CONFIG_SOURCE="${LITE_NAS_INTEGRITY_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/litenas-integrity.conf}"
readonly LITE_NAS_INTEGRITY_CONFIG_TARGET="${LITE_NAS_INTEGRITY_CONFIG_TARGET:-$LITE_NAS_INTEGRITY_CONFIG_DIR/litenas-integrity.conf}"
readonly LITE_NAS_INTEGRITY_UNIT_TEMPLATE="${LITE_NAS_INTEGRITY_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-integrity.service}"
readonly LITE_NAS_INTEGRITY_UNIT_TARGET="${LITE_NAS_INTEGRITY_UNIT_TARGET:-/etc/systemd/system/lite-nas-integrity.service}"
readonly LITE_NAS_INTEGRITY_SUDOERS_TEMPLATE="${LITE_NAS_INTEGRITY_SUDOERS_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/sudoers.d/lite-nas-integrity}"
readonly LITE_NAS_INTEGRITY_SUDOERS_TARGET="${LITE_NAS_INTEGRITY_SUDOERS_TARGET:-/etc/sudoers.d/lite-nas-integrity}"
readonly LITE_NAS_INTEGRITY_LOG_DIR="${LITE_NAS_INTEGRITY_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_INTEGRITY_LOG_FILE="${LITE_NAS_INTEGRITY_LOG_FILE:-$LITE_NAS_INTEGRITY_LOG_DIR/lite-nas-integrity.log}"

deploy.liteNASIntegrity.requireTools() {
	local tool
	local tools=(
		chmod
		chown
		getent
		groupadd
		id
		install
		realpath
		systemctl
		useradd
		usermod
		visudo
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.liteNASIntegrity.ensureGroup() {
	local group_name="$1"
	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating system group: $group_name"
	groupadd --system "$group_name"
}

deploy.liteNASIntegrity.ensureRuntimeUser() {
	deploy.liteNASIntegrity.ensureGroup "$LITE_NAS_INTEGRITY_CONFIG_GROUP"
	deploy.liteNASIntegrity.ensureGroup "$LITE_NAS_INTEGRITY_RUNTIME_GROUP"
	deploy.liteNASIntegrity.ensureGroup "$LITE_NAS_INTEGRITY_TRANSPORT_CERT_USER"
	deploy.liteNASIntegrity.ensureGroup "$LITE_NAS_INTEGRITY_SECURITY_GROUP"

	if ! id "$LITE_NAS_INTEGRITY_RUNTIME_USER" >/dev/null 2>&1; then
		log.info "Creating system user: $LITE_NAS_INTEGRITY_RUNTIME_USER"
		useradd \
			--system \
			--no-create-home \
			--home-dir /nonexistent \
			--shell /usr/sbin/nologin \
			--gid "$LITE_NAS_INTEGRITY_RUNTIME_GROUP" \
			--groups "$LITE_NAS_INTEGRITY_CONFIG_GROUP,$LITE_NAS_INTEGRITY_TRANSPORT_CERT_USER,$LITE_NAS_INTEGRITY_SECURITY_GROUP" \
			"$LITE_NAS_INTEGRITY_RUNTIME_USER"
		return 0
	fi

	log.info "Updating system user groups: $LITE_NAS_INTEGRITY_RUNTIME_USER"
	usermod \
		--gid "$LITE_NAS_INTEGRITY_RUNTIME_GROUP" \
		--append \
		--groups "$LITE_NAS_INTEGRITY_CONFIG_GROUP,$LITE_NAS_INTEGRITY_TRANSPORT_CERT_USER,$LITE_NAS_INTEGRITY_SECURITY_GROUP" \
		"$LITE_NAS_INTEGRITY_RUNTIME_USER"
}

deploy.liteNASIntegrity.installBinary() {
	local source_binary="$1"
	local target_dir

	if [ ! -f "$source_binary" ]; then
		log.error "Missing litenas-integrity binary: $source_binary"
		exit 1
	fi

	target_dir="$(dirname "$LITE_NAS_INTEGRITY_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"

	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_INTEGRITY_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_INTEGRITY_BINARY_TARGET"
		return 0
	fi

	install -m 0755 "$source_binary" "$LITE_NAS_INTEGRITY_BINARY_TARGET"
}

deploy.liteNASIntegrity.installConfig() {
	if [ ! -f "$LITE_NAS_INTEGRITY_CONFIG_SOURCE" ]; then
		log.error "Missing litenas-integrity config source: $LITE_NAS_INTEGRITY_CONFIG_SOURCE"
		exit 1
	fi

	install -d -m 0711 -o root -g "$LITE_NAS_INTEGRITY_CONFIG_GROUP" "$LITE_NAS_INTEGRITY_CONFIG_DIR"
	install -m 0640 -o "$LITE_NAS_INTEGRITY_RUNTIME_USER" -g "$LITE_NAS_INTEGRITY_CONFIG_GROUP" \
		"$LITE_NAS_INTEGRITY_CONFIG_SOURCE" \
		"$LITE_NAS_INTEGRITY_CONFIG_TARGET"
}

deploy.liteNASIntegrity.installSudoers() {
	if [ ! -f "$LITE_NAS_INTEGRITY_SUDOERS_TEMPLATE" ]; then
		log.error "Missing litenas-integrity sudoers template: $LITE_NAS_INTEGRITY_SUDOERS_TEMPLATE"
		exit 1
	fi

	deploy.installSudoersTemplate "$LITE_NAS_INTEGRITY_SUDOERS_TEMPLATE" "$LITE_NAS_INTEGRITY_SUDOERS_TARGET"
}

deploy.liteNASIntegrity.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_INTEGRITY_CONFIG_GROUP" "$LITE_NAS_INTEGRITY_LOG_DIR"

	if [ ! -f "$LITE_NAS_INTEGRITY_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_INTEGRITY_RUNTIME_USER" -g "$LITE_NAS_INTEGRITY_RUNTIME_GROUP" \
			/dev/null \
			"$LITE_NAS_INTEGRITY_LOG_FILE"
		return 0
	fi

	chown "$LITE_NAS_INTEGRITY_RUNTIME_USER:$LITE_NAS_INTEGRITY_RUNTIME_GROUP" "$LITE_NAS_INTEGRITY_LOG_FILE"
	chmod 0640 "$LITE_NAS_INTEGRITY_LOG_FILE"
}

deploy.liteNASIntegrity.installUnitFile() {
	if [ ! -f "$LITE_NAS_INTEGRITY_UNIT_TEMPLATE" ]; then
		log.error "Missing litenas-integrity systemd unit template: $LITE_NAS_INTEGRITY_UNIT_TEMPLATE"
		exit 1
	fi

	install -D -m 0644 "$LITE_NAS_INTEGRITY_UNIT_TEMPLATE" "$LITE_NAS_INTEGRITY_UNIT_TARGET"
}

deploy.liteNASIntegrity.ensureCertificates() {
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing --user "$LITE_NAS_INTEGRITY_TRANSPORT_CERT_USER"
	LITE_NAS_INTEGRITY_IDENTITY_USER="$LITE_NAS_INTEGRITY_IDENTITY_CERT_USER" \
		"$LITE_NAS_REPO_ROOT/scripts/rotate-auth-token-certificates.sh" --if-missing
}

deploy.liteNASIntegrity.enableAndStart() {
	deploy.enableAndRefreshService "$LITE_NAS_INTEGRITY_SERVICE_NAME.service"
}

deploy.liteNASIntegrity.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.liteNASIntegrity.ensureRuntimeUser
	deploy.liteNASIntegrity.ensureCertificates
	deploy.liteNASIntegrity.installBinary "$source_binary"
	deploy.liteNASIntegrity.installConfig
	deploy.liteNASIntegrity.installSudoers
	deploy.liteNASIntegrity.installLogTarget
	deploy.liteNASIntegrity.installUnitFile

	if [ "$should_start" = "1" ]; then
		deploy.liteNASIntegrity.enableAndStart
		return 0
	fi

	log.info "Skipping litenas-integrity enable/start."
}
