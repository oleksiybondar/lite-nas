#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_RBAC_SERVICE_NAME="${LITE_NAS_RBAC_SERVICE_NAME:-lite-nas-rbac}"
readonly LITE_NAS_RBAC_RUNTIME_USER="${LITE_NAS_RBAC_RUNTIME_USER:-lite-nas-rbac}"
readonly LITE_NAS_RBAC_RUNTIME_GROUP="${LITE_NAS_RBAC_RUNTIME_GROUP:-$LITE_NAS_RBAC_RUNTIME_USER}"
readonly LITE_NAS_RBAC_CERT_USER="${LITE_NAS_RBAC_CERT_USER:-lite-nas-rbac-service}"
readonly LITE_NAS_RBAC_CONFIG_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_RBAC_BINARY_TARGET="${LITE_NAS_RBAC_BINARY_TARGET:-/usr/libexec/lite-nas/rbac-service}"
readonly LITE_NAS_RBAC_CONFIG_DIR="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
readonly LITE_NAS_RBAC_CONFIG_SOURCE="${LITE_NAS_RBAC_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/rbac-service.conf}"
readonly LITE_NAS_RBAC_CONFIG_TARGET="${LITE_NAS_RBAC_CONFIG_TARGET:-$LITE_NAS_RBAC_CONFIG_DIR/rbac-service.conf}"
readonly LITE_NAS_RBAC_UNIT_TEMPLATE="${LITE_NAS_RBAC_UNIT_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-rbac.service}"
readonly LITE_NAS_RBAC_UNIT_TARGET="${LITE_NAS_RBAC_UNIT_TARGET:-/etc/systemd/system/lite-nas-rbac.service}"
readonly LITE_NAS_RBAC_SUDOERS_TEMPLATE="${LITE_NAS_RBAC_SUDOERS_TEMPLATE:-$LITE_NAS_REPO_ROOT/configs/etc/sudoers.d/lite-nas-rbac}"
readonly LITE_NAS_RBAC_SUDOERS_TARGET="${LITE_NAS_RBAC_SUDOERS_TARGET:-/etc/sudoers.d/lite-nas-rbac}"
readonly LITE_NAS_RBAC_LOG_DIR="${LITE_NAS_RBAC_LOG_DIR:-/var/log/lite-nas}"
readonly LITE_NAS_RBAC_LOG_FILE="${LITE_NAS_RBAC_LOG_FILE:-$LITE_NAS_RBAC_LOG_DIR/rbac-service.log}"

deploy.rbacService.usage() {
	cat <<'MSG'
Usage: scripts/deploy-rbac-service.sh [options]

Options:
  --binary PATH       Install an existing binary instead of building one.
  --no-start          Install files but do not enable or start the service.
  --skip-bootstrap    Install files without running LiteNAS bootstrap first.
  -h, --help          Show this help.
MSG
}

