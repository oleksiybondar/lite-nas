#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

# The service name is supplied by a fixed entrypoint, never by a runtime request.
deploy.securityMonitor.requireTools() {
	local tool
	for tool in getent groupadd id useradd usermod install chmod chown realpath systemctl visudo; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.securityMonitor.ensureGroup() {
	if ! getent group "$1" >/dev/null 2>&1; then
		groupadd --system "$1"
	fi
}

deploy.securityMonitor.deploy() {
	local name="$1"
	local source_binary="$2"
	local should_start="${3:-1}"
	local user="lite-nas-$name"
	local extra_groups="lite-nas,$user"
	local target="/usr/libexec/lite-nas/$name"
	local config_target="/etc/lite-nas/$name.conf"
	local log_target="/var/log/lite-nas/$name.log"
	local config_source="$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/$name.conf"
	local unit_source="$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/$user.service"
	local rules_source="$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/log-analysis/rules.json"
	local rules_target=/etc/lite-nas/log-analysis/rules.json

	case "$name" in
	mount-monitor | antivirus | system-service-monitor | service-watchdog | log-analysis) ;;
	*)
		log.error "Unknown security monitor: $name"
		return 2
		;;
	esac
	if [ ! -f "$source_binary" ] || [ ! -f "$config_source" ] || [ ! -f "$unit_source" ]; then
		log.error "Missing binary, configuration, or unit for $name"
		return 1
	fi
	deploy.securityMonitor.ensureGroup lite-nas
	deploy.securityMonitor.ensureGroup "$user"
	case "$name" in
	mount-monitor | system-service-monitor)
		deploy.securityMonitor.ensureGroup lite-nas-operator
		extra_groups="$extra_groups,lite-nas-operator"
		;;
	antivirus)
		deploy.securityMonitor.ensureGroup lite-nas-operator
		deploy.securityMonitor.ensureGroup lite-nas-security
		extra_groups="$extra_groups,lite-nas-operator,lite-nas-security"
		;;
	service-watchdog | log-analysis)
		deploy.securityMonitor.ensureGroup lite-nas-security
		extra_groups="$extra_groups,lite-nas-security"
		;;
	esac
	if [ "$name" = "log-analysis" ]; then
		extra_groups="$extra_groups,adm"
	fi
	if ! id "$user" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin --gid "$user" --groups "$extra_groups" "$user"
	else
		usermod --gid "$user" --append --groups "$extra_groups" "$user"
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
	if [ ! -f "$config_target" ]; then
		install -m 0640 -o "$user" -g lite-nas "$config_source" "$config_target"
	fi
	if [ "$name" = "log-analysis" ]; then
		install -d -m 0750 -o "$user" -g lite-nas /etc/lite-nas/log-analysis
		if [ ! -f "$rules_target" ]; then
			install -m 0640 -o "$user" -g lite-nas "$rules_source" "$rules_target"
		fi
	fi
	install -d -m 0751 -o root -g lite-nas /var/log/lite-nas
	if [ ! -f "$log_target" ]; then
		install -m 0640 -o "$user" -g "$user" /dev/null "$log_target"
	fi
	install -D -m 0644 "$unit_source" "/etc/systemd/system/$user.service"
	if [ "$name" = "service-watchdog" ]; then
		deploy.installSudoersTemplate "$LITE_NAS_REPO_ROOT/configs/etc/sudoers.d/lite-nas-service-watchdog" /etc/sudoers.d/lite-nas-service-watchdog
	fi
	if [ "$should_start" = "1" ]; then
		deploy.enableAndRefreshService "$user.service"
	fi
}
