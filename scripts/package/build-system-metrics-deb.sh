#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

package_name="lite-nas-system-metrics"
package_version="${LITE_NAS_PACKAGE_VERSION:-0.1.1}"
binary_path=""
output_dir="$LITE_NAS_REPO_ROOT/.build/packages"
package_template_dir="$LITE_NAS_REPO_ROOT/packaging/debian/$package_name"
package_root=""

usage() {
	cat <<'MSG'
Usage: scripts/package/build-system-metrics-deb.sh [options]

Options:
  --version=VERSION   Debian package version. Defaults to LITE_NAS_PACKAGE_VERSION or 0.1.0.
  --binary PATH       Use an existing system-metrics binary.
  --output-dir PATH   Directory where the package and build root will be written.
  -h, --help          Show this help.
MSG
}

args.parse "$@"
if ! args.assertKnown version binary output-dir help h; then
	log.error "Unknown option: --$(args.unknownKeys version binary output-dir help h | head -n 1)"
	usage >&2
	exit 64
fi
if args.has h || args.has help; then
	usage
	exit 0
fi
if args.has version && ! package_version="$(args.require_arg version)"; then
	log.error "Missing value for --version"
	usage >&2
	exit 64
fi
if args.has binary && ! binary_path="$(args.require_arg binary)"; then
	log.error "Missing value for --binary"
	usage >&2
	exit 64
fi
if args.has output-dir && ! output_dir="$(args.require_arg output-dir)"; then
	log.error "Missing value for --output-dir"
	usage >&2
	exit 64
fi

package.requireBuildCommands
package_arch="$(package.resolveArchitecture)"

if [ -z "$binary_path" ]; then
	binary_path="$output_dir/${package_name}-${package_arch}/system-metrics"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-metrics-binary.sh" \
		"--output=${binary_path}"
fi

if [ ! -f "$binary_path" ]; then
	log.error "Missing system-metrics binary: $binary_path"
	exit 1
fi

package_root="$output_dir/${package_name}_${package_version}_${package_arch}.root"
deb_path="$output_dir/${package_name}_${package_version}_${package_arch}.deb"

log.pushTask "Preparing Debian package root for ${package_name} ${package_version} (${package_arch})"
package.prepareMetadata \
	"$package_arch" \
	"$package_version" \
	"$package_template_dir" \
	"$package_root"
for maintainer_script in postinst prerm postrm config; do
	package.installMaintainerScript "$package_template_dir" "$package_root" "$maintainer_script"
done

package.copyDocTreeAndCompressChangelog \
	"$package_template_dir/usr/share/doc" \
	"$package_root/usr/share/doc" \
	"$package_root/usr/share/doc/$package_name/changelog.Debian"

install -D -m 0755 "$binary_path" "$package_root/usr/libexec/lite-nas/system-metrics"
package.copyTree "$LITE_NAS_REPO_ROOT/configs" "$package_root/usr/libexec/lite-nas/configs"
package.copyTree "$LITE_NAS_REPO_ROOT/scripts/helpers" "$package_root/usr/libexec/lite-nas/scripts/helpers"
package.copyTree "$LITE_NAS_REPO_ROOT/scripts/deploy" "$package_root/usr/libexec/lite-nas/scripts/deploy"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-configs.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-configs.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-system-metrics.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-nats-certificates.sh"

find "$package_root/usr" -type d -exec chmod 0755 {} +
find "$package_root/usr" -type f -exec chmod 0644 {} +
chmod 0755 \
	"$package_root/usr/libexec/lite-nas/system-metrics" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-configs.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-nats-certificates.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/helpers/go-modules.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/helpers/tool-paths.sh"

package.writeMd5sums "$package_root"
log.popTask

log.pushTask "Building Debian package ${deb_path}"
dpkg-deb --root-owner-group --build "$package_root" "$deb_path"
log.popTask

log.info "Built Debian package: $deb_path"
