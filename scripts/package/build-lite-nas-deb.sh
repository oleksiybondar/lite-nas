#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

package_name="lite-nas"
package_arch=""
package_version="${LITE_NAS_PACKAGE_VERSION:-0.1.0}"
system_metrics_binary_path=""
system_metrics_cli_binary_path=""
output_dir="$LITE_NAS_REPO_ROOT/.build/packages"
package_template_dir="$LITE_NAS_REPO_ROOT/packaging/debian/$package_name"
package_root=""

usage() {
	cat <<'MSG'
Usage: scripts/package/build-lite-nas-deb.sh [options]

Options:
  --arch=amd64|arm64                 Debian package architecture to build.
  --version=VERSION                  Debian package version. Defaults to LITE_NAS_PACKAGE_VERSION or 0.1.0.
  --system-metrics-binary=PATH       Use an existing system-metrics binary.
  --system-metrics-cli-binary=PATH   Use an existing system-metrics-cli binary.
  --output-dir=PATH                  Directory where the package and build root will be written.
  -h, --help                         Show this help.
MSG
}

for arg in "$@"; do
	case "$arg" in
	--arch=amd64)
		package_arch="amd64"
		;;
	--arch=arm64)
		package_arch="arm64"
		;;
	--version=*)
		package_version="${arg#--version=}"
		;;
	--system-metrics-binary=*)
		system_metrics_binary_path="${arg#--system-metrics-binary=}"
		;;
	--system-metrics-cli-binary=*)
		system_metrics_cli_binary_path="${arg#--system-metrics-cli-binary=}"
		;;
	--output-dir=*)
		output_dir="${arg#--output-dir=}"
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		log.error "Unknown option: $arg"
		usage >&2
		exit 64
		;;
	esac
done

if [ -z "$package_arch" ]; then
	log.error "Missing required option: --arch=amd64|arm64"
	usage >&2
	exit 64
fi

log.requireCommand "dpkg-deb" "Install dpkg-deb and retry."
log.requireCommand "gzip" "Install gzip and retry."

if [ -z "$system_metrics_binary_path" ]; then
	system_metrics_binary_path="$output_dir/${package_name}-${package_arch}/system-metrics"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-metrics-binary.sh" \
		"--arch=${package_arch}" \
		"--output=${system_metrics_binary_path}"
fi

if [ -z "$system_metrics_cli_binary_path" ]; then
	system_metrics_cli_binary_path="$output_dir/${package_name}-${package_arch}/system-metrics-cli"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-metrics-cli-binary.sh" \
		"--arch=${package_arch}" \
		"--output=${system_metrics_cli_binary_path}"
fi

if [ ! -f "$system_metrics_binary_path" ]; then
	log.error "Missing system-metrics binary: $system_metrics_binary_path"
	exit 1
fi

if [ ! -f "$system_metrics_cli_binary_path" ]; then
	log.error "Missing system-metrics-cli binary: $system_metrics_cli_binary_path"
	exit 1
fi

package_root="$output_dir/${package_name}_${package_version}_${package_arch}.root"
deb_path="$output_dir/${package_name}_${package_version}_${package_arch}.deb"

rm -rf "$package_root"
mkdir -p "$package_root/DEBIAN"

render_template() {
	local template_path="$1"
	local destination_path="$2"

	sed \
		-e "s|@PACKAGE_ARCH@|$package_arch|g" \
		-e "s|@PACKAGE_VERSION@|$package_version|g" \
		"$template_path" >"$destination_path"
}

copy_tree() {
	local source_dir="$1"
	local destination_dir="$2"

	mkdir -p "$destination_dir"
	cp -a "$source_dir/." "$destination_dir/"
}

log.pushTask "Preparing Debian package root for ${package_name} ${package_version} (${package_arch})"
render_template "$package_template_dir/DEBIAN/control.in" "$package_root/DEBIAN/control"
cp "$package_template_dir/DEBIAN/templates" "$package_root/DEBIAN/templates"
cp "$package_template_dir/DEBIAN/config" "$package_root/DEBIAN/config"
cp "$package_template_dir/DEBIAN/postinst" "$package_root/DEBIAN/postinst"
cp "$package_template_dir/DEBIAN/prerm" "$package_root/DEBIAN/prerm"
cp "$package_template_dir/DEBIAN/postrm" "$package_root/DEBIAN/postrm"
chmod 0755 \
	"$package_root/DEBIAN/config" \
	"$package_root/DEBIAN/postinst" \
	"$package_root/DEBIAN/prerm" \
	"$package_root/DEBIAN/postrm"

copy_tree "$package_template_dir/usr/share/doc" "$package_root/usr/share/doc"
gzip -n -9 "$package_root/usr/share/doc/$package_name/changelog.Debian"
mv "$package_root/usr/share/doc/$package_name/changelog.Debian.gz" \
	"$package_root/usr/share/doc/$package_name/changelog.gz"

copy_tree "$LITE_NAS_REPO_ROOT/configs" "$package_root/usr/libexec/lite-nas/configs"
copy_tree "$LITE_NAS_REPO_ROOT/scripts/helpers" "$package_root/usr/libexec/lite-nas/scripts/helpers"
copy_tree "$LITE_NAS_REPO_ROOT/scripts/deploy" "$package_root/usr/libexec/lite-nas/scripts/deploy"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-configs.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-configs.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-lite-nas.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-lite-nas.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-system-metrics.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-system-metrics-cli.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics-cli.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-nats-certificates.sh"
install -D -m 0755 "$system_metrics_binary_path" \
	"$package_root/usr/libexec/lite-nas/system-metrics"
install -D -m 0755 "$system_metrics_cli_binary_path" \
	"$package_root/usr/libexec/lite-nas/system-metrics-cli"

find "$package_root/usr/libexec/lite-nas/scripts" -type f -name "*.sh" -exec chmod 0755 {} +

find "$package_root/usr" -type d -exec chmod 0755 {} +
find "$package_root/usr" -type f -exec chmod 0644 {} +
chmod 0755 \
	"$package_root/usr/libexec/lite-nas/system-metrics" \
	"$package_root/usr/libexec/lite-nas/system-metrics-cli" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-configs.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-lite-nas.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics-cli.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-nats-certificates.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/helpers/go-modules.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/helpers/tool-paths.sh"

(
	cd "$package_root"
	while IFS= read -r -d '' file; do
		printf '%s  %s\n' \
			"$(md5sum "$file" | awk '{print $1}')" \
			"${file#./}"
	done < <(find . -path './DEBIAN' -prune -o -type f -print0 | LC_ALL=C sort -z) >DEBIAN/md5sums
)
log.popTask

log.pushTask "Building Debian package ${deb_path}"
dpkg-deb --root-owner-group --build "$package_root" "$deb_path"
log.popTask

log.info "Built Debian package: $deb_path"
