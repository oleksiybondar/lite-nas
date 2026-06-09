#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

if [ "$#" -ne 2 ]; then
	log.error "Usage: scripts/ci/validate-lite-nas-deb-install.sh <package.deb> <amd64|arm64>"
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

log.requireCommand "docker" "Install Docker and retry."

package_dir="$(dirname "$package_path")"
package_name="$(basename "$package_path")"

log.pushTask "Validating LiteNAS package installability for ${target_arch}"
docker run --rm \
	--platform "linux/${target_arch}" \
	-e DEBIAN_FRONTEND=noninteractive \
	-e LITE_NAS_PACKAGE_INSTALL_MODE=validate \
	-v "${package_dir}:/packages:ro" \
	ubuntu:noble \
	bash -lc "
		set -euo pipefail
		apt-get update
		apt-get install --no-install-recommends -y /packages/${package_name}
		dpkg -s lite-nas >/dev/null
		dpkg -s postfix >/dev/null
		dpkg -s sudo >/dev/null
		dpkg -s aide >/dev/null
		test -x /usr/libexec/lite-nas/auth-service
		test -x /usr/libexec/lite-nas/rbac-service
		test -x /usr/libexec/lite-nas/system-logging-manager
		test -x /usr/libexec/lite-nas/security-logging-manager
		test -x /usr/libexec/lite-nas/system-email-notifier
		test -x /usr/libexec/lite-nas/security-email-notifier
		test -x /usr/libexec/lite-nas/system-metrics
		test -x /usr/libexec/lite-nas/resources-monitor
		test -x /usr/libexec/lite-nas/system-logging-manager-cli
		test -x /usr/libexec/lite-nas/security-logging-manager-cli
		test -x /usr/libexec/lite-nas/system-metrics-cli
		test -L /usr/bin/system-logging-manager-cli
		test -x /usr/bin/system-logging-manager-cli
		test -L /usr/bin/security-logging-manager-cli
		test -x /usr/bin/security-logging-manager-cli
		test -L /usr/bin/system-metrics-cli
		test -x /usr/bin/system-metrics-cli
		test -x /usr/libexec/lite-nas/web-gateway
		test \"\$(stat -c '%U:%G %a' /etc/lite-nas)\" = 'root:lite-nas 711'
		test -f /etc/lite-nas/auth.conf
		test -f /etc/lite-nas/rbac-service.conf
		test -f /etc/lite-nas/system-logging-manager.conf
		test -f /etc/lite-nas/security-logging-manager.conf
		test -f /etc/lite-nas/system-metrics.conf
		test -f /etc/lite-nas/system-logging-manager-cli.conf
		test -f /etc/lite-nas/security-logging-manager-cli.conf
		test -f /etc/lite-nas/system-email-notifier.conf
		test -f /etc/lite-nas/security-email-notifier.conf
		test -f /etc/lite-nas/resources-monitor.conf
		test -f /etc/lite-nas/resources-monitor/rules/system-metrics.json
		test -f /etc/lite-nas/resources-monitor/rules/zfs-metrics.json
		id lite-nas-resources-monitor >/dev/null
		id lite-nas-sys-log-mgr >/dev/null
		id lite-nas-sec-log-mgr >/dev/null
		id lite-nas-sys-email-notifier >/dev/null
		id lite-nas-sec-email-notifier >/dev/null
		test -f /etc/systemd/system/lite-nas-resources-monitor.service
		test -f /etc/systemd/system/lite-nas-system-logging-manager.service
		test -f /etc/systemd/system/lite-nas-security-logging-manager.service
		test -f /etc/systemd/system/lite-nas-system-email-notifier.service
		test -f /etc/systemd/system/lite-nas-security-email-notifier.service
		test -f /etc/systemd/system/lite-nas-rbac.service
		test \"\$(stat -c '%U:%G %a' /etc/lite-nas/system-metrics-cli.conf)\" = 'lite-nas-system-metrics-cli:users 640'
		test -f /etc/lite-nas/web-gateway.conf
		test -f /etc/postfix/main.cf
		test -f /etc/postfix/master.cf
		test -f /etc/apparmor.d/usr.lib.postfix.sbin
		test -f /etc/apparmor.d/usr.libexec.lite-nas.auth-service
		test -f /etc/apparmor.d/usr.libexec.lite-nas.rbac-service
		test -f /etc/apparmor.d/usr.libexec.lite-nas.system-metrics
		test -f /etc/apparmor.d/usr.libexec.lite-nas.resources-monitor
		test -f /etc/apparmor.d/usr.libexec.lite-nas.system-logging-manager
		test -f /etc/apparmor.d/usr.libexec.lite-nas.security-logging-manager
		test -f /etc/apparmor.d/usr.libexec.lite-nas.system-email-notifier
		test -f /etc/apparmor.d/usr.libexec.lite-nas.security-email-notifier
		test -f /etc/apparmor.d/usr.libexec.lite-nas.web-gateway
		test -f /etc/apparmor.d/usr.libexec.lite-nas.system-metrics-cli
		test -f /etc/apparmor.d/usr.libexec.lite-nas.system-logging-manager-cli
		test -f /etc/apparmor.d/usr.libexec.lite-nas.security-logging-manager-cli
		test -f /etc/apparmor.d/usr.sbin.nginx
		test -f /etc/pam.d/litenas-auth
		test -f /etc/lite-nas/certificates/auth/token-signing.key
		test -f /etc/lite-nas/certificates/auth/token-signing.crt
		test -f /etc/lite-nas/certificates/identities/root-ca.key
		test -f /etc/lite-nas/certificates/identities/root-ca.crt
		test -f /etc/lite-nas/certificates/identities/lite-nas-auth-service/client.crt
		test -f /etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.crt
		test -f /etc/lite-nas/certificates/identities/lite-nas-sys-log-mgr-cli/client.crt
		test -f /etc/lite-nas/certificates/identities/lite-nas-sec-log-mgr-cli/client.crt
		test -f /etc/lite-nas/certificates/transport/lite-nas-auth-service/client.crt
		test -f /etc/lite-nas/certificates/transport/lite-nas-sys-email-notifier/client.crt
		test -f /etc/lite-nas/certificates/transport/lite-nas-sec-email-notifier/client.crt
		test -f /etc/lite-nas/certificates/transport/lite-nas-rbac-service/client.crt
		test -f /etc/lite-nas/certificates/transport/lite-nas-rbac-service/client.key
		test -d /var/log/lite-nas
		test \"\$(stat -c '%U:%G %a' /var/log/lite-nas)\" = 'root:lite-nas 751'
		test -f /var/log/lite-nas/auth-service.log
		test -f /var/log/lite-nas/system-metrics.log
		test -f /var/log/lite-nas/system-metrics-cli.log
		test \"\$(stat -c '%U:%G %a' /var/log/lite-nas/system-metrics-cli.log)\" = 'root:root 666'
		test -f /var/log/lite-nas/web-gateway.log
		test -f /etc/nginx/sites-available/lite-nas-web-gateway.conf
		test -f /etc/default/ufw
		test -f /etc/ufw/ufw.conf
		test -f /usr/share/lite-nas/web-gateway/assets/index.html
		test -f /usr/share/lite-nas/web-gateway/assets/index.css
		test -f /usr/share/lite-nas/web-gateway/assets/index.js
	"
log.popTask

log.info "Validated package installability: $package_path"
