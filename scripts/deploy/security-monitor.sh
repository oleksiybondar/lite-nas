#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

# The service name is supplied by a fixed entrypoint, never by a runtime request.
deploy.securityMonitor.requireTools() {
	local tool
	for tool in getent groupadd id useradd usermod install chmod chown realpath systemctl visudo rsyslogd; do
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
	deploy.installMergedConfig "$config_source" "$config_target" "$user" lite-nas
	if [ "$name" = "log-analysis" ]; then
		install -d -m 0750 -o "$user" -g lite-nas /etc/lite-nas/log-analysis
		deploy.securityMonitor.retireLegacyLogAnalysisRules "$rules_target"
		if [ ! -f "$rules_target" ]; then
			install -m 0640 -o "$user" -g lite-nas "$rules_source" "$rules_target"
		fi
		install -D -m 0644 -o root -g root "$LITE_NAS_REPO_ROOT/configs/etc/rsyslog.d/30-lite-nas-log-analysis.conf" \
			/etc/rsyslog.d/30-lite-nas-log-analysis.conf
		install -D -m 0644 -o root -g root "$LITE_NAS_REPO_ROOT/configs/etc/apparmor.d/rsyslog.d/lite-nas-log-analysis" \
			/etc/apparmor.d/rsyslog.d/lite-nas-log-analysis
	fi
	install -d -m 0751 -o root -g lite-nas /var/log/lite-nas
	if [ ! -f "$log_target" ]; then
		install -m 0640 -o "$user" -g "$user" /dev/null "$log_target"
	fi
	install -D -m 0644 "$unit_source" "/etc/systemd/system/$user.service"
	if [ "$name" = "log-analysis" ]; then
		install -D -m 0644 "$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/$user.socket" "/etc/systemd/system/$user.socket"
	fi
	if [ "$name" = "service-watchdog" ]; then
		deploy.installSudoersTemplate "$LITE_NAS_REPO_ROOT/configs/etc/sudoers.d/lite-nas-service-watchdog" /etc/sudoers.d/lite-nas-service-watchdog
	fi
	if [ "$should_start" = "1" ]; then
		if [ "$name" = "log-analysis" ]; then
			deploy.securityMonitor.refreshLogAnalysis
		else
			deploy.enableAndRefreshService "$user.service"
		fi
	fi
}

# Log Analysis receives rsyslog records on a systemd-owned datagram socket.
# The service is stopped while the socket is recreated, so an updated socket
# unit takes effect; rsyslog's disk-assisted action queue holds records
# meanwhile.
deploy.securityMonitor.refreshLogAnalysis() {
	if ! deploy.hasUsableSystemd; then
		log.warn "No usable systemd is available for lite-nas-log-analysis; skipping enable/start."
		return 0
	fi
	systemctl enable lite-nas-log-analysis.socket lite-nas-log-analysis.service
	# Stop socket activation before stopping its service, otherwise systemd
	# reports that the triggering socket remains active during the service stop.
	systemctl stop lite-nas-log-analysis.socket
	systemctl stop lite-nas-log-analysis.service
	systemctl start lite-nas-log-analysis.socket
	systemctl start lite-nas-log-analysis.service
	deploy.securityMonitor.activateRsyslogForwarding
}

# rsyslog loads /etc/rsyslog.d only at startup, so the forwarding action is
# active only after a restart. The full configuration is validated first,
# because a broken drop-in would stop host logging. The distribution rsyslogd
# AppArmor profile is reloaded only when it is already loaded, so the
# lite-nas-log-analysis include applies without enforcing a disabled profile.
deploy.securityMonitor.activateRsyslogForwarding() {
	local profiles=/sys/kernel/security/apparmor/profiles
	local rsyslog_profile=/etc/apparmor.d/usr.sbin.rsyslogd

	rsyslogd -N1 >/dev/null
	if [ -r "$profiles" ] && [ -f "$rsyslog_profile" ] && grep -q '^rsyslogd ' "$profiles"; then
		apparmor_parser -r -T -W "$rsyslog_profile"
	fi
	systemctl restart rsyslog.service
}

# Rules written for the removed file-tailing source (`path` with the
# `rsyslog_auth` format) cannot be loaded by the socket ingress. Keep the old
# file for manual review and let deployment install the packaged default.
deploy.securityMonitor.retireLegacyLogAnalysisRules() {
	local rules_target="$1"
	local backup="$rules_target.pre-rsyslog-socket"

	if [ ! -f "$rules_target" ] || ! grep -q '"rsyslog_auth"' "$rules_target"; then
		return 0
	fi
	log.warn "Replacing legacy Log Analysis rules; the previous file is kept at $backup"
	mv -f "$rules_target" "$backup"
}
