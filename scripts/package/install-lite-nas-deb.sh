#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

package_path=""
install_recommends=1

usage() {
	cat <<'MSG'
Usage: scripts/package/install-lite-nas-deb.sh [options]

Options:
  --package PATH             Path to the lite-nas .deb to install. Defaults to the newest .build/packages/lite-nas_*.deb.
  --no-install-recommends    Install only hard package dependencies.
  -h, --help                 Show this help.
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
	--no-install-recommends)
		install_recommends=0
		shift
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
log.requireCommand "dpkg-deb" "Install Debian packaging tools and retry."

if [ -z "$package_path" ]; then
	package_path="$(find "$LITE_NAS_REPO_ROOT/.build/packages" -maxdepth 1 -type f -name 'lite-nas_*.deb' | LC_ALL=C sort | tail -n1)"
fi

if [ -z "$package_path" ] || [ ! -f "$package_path" ]; then
	log.error "LiteNAS package not found: ${package_path:-<none>}"
	exit 1
fi

package_path="$(realpath "$package_path")"
package_version="$(dpkg-deb -f "$package_path" Version)"
file_explorer_package="$(dirname "$package_path")/lite-nas-file-explorer_${package_version}_all.deb"
if [ ! -f "$file_explorer_package" ]; then
	log.error "Matching File Explorer package not found: $file_explorer_package"
	exit 1
fi

log.pushTask "Installing LiteNAS and File Explorer packages with apt dependency resolution"
apt_args=()
if [ "$install_recommends" -eq 0 ]; then
	apt_args+=(--no-install-recommends)
fi
apt-get update
apt-get install "${apt_args[@]}" -y "$package_path" "$file_explorer_package"
log.popTask

log.info "Installed package: $package_path"
log.info "Installed package: $file_explorer_package"
