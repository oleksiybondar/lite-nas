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

assert_file_contains() {
	local path="$1"
	local expected_line="$2"
	grep -Fq -- "$expected_line" "$package_root/$path"
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
assert_cmd "package depends on AppArmor" assert_depends_contains "apparmor"
assert_cmd "package depends on OS ffmpeg" assert_depends_contains "ffmpeg"
assert_cmd "package depends on HEIC decoder tools" assert_depends_contains "libheif-examples"
assert_cmd "package depends on HEIC HEVC plugin" assert_depends_contains "libheif-plugin-libde265"
assert_cmd "package depends on aide" assert_depends_contains "aide"
assert_cmd "package includes postinst maintainer script" test -f "$package_control/postinst"
assert_cmd "postinst invokes package runtime" grep -Fq "\"\$DEPLOY_RUNTIME_SCRIPT\" --mode \"\$mode\" \"\$@\"" "$package_control/postinst"

assert_cmd "auth-service binary packaged" assert_executable usr/libexec/lite-nas/auth-service
assert_cmd "rbac-service binary packaged" assert_executable usr/libexec/lite-nas/rbac-service
assert_cmd "system-logging-manager binary packaged" assert_executable usr/libexec/lite-nas/system-logging-manager
assert_cmd "security-logging-manager binary packaged" assert_executable usr/libexec/lite-nas/security-logging-manager
assert_cmd "system-email-notifier binary packaged" assert_executable usr/libexec/lite-nas/system-email-notifier
assert_cmd "security-email-notifier binary packaged" assert_executable usr/libexec/lite-nas/security-email-notifier
assert_cmd "network-metrics binary packaged" assert_executable usr/libexec/lite-nas/network-metrics
assert_cmd "process-metrics binary packaged" assert_executable usr/libexec/lite-nas/process-metrics
assert_cmd "disk-metrics binary packaged" assert_executable usr/libexec/lite-nas/disk-metrics
assert_cmd "service-metrics binary packaged" assert_executable usr/libexec/lite-nas/service-metrics
assert_cmd "system-metrics binary packaged" assert_executable usr/libexec/lite-nas/system-metrics
assert_cmd "zfs-metrics binary packaged" assert_executable usr/libexec/lite-nas/zfs-metrics
assert_cmd "resources-monitor binary packaged" assert_executable usr/libexec/lite-nas/resources-monitor
assert_cmd "litenas-integrity binary packaged" assert_executable usr/libexec/lite-nas/litenas-integrity
assert_cmd "system-logging-manager-cli binary packaged" assert_executable usr/libexec/lite-nas/system-logging-manager-cli
assert_cmd "security-logging-manager-cli binary packaged" assert_executable usr/libexec/lite-nas/security-logging-manager-cli
assert_cmd "system-metrics-cli binary packaged" assert_executable usr/libexec/lite-nas/system-metrics-cli
assert_cmd "network-metrics-cli binary packaged" assert_executable usr/libexec/lite-nas/network-metrics-cli
assert_cmd "zfs-metrics-cli binary packaged" assert_executable usr/libexec/lite-nas/zfs-metrics-cli
assert_cmd "file-writer binary packaged" assert_executable usr/libexec/lite-nas/lite-nas-file-writer
assert_cmd "file-service binary packaged" assert_executable usr/libexec/lite-nas/file-service
assert_cmd "stream-service binary packaged" assert_executable usr/libexec/lite-nas/stream-service
assert_cmd "web-gateway binary packaged" assert_executable usr/libexec/lite-nas/web-gateway
assert_cmd "Moments binary packaged" assert_executable usr/libexec/lite-nas/moments
assert_cmd "Moments frontend packaged" assert_file usr/libexec/lite-nas/moments-assets/index.html

assert_cmd "package runtime deploy entrypoint packaged" assert_executable usr/libexec/lite-nas/scripts/runtime/deploy-package-runtime.sh
assert_cmd "file-writer deploy helper packaged" assert_file usr/libexec/lite-nas/scripts/deploy/file-writer.sh
assert_cmd "file-service deploy helper packaged" assert_file usr/libexec/lite-nas/scripts/deploy/file-service.sh
assert_cmd "stream-service deploy helper packaged" assert_file usr/libexec/lite-nas/scripts/deploy/stream-service.sh
assert_cmd "stream-service sudoers packaged" assert_file "$packaged_config_root/sudoers.d/lite-nas-stream-service-read"
assert_cmd "stream-service preview helper is permitted" assert_file_contains "$packaged_config_root/sudoers.d/lite-nas-stream-service-read" '--preview-file *'
assert_cmd "file-service deploy provisions read ACLs" assert_file_contains usr/libexec/lite-nas/scripts/deploy/file-service.sh 'deploy.fileService.provisionReadACL'
assert_cmd "file-service deploy rotates its NATS certificate" assert_file_contains usr/libexec/lite-nas/scripts/deploy/file-service.sh 'rotate-nats-certificates.sh'
assert_cmd "file-writer deploy installs the AppArmor profile" assert_file_contains usr/libexec/lite-nas/scripts/deploy/file-writer.sh "LITE_NAS_FILE_WRITER_PROFILE_TARGET=\"\${LITE_NAS_FILE_WRITER_PROFILE_TARGET:-/etc/apparmor.d/usr.libexec.lite-nas.lite-nas-file-writer}\""
assert_cmd "file-writer deploy reloads the AppArmor profile" assert_file_contains usr/libexec/lite-nas/scripts/deploy/file-writer.sh 'apparmor_parser -r'
assert_cmd "file-writer deploy validates sudoers" assert_file_contains usr/libexec/lite-nas/scripts/deploy/file-writer.sh 'deploy.validateSudoersDropIn'
assert_cmd "deploy configs entrypoint packaged" assert_executable usr/libexec/lite-nas/scripts/deploy-configs.sh
assert_cmd "service-admin kill helper packaged" assert_executable usr/libexec/lite-nas/scripts/helpers/lite-nas-kill-pid.sh
assert_cmd "deploy configs installs service-admin kill helper" assert_file_contains usr/libexec/lite-nas/scripts/deploy-configs.sh "kill_helper_target=\"\${LITE_NAS_KILL_HELPER_TARGET:-/usr/libexec/lite-nas/lite-nas-kill-pid}\""
assert_cmd "package deploy applies configs with NATS preservation option" assert_file_contains usr/libexec/lite-nas/scripts/deploy/lite-nas.sh "\"\$LITE_NAS_REPO_ROOT/scripts/deploy-configs.sh\" \"\${config_deploy_args[@]}\""
assert_cmd "package postinstall deploys file writer" assert_file_contains usr/libexec/lite-nas/scripts/runtime/deploy-package-runtime.sh "deploy.fileWriter.deploy \"\$PACKAGE_ROOT/lite-nas-file-writer\" \"\$load_writer_profile\""
assert_cmd "package postinstall deploys file service" assert_file_contains usr/libexec/lite-nas/scripts/runtime/deploy-package-runtime.sh "deploy.fileService.deploy \"\$PACKAGE_ROOT/file-service\" 0"
assert_cmd "package postinstall deploys stream service" assert_file_contains usr/libexec/lite-nas/scripts/runtime/deploy-package-runtime.sh "deploy.streamService.deploy \"\$PACKAGE_ROOT/stream-service\" 0"
assert_cmd "NATS preservation skips only NATS config entries" assert_file_contains usr/libexec/lite-nas/scripts/deploy-configs.sh "[ \"\$config_name\" = \"nats-server.conf\" ]"
assert_cmd "resources-monitor deploy script packaged" assert_executable usr/libexec/lite-nas/scripts/deploy/resources-monitor.sh
assert_cmd "etc permission normalizer packaged" assert_executable usr/libexec/lite-nas/scripts/deploy/normalize-etc-permissions.sh
assert_cmd "package version config packaged" assert_file usr/libexec/lite-nas/scripts/config/version.conf

assert_cmd "system-logging-manager-cli symlink packaged" assert_symlink_target usr/bin/system-logging-manager-cli /usr/libexec/lite-nas/system-logging-manager-cli
assert_cmd "security-logging-manager-cli symlink packaged" assert_symlink_target usr/bin/security-logging-manager-cli /usr/libexec/lite-nas/security-logging-manager-cli
assert_cmd "system-metrics-cli symlink packaged" assert_symlink_target usr/bin/system-metrics-cli /usr/libexec/lite-nas/system-metrics-cli
assert_cmd "network-metrics-cli symlink packaged" assert_symlink_target usr/bin/network-metrics-cli /usr/libexec/lite-nas/network-metrics-cli
assert_cmd "zfs-metrics-cli symlink packaged" assert_symlink_target usr/bin/zfs-metrics-cli /usr/libexec/lite-nas/zfs-metrics-cli
assert_cmd "file-writer symlink packaged" assert_symlink_target usr/bin/lite-nas-file-writer /usr/libexec/lite-nas/lite-nas-file-writer

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
assert_cmd "network-metrics.conf packaged" assert_file "$packaged_config_root/lite-nas/network-metrics.conf"
assert_cmd "process-metrics.conf packaged" assert_file "$packaged_config_root/lite-nas/process-metrics.conf"
assert_cmd "disk-metrics.conf packaged" assert_file "$packaged_config_root/lite-nas/disk-metrics.conf"
assert_cmd "service-metrics.conf packaged" assert_file "$packaged_config_root/lite-nas/service-metrics.conf"
assert_cmd "network-metrics-cli.conf packaged" assert_file "$packaged_config_root/lite-nas/network-metrics-cli.conf"
assert_cmd "resources-monitor.conf packaged" assert_file "$packaged_config_root/lite-nas/resources-monitor.conf"
assert_cmd "litenas-integrity.conf packaged" assert_file "$packaged_config_root/lite-nas/litenas-integrity.conf"
assert_cmd "zfs-metrics-cli.conf packaged" assert_file "$packaged_config_root/lite-nas/zfs-metrics-cli.conf"
assert_cmd "system metrics rule packaged" assert_file "$packaged_config_root/lite-nas/resources-monitor/rules/system-metrics.json"
assert_cmd "network metrics rule packaged" assert_file "$packaged_config_root/lite-nas/resources-monitor/rules/network-metrics.json"
assert_cmd "disk metrics rule packaged" assert_file "$packaged_config_root/lite-nas/resources-monitor/rules/disk-metrics.json"
assert_cmd "zfs metrics rule packaged" assert_file "$packaged_config_root/lite-nas/resources-monitor/rules/zfs-metrics.json"
assert_cmd "service metrics rule packaged" assert_file "$packaged_config_root/lite-nas/resources-monitor/rules/service-metrics.json"
assert_cmd "web-gateway.conf packaged" assert_file "$packaged_config_root/lite-nas/web-gateway.conf"
assert_cmd "moments.conf packaged" assert_file "$packaged_config_root/lite-nas/moments.conf"
assert_cmd "Moments Nginx host packaged" assert_file "$packaged_config_root/nginx/sites-available/lite-nas-moments.conf"
assert_cmd "Moments embedded Nginx mount packaged" assert_file "$packaged_config_root/nginx/snippets/lite-nas-moments-embedded.conf"
assert_cmd "admin Nginx site includes Moments mount" assert_file_contains "$packaged_config_root/nginx/sites-available/lite-nas-web-gateway.conf" 'include /etc/nginx/snippets/lite-nas-moments-embedded.conf;'
assert_cmd "Moments AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.moments"
assert_cmd "Moments unit packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-moments.service"
assert_cmd "file-service.conf packaged" assert_file "$packaged_config_root/lite-nas/file-service.conf"
assert_cmd "stream-service.conf packaged" assert_file "$packaged_config_root/lite-nas/stream-service.conf"
assert_cmd "stream-service uses dedicated NATS certificate" assert_file_contains "$packaged_config_root/lite-nas/stream-service.conf" 'lite-nas-stream-service/client.crt'
assert_cmd "stream-service listens on loopback" assert_file_contains "$packaged_config_root/lite-nas/stream-service.conf" 'address=127.0.0.1:8083'
assert_cmd "stream-service NATS identity is authorized" assert_file_contains "$packaged_config_root/nats-server/authorization.conf" 'user: "lite-nas-stream-service"'

assert_cmd "resources-monitor unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-resources-monitor.service"
assert_cmd "system logging manager unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-system-logging-manager.service"
assert_cmd "security logging manager unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-security-logging-manager.service"
assert_cmd "system email notifier unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-system-email-notifier.service"
assert_cmd "security email notifier unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-security-email-notifier.service"
assert_cmd "rbac unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-rbac.service"
assert_cmd "rbac unit can inspect home ACLs read-only" assert_file_contains "$packaged_config_root/systemd/system/lite-nas-rbac.service" 'ProtectHome=read-only'
assert_cmd "rbac unit joins storage read group" assert_file_contains "$packaged_config_root/systemd/system/lite-nas-rbac.service" 'SupplementaryGroups=lite-nas lite-nas-fs-rd'
assert_cmd "network-metrics unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-network-metrics.service"
assert_cmd "disk-metrics unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-disk-metrics.service"
assert_cmd "service-metrics unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-service-metrics.service"
assert_cmd "process-metrics unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-process-metrics.service"
assert_cmd "zfs-metrics unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-zfs-metrics.service"
assert_cmd "litenas-integrity unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-integrity.service"
assert_cmd "file-service unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-file-service.service"
assert_cmd "stream-service unit template packaged" assert_file "$packaged_config_root/systemd/system/lite-nas-stream-service.service"
assert_cmd "stream-service unit runs as dedicated user" assert_file_contains "$packaged_config_root/systemd/system/lite-nas-stream-service.service" 'User=lite-nas-stream-service'
assert_cmd "stream-service unit has no file write grant" assert_file_contains "$packaged_config_root/systemd/system/lite-nas-stream-service.service" 'ReadWritePaths=/var/log/lite-nas/stream-service.log'
assert_cmd "stream-service preview uses bounded private tmpfs" assert_file_contains "$packaged_config_root/systemd/system/lite-nas-stream-service.service" 'TemporaryFileSystem=/tmp:rw,size=64M,mode=1777'

assert_cmd "postfix main.cf template packaged" assert_file "$packaged_config_root/postfix/main.cf"
assert_cmd "postfix master.cf template packaged" assert_file "$packaged_config_root/postfix/master.cf"
assert_cmd "postfix AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.lib.postfix.sbin"
assert_cmd "auth AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.auth-service"
assert_cmd "rbac AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.rbac-service"
assert_cmd "system-metrics AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-metrics"
assert_cmd "zfs-metrics AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.zfs-metrics"
assert_cmd "resources-monitor AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.resources-monitor"
assert_cmd "litenas-integrity AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.litenas-integrity"
assert_cmd "system-logging-manager AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-logging-manager"
assert_cmd "security-logging-manager AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.security-logging-manager"
assert_cmd "system-email-notifier AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-email-notifier"
assert_cmd "security-email-notifier AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.security-email-notifier"
assert_cmd "network-metrics AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.network-metrics"
assert_cmd "process-metrics AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.process-metrics"
assert_cmd "disk-metrics AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.disk-metrics"
assert_cmd "service-metrics AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.service-metrics"
assert_cmd "web-gateway AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.web-gateway"
assert_cmd "system-metrics-cli AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-metrics-cli"
assert_cmd "network-metrics-cli AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.network-metrics-cli"
assert_cmd "zfs-metrics-cli AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.zfs-metrics-cli"
assert_cmd "file-writer AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.lite-nas-file-writer"
assert_cmd "file-service AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.file-service"
assert_cmd "stream-service AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.stream-service"
assert_cmd "stream-service profile confines installed binary" assert_file_contains "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.stream-service" 'profile /usr/libexec/lite-nas/stream-service'
assert_cmd "stream-service profile permits HEIC decoder" assert_file_contains "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.stream-service" '/usr/bin/heif-dec ix,'
assert_cmd "file-service AppArmor permits sudo UID changes" assert_file_contains "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.file-service" 'capability setuid,'
assert_cmd "file-service AppArmor permits sudo GID changes" assert_file_contains "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.file-service" 'capability setgid,'
assert_cmd "file-service AppArmor permits sudo PAM reads" assert_file_contains "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.file-service" 'include <abstractions/authentication>'
assert_cmd "system-logging-manager-cli AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.system-logging-manager-cli"
assert_cmd "security-logging-manager-cli AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.libexec.lite-nas.security-logging-manager-cli"
assert_cmd "nginx AppArmor profile packaged" assert_file "$packaged_config_root/apparmor.d/usr.sbin.nginx"

assert_cmd "PAM config template packaged" assert_file "$packaged_config_root/pam.d/litenas-auth"
assert_cmd "nginx site config template packaged" assert_file "$packaged_config_root/nginx/sites-available/lite-nas-web-gateway.conf"
assert_cmd "file-service nginx snippet packaged" assert_file "$packaged_config_root/nginx/snippets/lite-nas-file-service.conf"
assert_cmd "stream-service nginx snippet packaged" assert_file "$packaged_config_root/nginx/snippets/lite-nas-stream-service.conf"
assert_cmd "admin Nginx site includes stream mount" assert_file_contains "$packaged_config_root/nginx/sites-available/lite-nas-web-gateway.conf" 'include /etc/nginx/snippets/lite-nas-stream-service.conf;'
assert_cmd "Moments Nginx site includes stream mount" assert_file_contains "$packaged_config_root/nginx/sites-available/lite-nas-moments.conf" 'include /etc/nginx/snippets/lite-nas-stream-service.conf;'
assert_cmd "file-service nginx snippet forwards browser cookies" assert_file_contains "$packaged_config_root/nginx/snippets/lite-nas-file-service.conf" "proxy_set_header Cookie \$http_cookie;"
assert_cmd "nginx site includes file-service snippet" assert_file_contains "$packaged_config_root/nginx/sites-available/lite-nas-web-gateway.conf" 'include /etc/nginx/snippets/lite-nas-file-service.conf;'
assert_cmd "ufw default config template packaged" assert_file "$packaged_config_root/default/ufw"
assert_cmd "ufw config template packaged" assert_file "$packaged_config_root/ufw/ufw.conf"
assert_cmd "service admin sudoers template packaged" assert_file "$packaged_config_root/sudoers.d/lite-nas-svc-adm"
assert_cmd "service admin sudoers rule uses kill helper" assert_file_contains "$packaged_config_root/sudoers.d/lite-nas-svc-adm" $'\t/usr/libexec/lite-nas/lite-nas-kill-pid'
assert_cmd "auth unit permits PAM helper transition" assert_file_contains "$packaged_config_root/systemd/system/lite-nas-auth.service" "NoNewPrivileges=no"
assert_cmd "auth unit grants PAM credential read capability" assert_file_contains "$packaged_config_root/systemd/system/lite-nas-auth.service" "CapabilityBoundingSet=CAP_DAC_READ_SEARCH"
assert_cmd "auth unit makes PAM credential read capability effective" assert_file_contains "$packaged_config_root/systemd/system/lite-nas-auth.service" "AmbientCapabilities=CAP_DAC_READ_SEARCH"
assert_cmd "file-service unit bounds sudo identity capabilities" assert_file_contains "$packaged_config_root/systemd/system/lite-nas-file-service.service" 'CapabilityBoundingSet=CAP_SETUID CAP_SETGID'
assert_cmd "zfs admin sudoers template packaged" assert_file "$packaged_config_root/sudoers.d/lite-nas-zfs-adm"
assert_cmd "rbac sudoers template packaged" assert_file "$packaged_config_root/sudoers.d/lite-nas-rbac"
assert_cmd "litenas-integrity sudoers template packaged" assert_file "$packaged_config_root/sudoers.d/lite-nas-integrity"
assert_cmd "file-writer sudoers template packaged" assert_file "$packaged_config_root/sudoers.d/lite-nas-file-writer"
assert_cmd "file-service sudoers template packaged" assert_file "$packaged_config_root/sudoers.d/lite-nas-file-service-commands"
assert_cmd "file-writer sudoers excludes root" assert_file_contains "$packaged_config_root/sudoers.d/lite-nas-file-writer" 'ALL=(ALL,!root,!#0) NOPASSWD: /usr/libexec/lite-nas/lite-nas-file-writer *'

assert_cmd "web-gateway index.html packaged" assert_file usr/share/lite-nas/web-gateway/assets/index.html
assert_cmd "web-gateway index.css packaged" assert_file usr/share/lite-nas/web-gateway/assets/index.css
assert_cmd "web-gateway index.js packaged" assert_file usr/share/lite-nas/web-gateway/assets/index.js
log.popTask

log.info "Package content checks passed: $checks_passed"
log.info "Validated static package contents: $package_path"
