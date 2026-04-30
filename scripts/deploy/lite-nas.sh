#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/ufw.sh"

readonly LITE_NAS_BOOTSTRAP_GROUP="${LITE_NAS_GROUP:-lite-nas}"

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
		systemctl
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done

	deploy.ufw.requireTools
}

deploy.liteNAS.ensureCommonGroup() {
	if getent group "$LITE_NAS_BOOTSTRAP_GROUP" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating common LiteNAS group: $LITE_NAS_BOOTSTRAP_GROUP"
	groupadd --system "$LITE_NAS_BOOTSTRAP_GROUP"
}

deploy.liteNAS.bootstrap() {
	local manage_nats_config="$1"

	"$LITE_NAS_REPO_ROOT/scripts/install-runtime-dependencies.sh"
	deploy.liteNAS.ensureCommonGroup

	if [ "$manage_nats_config" = "1" ]; then
		"$LITE_NAS_REPO_ROOT/scripts/deploy-configs.sh" --no-restart
		"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing
		systemctl restart nats-server.service
	else
		log.warn "Skipping NATS config replacement; LiteNAS services may require manual NATS configuration."
	fi

	"$LITE_NAS_REPO_ROOT/scripts/rotate-nginx-certificates.sh" --if-missing
	"$LITE_NAS_REPO_ROOT/scripts/rotate-auth-token-certificates.sh" --if-missing
	deploy.ufw.deploy
}
