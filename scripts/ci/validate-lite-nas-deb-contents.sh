#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

if [ "$#" -ne 2 ]; then
	log.error "Usage: scripts/ci/validate-lite-nas-deb-contents.sh <package.deb> <amd64|arm64>"
	exit 64
fi

package_path="$1"
target_arch="$2"

if [ ! -f "$package_path" ]; then
	log.error "Missing package file: $package_path"
	exit 1
fi

package_path="$(realpath "$package_path")"

case "$target_arch" in
amd64 | arm64) ;;
*)
	log.error "Unsupported architecture: $target_arch"
	exit 64
	;;
esac

log.requireCommand "dpkg-deb" "Install dpkg-deb and retry."

temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

package_root="$temp_dir/package-root"
package_control="$temp_dir/package-control"
checks_passed=0
packaged_config_root="usr/libexec/lite-nas/configs/etc"

assert_cmd() {
	local description="$1"
	shift

	log.info "CHECK: $description"
	if "$@"; then
		log.info "PASS: $description"
		checks_passed=$((checks_passed + 1))
		return 0
	fi

	log.error "FAIL: $description"
	log.error "Command: $*"
	exit 1
}

assert_file() {
	local path="$1"
	test -f "$package_root/$path"
}

assert_dir() {
	local path="$1"
	test -d "$package_root/$path"
}

assert_executable() {
	local path="$1"
	test -x "$package_root/$path"
}

assert_symlink_target() {
	local path="$1"
	local expected_target="$2"
	test -L "$package_root/$path"
	test "$(readlink "$package_root/$path")" = "$expected_target"
}

assert_control_contains() {
	local field="$1"
	local expected="$2"
	dpkg-deb -f "$package_path" "$field" | grep -Fqx "$expected"
}

assert_depends_contains() {
	local expected="$1"
	dpkg-deb -f "$package_path" Depends | grep -Fq "$expected"
}

log.pushTask "Validating LiteNAS Debian package contents for ${target_arch}"
dpkg-deb -x "$package_path" "$package_root"
dpkg-deb -e "$package_path" "$package_control"

assert_cmd "package architecture matches ${target_arch}" assert_control_contains Architecture "$target_arch"
assert_cmd "package name is lite-nas" assert_control_contains Package "lite-nas"
assert_cmd "package depends on postfix" assert_depends_contains "postfix"
assert_cmd "package depends on sudo" assert_depends_contains "sudo"
assert_cmd "package depends on aide" assert_depends_contains "aide"
assert_cmd "package includes postinst maintainer script" test -f "$package_control/postinst"

assert_cmd "auth-service binary packaged" assert_executable usr/libexec/lite-nas/auth-service
assert_cmd "rbac-service binary packaged" assert_executable usr/libexec/lite-nas/rbac-service
assert_cmd "system-logging-manager binary packaged" assert_executable usr/libexec/lite-nas/system-logging-manager
assert_cmd "security-logging-manager binary packaged" assert_executable usr/libexec/lite-nas/security-logging-manager
assert_cmd "system-email-notifier binary packaged" assert_executable usr/libexec/lite-nas/system-email-notifier
assert_cmd "security-email-notifier binary packaged" assert_executable usr/libexec/lite-nas/security-email-notifier
assert_cmd "system-metrics binary packaged" assert_executable usr/libexec/lite-nas/system-metrics
assert_cmd "zfs-metrics binary packaged" assert_executable usr/libexec/lite-nas/zfs-metrics
assert_cmd "resources-monitor binary packaged" assert_executable usr/libexec/lite-nas/resources-monitor
assert_cmd "system-logging-manager-cli binary packaged" assert_executable usr/libexec/lite-nas/system-logging-manager-cli
assert_cmd "security-logging-manager-cli binary packaged" assert_executable usr/libexec/lite-nas/security-logging-manager-cli
assert_cmd "system-metrics-cli binary packaged" assert_executable usr/libexec/lite-nas/system-metrics-cli
assert_cmd "zfs-metrics-cli binary packaged" assert_executable usr/libexec/lite-nas/zfs-metrics-cli
assert_cmd "web-gateway binary packaged" assert_executable usr/libexec/lite-nas/web-gateway

assert_cmd "system-logging-manager-cli symlink packaged" assert_symlink_target usr/bin/system-logging-manager-cli /usr/libexec/lite-nas/system-logging-manager-cli
assert_cmd "security-logging-manager-cli symlink packaged" assert_symlink_target usr/bin/security-logging-manager-cli /usr/libexec/lite-nas/security-logging-manager-cli
assert_cmd "system-metrics-cli symlink packaged" assert_symlink_target usr/bin/system-metrics-cli /usr/libexec/lite-nas/system-metrics-cli
assert_cmd "zfs-metrics-cli symlink packaged" assert_symlink_target usr/bin/zfs-metrics-cli /usr/libexec/lite-nas/zfs-metrics-cli

