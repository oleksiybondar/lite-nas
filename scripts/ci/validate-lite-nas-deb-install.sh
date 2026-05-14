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
		dpkg -s aide >/dev/null
		test -x /usr/libexec/lite-nas/auth-service
		test -x /usr/libexec/lite-nas/system-metrics
		test -x /usr/libexec/lite-nas/system-metrics-cli
		test -L /usr/bin/system-metrics-cli
		test -x /usr/bin/system-metrics-cli
		test -x /usr/libexec/lite-nas/web-gateway
		test \"\$(stat -c '%U:%G %a' /etc/lite-nas)\" = 'root:lite-nas 711'
		test -f /etc/lite-nas/auth.conf
		test -f /etc/lite-nas/system-metrics.conf
		test -f /etc/lite-nas/resources-monitor.conf
		test -f /etc/lite-nas/resources-monitor/rules/system-metrics.json
		id lite-nas-resources-monitor >/dev/null
		test -f /etc/systemd/system/lite-nas-resources-monitor.service
		test -f /etc/lite-nas/system-metrics-cli.conf
		test \"\$(stat -c '%U:%G %a' /etc/lite-nas/system-metrics-cli.conf)\" = 'root:root 644'
		test -f /etc/lite-nas/web-gateway.conf
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
