#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

# The authenticated CLI is installed for members of the existing security role.
deploy.securityAutomationCLI.deploy() {
	local source_binary="$1"
	local user=lite-nas-security-automation-cli
	local group=lite-nas-security
	local target=/usr/libexec/lite-nas/security-automation-cli
	local config_source="$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/security-automation-cli.conf"
	if [ ! -f "$source_binary" ] || [ ! -f "$config_source" ]; then
		log.error "Missing security automation CLI binary or config"
		return 1
	fi
	if ! getent group "$group" >/dev/null 2>&1; then
		groupadd --system "$group"
	fi
	if ! getent group "$user" >/dev/null 2>&1; then
		groupadd --system "$user"
	fi
	if ! id "$user" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin \
			--gid "$user" --groups "$group" "$user"
	else
		usermod --gid "$user" --append --groups "$group" "$user"
	fi
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing --user "$user"
	install -d -m 0755 /usr/libexec/lite-nas /usr/bin
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$target")" ]; then
		chmod 0755 "$target"
	else
		install -m 0755 "$source_binary" "$target"
	fi
	ln -sfn "$target" /usr/bin/security-automation-cli
	install -d -m 0711 -o root -g lite-nas /etc/lite-nas
	deploy.installMergedConfig "$config_source" /etc/lite-nas/security-automation-cli.conf "$user" "$group"
	install -d -m 0751 -o root -g lite-nas /var/log/lite-nas
	if [ ! -f /var/log/lite-nas/security-automation-cli.log ]; then
		install -m 0660 -o root -g "$group" /dev/null /var/log/lite-nas/security-automation-cli.log
	fi
}