assert_cmd "packaged lite-nas config directory exists" assert_dir "$packaged_config_root/lite-nas"
assert_cmd "auth.conf packaged" assert_file "$packaged_config_root/lite-nas/auth.conf"
assert_cmd "rbac-service.conf packaged" assert_file "$packaged_config_root/lite-nas/rbac-service.conf"
assert_cmd "system-logging-manager.conf packaged" assert_file "$packaged_config_root/lite-nas/system-logging-manager.conf"
assert_cmd "security-logging-manager.conf packaged" assert_file "$packaged_config_root/lite-nas/security-logging-manager.conf"
assert_cmd "system-metrics.conf packaged" assert_file "$packaged_config_root/lite-nas/system-metrics.conf"
assert_cmd "zfs-metrics.conf packaged" assert_file "$packaged_config_root/lite-nas/zfs-metrics.conf"
assert_cmd "system-logging-manager-cli.conf packaged" assert_file "$packaged_config_root/lite-nas/system-logging-manager-cli.conf"
assert_cmd "security-logging-manager-cli.conf packaged" assert_file "$packaged_config_root/lite-nas/security-logging-manager-cli.conf"
assert_cmd "system-email-notifier.conf packaged" assert_file "$packaged_config_root/lite-nas/system-email-notifier.conf"
assert_cmd "security-email-notifier.conf packaged" assert_file "$packaged_config_root/lite-nas/security-email-notifier.conf"
assert_cmd "resources-monitor.conf packaged" assert_file "$packaged_config_root/lite-nas/resources-monitor.conf"
assert_cmd "zfs-metrics-cli.conf packaged" assert_file "$packaged_config_root/lite-nas/zfs-metrics-cli.conf"
assert_cmd "system metrics rule packaged" assert_file "$packaged_config_root/lite-nas/resources-monitor/rules/system-metrics.json"
assert_cmd "zfs metrics rule packaged" assert_file "$packaged_config_root/lite-nas/resources-monitor/rules/zfs-metrics.json"
assert_cmd "web-gateway.conf packaged" assert_file "$packaged_config_root/lite-nas/web-gateway.conf"

assert_cmd "resources-monitor unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-resources-monitor.service"
assert_cmd "system logging manager unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-system-logging-manager.service"
assert_cmd "security logging manager unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-security-logging-manager.service"
assert_cmd "system email notifier unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-system-email-notifier.service"
assert_cmd "security email notifier unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-security-email-notifier.service"
assert_cmd "rbac unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-rbac.service"
assert_cmd "zfs-metrics unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-zfs-metrics.service"

assert_cmd "postfix main.cf template packaged" assert_file "$packaged_config_root/postfix/main.cf"
assert_cmd "postfix master.cf template packaged" assert_file "$packaged_config_root/postfix/master.cf"
assert_cmd "postfix AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.lib.postfix.sbin"
assert_cmd "auth AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.auth-service"
assert_cmd "rbac AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.rbac-service"
assert_cmd "system-metrics AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-metrics"
assert_cmd "zfs-metrics AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.zfs-metrics"
assert_cmd "resources-monitor AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.resources-monitor"
assert_cmd "system-logging-manager AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-logging-manager"
assert_cmd "security-logging-manager AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.security-logging-manager"
assert_cmd "system-email-notifier AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-email-notifier"
assert_cmd "security-email-notifier AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.security-email-notifier"
assert_cmd "web-gateway AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.web-gateway"
assert_cmd "system-metrics-cli AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-metrics-cli"
assert_cmd "zfs-metrics-cli AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.zfs-metrics-cli"
assert_cmd "system-logging-manager-cli AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-logging-manager-cli"
assert_cmd "security-logging-manager-cli AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.security-logging-manager-cli"
assert_cmd "nginx AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.sbin.nginx"

assert_cmd "PAM config template packaged" assert_file "$packaged_config_root/pam.d/litenas-auth"
assert_cmd "nginx site config template packaged" assert_file "$packaged_config_root/nginx/sites-available/lite-nas-web-gateway.conf"
assert_cmd "ufw default config template packaged" assert_file "$packaged_config_root/default/ufw"
assert_cmd "ufw config template packaged" assert_file "$packaged_config_root/ufw/ufw.conf"

assert_cmd "web-gateway index.html packaged" assert_file usr/share/lite-nas/web-gateway/assets/index.html
assert_cmd "web-gateway index.css packaged" assert_file usr/share/lite-nas/web-gateway/assets/index.css
assert_cmd "web-gateway index.js packaged" assert_file usr/share/lite-nas/web-gateway/assets/index.js
log.popTask

log.info "Package content checks passed: $checks_passed"
log.info "Validated static package contents: $package_path"
