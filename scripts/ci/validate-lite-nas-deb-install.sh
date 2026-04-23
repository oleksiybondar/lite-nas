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
		apt-get install -y /packages/${package_name}
		dpkg -s lite-nas >/dev/null
		test -x /usr/libexec/lite-nas/system-metrics
		test -x /usr/libexec/lite-nas/system-metrics-cli
		test -f /etc/liteNAS/system-metrics.conf
		test -f /etc/liteNAS/system-metrics-cli.conf
	"
log.popTask

log.info "Validated package installability: $package_path"
