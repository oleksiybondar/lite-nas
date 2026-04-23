#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

package_path=""

usage() {
	cat <<'MSG'
Usage: scripts/package/install-lite-nas-deb.sh [options]

Options:
  --package PATH  Path to the lite-nas .deb to install. Defaults to the newest .build/packages/lite-nas_*.deb.
  -h, --help      Show this help.
MSG
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--package)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --package"
			usage >&2
			exit 2
		fi
		package_path="$2"
		shift 2
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		usage >&2
		exit 2
		;;
	esac
done

sudo.guard.requireRoot "scripts/package/install-lite-nas-deb.sh"
log.requireCommand "apt-get" "Install apt-get and retry."

if [ -z "$package_path" ]; then
	package_path="$(find "$LITE_NAS_REPO_ROOT/.build/packages" -maxdepth 1 -type f -name 'lite-nas_*.deb' | LC_ALL=C sort | tail -n1)"
fi

if [ -z "$package_path" ] || [ ! -f "$package_path" ]; then
	log.error "LiteNAS package not found: ${package_path:-<none>}"
	exit 1
fi

package_path="$(realpath "$package_path")"

log.pushTask "Installing LiteNAS package with apt dependency resolution"
apt-get update
apt-get install -y "$package_path"
log.popTask

log.info "Installed package: $package_path"
