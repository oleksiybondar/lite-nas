#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/ufw.sh"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/apparmor.sh"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/postfix.sh"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/normalize-etc-permissions.sh"

readonly LITE_NAS_BOOTSTRAP_GROUP="${LITE_NAS_GROUP:-lite-nas}"
readonly LITE_NAS_SECURITY_GROUP="${LITE_NAS_SECURITY_GROUP:-lite-nas-security}"
readonly LITE_NAS_OPERATOR_GROUP="${LITE_NAS_OPERATOR_GROUP:-lite-nas-operator}"
readonly LITE_NAS_SERVICE_ADMIN_GROUP="${LITE_NAS_SERVICE_ADMIN_GROUP:-lite-nas-svc-adm}"
readonly LITE_NAS_ZFS_ADMIN_GROUP="${LITE_NAS_ZFS_ADMIN_GROUP:-zfs-adm}"
readonly LITE_NAS_BOOTSTRAP_LOG_DIR="${LITE_NAS_BOOTSTRAP_LOG_DIR:-/var/log/lite-nas}"

deploy.liteNAS.usage() {
	cat <<'MSG'
Usage: scripts/deploy-lite-nas.sh [options]

Options:
  --no-nats-config  Keep the current NATS configuration unchanged.
  -h, --help        Show this help.
MSG
}

deploy.liteNAS.requireTools() {
	local tool
	local tools=(
		getent
		groupadd
		install
		ln
		usermod
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done

	deploy.ufw.requireTools
}

deploy.liteNAS.ensureCommonGroup() {
	deploy.ensureSystemGroup "$LITE_NAS_BOOTSTRAP_GROUP" "common LiteNAS group"
}

deploy.liteNAS.ensureRoleGroups() {
	deploy.ensureSystemGroup "$LITE_NAS_SECURITY_GROUP" "LiteNAS security group"
	deploy.ensureSystemGroup "$LITE_NAS_OPERATOR_GROUP" "LiteNAS operator group"
	deploy.ensureSystemGroup "$LITE_NAS_SERVICE_ADMIN_GROUP" "LiteNAS service admin group"
	deploy.ensureSystemGroup "$LITE_NAS_ZFS_ADMIN_GROUP" "ZFS admin group"
}

# Give every regular account operator access. UID_MIN follows the host's
# account policy; Debian's standard fallback is 1000 when login.defs omits it.
deploy.liteNAS.addRegularUsersToOperatorGroup() {
	local uid_min
	local login _ uid _rest

	uid_min="$(awk '$1 == "UID_MIN" { print $2; exit }' /etc/login.defs 2>/dev/null || true)"
	if [[ ! "$uid_min" =~ ^[0-9]+$ ]]; then
		uid_min=1000
	fi

	while IFS=: read -r login _ uid _rest; do
		if [[ "$uid" =~ ^[0-9]+$ ]] && ((uid >= uid_min)); then
			usermod --append --groups "$LITE_NAS_OPERATOR_GROUP" "$login"
		fi
	done < <(getent passwd)
}

deploy.liteNAS.ensureLogDir() {
	install -d -m 0751 -o root -g "$LITE_NAS_BOOTSTRAP_GROUP" "$LITE_NAS_BOOTSTRAP_LOG_DIR"
}

# Install the root-owned password rotator behind its public command name.
deploy.liteNAS.installPasswordRotator() {
	local script_target=/usr/libexec/lite-nas/scripts/rotate-user-password.sh
	local command_target=/usr/bin/lite-nas-rotate-password

	install -D -m 0755 -o root -g root \
		"$LITE_NAS_REPO_ROOT/scripts/rotate-user-password.sh" "$script_target"
	ln -sfn "$script_target" "$command_target"
}

deploy.liteNAS.bootstrap() {
	local manage_nats_config="$1"
	local config_deploy_args=(--no-restart)

	"$LITE_NAS_REPO_ROOT/scripts/install-runtime-dependencies.sh"
	deploy.liteNAS.ensureCommonGroup
	deploy.liteNAS.ensureRoleGroups
	deploy.liteNAS.ensureLogDir
	deploy.apparmor.requireTools
	deploy.postfix.requireTools

	if [ "$manage_nats_config" != "1" ]; then
		log.warn "Skipping NATS config replacement; LiteNAS services may require manual NATS configuration."
		config_deploy_args+=(--no-nats-config)
	fi
	"$LITE_NAS_REPO_ROOT/scripts/deploy-configs.sh" "${config_deploy_args[@]}"

	if [ "$manage_nats_config" = "1" ]; then
		"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing
		if deploy.hasUsableSystemd; then
			systemctl restart nats-server.service
		elif deploy.hasServiceCommand && service nats-server restart >/dev/null 2>&1; then
			:
		else
			log.warn "No usable service manager is available for nats-server.service; skipping restart."
		fi
	fi

	deploy.apparmor.deploy 1
	deploy.postfix.deploy 1
	deploy.normalizeEtcPermissions /etc

	"$LITE_NAS_REPO_ROOT/scripts/rotate-nginx-certificates.sh" --if-missing
	"$LITE_NAS_REPO_ROOT/scripts/rotate-auth-token-certificates.sh" --if-missing
	deploy.ufw.deploy
}
