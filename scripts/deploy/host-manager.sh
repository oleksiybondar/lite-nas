#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

# Both manager entrypoints pass one fixed service name, not request data.
deploy.hostManager.requireTools() {
	local tool
	for tool in getent groupadd id useradd usermod install chmod realpath systemctl visudo apparmor_parser mktemp rm; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

# Validate the installed profile before loading it; the unit requires the
# profile by name and will fail to start if AppArmor cannot provide it.
deploy.hostManager.installProfile() (
	local name="$1"
	local should_start="$2"
	local source="$LITE_NAS_REPO_ROOT/configs/etc/apparmor.d/usr.libexec.lite-nas.$name"
	local target="/etc/apparmor.d/usr.libexec.lite-nas.$name"
	local cache_dir

	install -D -m 0644 -o root -g root "$source" "$target"
	cache_dir="$(mktemp -d)"
	trap 'rm -rf "$cache_dir"' EXIT
	apparmor_parser -Q -W -L "$cache_dir" -I /etc/apparmor.d "$target"
	if [ "$should_start" = "1" ]; then
		apparmor_parser -r -I /etc/apparmor.d "$target"
	fi
)

deploy.hostManager.deploy() {
	local name="$1"
	local source_binary="$2"
	local should_start="${3:-1}"
	local user="lite-nas-$name"
	local target="/usr/libexec/lite-nas/$name"
	local config_source="$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/$name.conf"
	local unit_source="$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/$user.service"
	local sudoers_source="$LITE_NAS_REPO_ROOT/configs/etc/sudoers.d/$user"
	local profile_source="$LITE_NAS_REPO_ROOT/configs/etc/apparmor.d/usr.libexec.lite-nas.$name"

	case "$name" in
	services-manager | session-manager | journal-manager | usb-manager) ;;
	*)
		log.error "Unknown host manager: $name"
		return 2
		;;
	esac
	if [ ! -f "$source_binary" ] || [ ! -f "$config_source" ] || [ ! -f "$unit_source" ] || [ ! -f "$sudoers_source" ] || [ ! -f "$profile_source" ]; then
		log.error "Missing host-manager deployment input for $name"
		return 1
	fi
	if ! getent group lite-nas >/dev/null 2>&1; then groupadd --system lite-nas; fi
	if ! getent group "$user" >/dev/null 2>&1; then groupadd --system "$user"; fi
	if ! id "$user" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin --gid "$user" --groups lite-nas "$user"
	else
		usermod --gid "$user" --append --groups lite-nas "$user"
	fi
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing --user "$user"
	"$LITE_NAS_REPO_ROOT/scripts/rotate-auth-token-certificates.sh" --if-missing
	install -d -m 0755 /usr/libexec/lite-nas
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$target")" ]; then
		chmod 0755 "$target"
	else
		install -m 0755 "$source_binary" "$target"
	fi
	install -d -m 0711 -o root -g lite-nas /etc/lite-nas
	deploy.installMergedConfig "$config_source" "/etc/lite-nas/$name.conf" "$user" lite-nas
	install -d -m 0751 -o root -g lite-nas /var/log/lite-nas
	if [ ! -f "/var/log/lite-nas/$name.log" ]; then
		install -m 0640 -o "$user" -g "$user" /dev/null "/var/log/lite-nas/$name.log"
	fi
	install -D -m 0644 "$unit_source" "/etc/systemd/system/$user.service"
	deploy.installSudoersTemplate "$sudoers_source" "/etc/sudoers.d/$user"
	deploy.hostManager.installProfile "$name" "$should_start"
	if [ "$should_start" = "1" ]; then deploy.enableAndRefreshService "$user.service"; fi
}