deploy.rbacService.requireTools() {
	local tool
	local tools=(
		cat
		chmod
		chown
		getent
		groupadd
		id
		install
		realpath
		rm
		systemctl
		touch
		useradd
		usermod
		visudo
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.rbacService.ensureGroup() {
	local group_name="$1"

	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating system group: $group_name"
	groupadd --system "$group_name"
}

deploy.rbacService.ensureRuntimeUser() {
	deploy.rbacService.ensureGroup "$LITE_NAS_RBAC_CONFIG_GROUP"
	deploy.rbacService.ensureGroup "$LITE_NAS_RBAC_RUNTIME_GROUP"
	deploy.rbacService.ensureGroup "$LITE_NAS_RBAC_CERT_USER"

	if ! id "$LITE_NAS_RBAC_RUNTIME_USER" >/dev/null 2>&1; then
		log.info "Creating system user: $LITE_NAS_RBAC_RUNTIME_USER"
		useradd \
			--system \
			--no-create-home \
			--home-dir /nonexistent \
			--shell /usr/sbin/nologin \
			--gid "$LITE_NAS_RBAC_RUNTIME_GROUP" \
			--groups "$LITE_NAS_RBAC_CONFIG_GROUP,$LITE_NAS_RBAC_CERT_USER" \
			"$LITE_NAS_RBAC_RUNTIME_USER"
		return 0
	fi

	log.info "Updating system user groups: $LITE_NAS_RBAC_RUNTIME_USER"
	usermod \
		--gid "$LITE_NAS_RBAC_RUNTIME_GROUP" \
		--append \
		--groups "$LITE_NAS_RBAC_CONFIG_GROUP,$LITE_NAS_RBAC_CERT_USER" \
		"$LITE_NAS_RBAC_RUNTIME_USER"
}

deploy.rbacService.installBinary() {
	local source_binary="$1"
	local target_dir

	if [ ! -f "$source_binary" ]; then
		log.error "Missing rbac-service binary: $source_binary"
		exit 1
	fi

	target_dir="$(dirname "$LITE_NAS_RBAC_BINARY_TARGET")"
	install -d -m 0755 "$target_dir"

	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_RBAC_BINARY_TARGET")" ]; then
		chmod 0755 "$LITE_NAS_RBAC_BINARY_TARGET"
		return 0
	fi

	install -m 0755 "$source_binary" "$LITE_NAS_RBAC_BINARY_TARGET"
}

deploy.rbacService.installConfig() {
	if [ ! -f "$LITE_NAS_RBAC_CONFIG_SOURCE" ]; then
		log.error "Missing rbac-service config source: $LITE_NAS_RBAC_CONFIG_SOURCE"
		exit 1
	fi

	install -d -m 0711 -o root -g "$LITE_NAS_RBAC_CONFIG_GROUP" "$LITE_NAS_RBAC_CONFIG_DIR"
	install -m 0640 -o "$LITE_NAS_RBAC_RUNTIME_USER" -g "$LITE_NAS_RBAC_CONFIG_GROUP" \
		"$LITE_NAS_RBAC_CONFIG_SOURCE" \
		"$LITE_NAS_RBAC_CONFIG_TARGET"
}

deploy.rbacService.installSudoers() {
	if [ ! -f "$LITE_NAS_RBAC_SUDOERS_TEMPLATE" ]; then
		log.error "Missing rbac sudoers template: $LITE_NAS_RBAC_SUDOERS_TEMPLATE"
		exit 1
	fi

	install -d -m 0750 -o root -g root "$(dirname "$LITE_NAS_RBAC_SUDOERS_TARGET")"
	install -m 0440 -o root -g root "$LITE_NAS_RBAC_SUDOERS_TEMPLATE" "$LITE_NAS_RBAC_SUDOERS_TARGET"
	visudo -c -f "$LITE_NAS_RBAC_SUDOERS_TARGET" >/dev/null
}

deploy.rbacService.installLogTarget() {
	install -d -m 0751 -o root -g "$LITE_NAS_RBAC_CONFIG_GROUP" "$LITE_NAS_RBAC_LOG_DIR"

	if [ ! -f "$LITE_NAS_RBAC_LOG_FILE" ]; then
		install -m 0640 -o "$LITE_NAS_RBAC_RUNTIME_USER" -g "$LITE_NAS_RBAC_RUNTIME_GROUP" \
			/dev/null \
			"$LITE_NAS_RBAC_LOG_FILE"
		return 0
	fi

	chown "$LITE_NAS_RBAC_RUNTIME_USER:$LITE_NAS_RBAC_RUNTIME_GROUP" "$LITE_NAS_RBAC_LOG_FILE"
	chmod 0640 "$LITE_NAS_RBAC_LOG_FILE"
}

deploy.rbacService.installUnitFile() {
	if [ ! -f "$LITE_NAS_RBAC_UNIT_TEMPLATE" ]; then
		log.error "Missing systemd unit template: $LITE_NAS_RBAC_UNIT_TEMPLATE"
		exit 1
	fi

	install -D -m 0644 "$LITE_NAS_RBAC_UNIT_TEMPLATE" "$LITE_NAS_RBAC_UNIT_TARGET"
}

deploy.rbacService.ensureCertificates() {
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing --user "$LITE_NAS_RBAC_CERT_USER"
}

deploy.rbacService.enableAndStart() {
	deploy.enableAndRefreshService "$LITE_NAS_RBAC_SERVICE_NAME.service"
}

deploy.rbacService.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"

	deploy.rbacService.ensureRuntimeUser
	deploy.rbacService.ensureCertificates
	deploy.rbacService.installBinary "$source_binary"
	deploy.rbacService.installConfig
	deploy.rbacService.installSudoers
	deploy.rbacService.installLogTarget
	deploy.rbacService.installUnitFile

	if [ "$should_start" = "1" ]; then
		deploy.rbacService.enableAndStart
		return 0
	fi

	log.info "Skipping rbac-service enable/start."
}
