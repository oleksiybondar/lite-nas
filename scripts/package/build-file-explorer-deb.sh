#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

package_version="${LITE_NAS_PACKAGE_VERSION:-$(packageVersion.localAlpha)}"
output_dir="$LITE_NAS_REPO_ROOT/.build/packages"
assets_dir=""

for arg in "$@"; do
	case "$arg" in
	--version=*) package_version="${arg#*=}" ;;
	--output-dir=*) output_dir="${arg#*=}" ;;
	--assets-dir=*) assets_dir="${arg#*=}" ;;
	-h | --help)
		printf 'Usage: scripts/package/build-file-explorer-deb.sh [--version=VERSION] [--output-dir=PATH] [--assets-dir=PATH]\n'
		exit 0
		;;
	*)
		log.error "Unknown option: $arg"
		exit 64
		;;
	esac
done

if [ -z "$package_version" ] || [ -z "$output_dir" ]; then
	log.error "Version and output directory must be nonempty."
	exit 64
fi
if [ "${output_dir#/}" = "$output_dir" ]; then
	output_dir="$LITE_NAS_REPO_ROOT/$output_dir"
fi
if [ -n "$assets_dir" ] && [ "${assets_dir#/}" = "$assets_dir" ]; then
	assets_dir="$LITE_NAS_REPO_ROOT/$assets_dir"
fi

if [ -z "$assets_dir" ]; then
	assets_dir="$output_dir/file-explorer-assets"
	"$LITE_NAS_REPO_ROOT/scripts/build-file-explorer.sh" --output-dir="$assets_dir"
fi
if [ ! -f "$assets_dir/index.html" ] || [ ! -d "$assets_dir/assets" ]; then
	log.error "Missing File Explorer Vite assets in: $assets_dir"
	exit 1
fi

log.requireCommand dpkg-deb "Install Debian packaging tools and retry."
if ! dpkg --validate-version "$package_version"; then
	log.error "Invalid Debian package version: $package_version"
	exit 64
fi
package_root="$output_dir/lite-nas-file-explorer_${package_version}_all.root"
deb_path="$output_dir/lite-nas-file-explorer_${package_version}_all.deb"
template_dir="$LITE_NAS_REPO_ROOT/packaging/debian/lite-nas-file-explorer"

package.prepareMetadata all "$package_version" "$template_dir" "$package_root"
package.installMaintainerScript "$template_dir" "$package_root" postinst
package.installMaintainerScript "$template_dir" "$package_root" postrm
install -d -m 0755 "$package_root/usr/share/lite-nas/file-explorer"
cp -R "$assets_dir/." "$package_root/usr/share/lite-nas/file-explorer/"
install -D -m 0644 "$template_dir/nginx/file-explorer.conf" \
	"$package_root/usr/share/lite-nas/file-explorer/nginx/file-explorer.conf"
install -D -m 0644 "$template_dir/usr/share/doc/lite-nas-file-explorer/copyright" \
	"$package_root/usr/share/doc/lite-nas-file-explorer/copyright"
find "$package_root" -type d -exec chmod 0755 {} +
find "$package_root/usr" -type f -exec chmod 0644 {} +
package.writeMd5sums "$package_root"

log.pushTask "Building Debian package ${deb_path}"
dpkg-deb --root-owner-group --build "$package_root" "$deb_path"
log.popTask
log.info "Built Debian package: $deb_path"
