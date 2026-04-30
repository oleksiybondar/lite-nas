#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

package_name="lite-nas"
package_version="${LITE_NAS_PACKAGE_VERSION:-0.1.1}"
auth_service_binary_path=""
system_metrics_binary_path=""
system_metrics_cli_binary_path=""
web_gateway_binary_path=""
output_dir="$LITE_NAS_REPO_ROOT/.build/packages"
package_template_dir="$LITE_NAS_REPO_ROOT/packaging/debian/$package_name"
package_root=""

usage() {
	cat <<'MSG'
Usage: scripts/package/build-lite-nas-deb.sh [options]

Options:
  --version=VERSION                  Debian package version. Defaults to LITE_NAS_PACKAGE_VERSION or 0.1.0.
  --auth-service-binary=PATH         Use an existing auth-service binary.
  --system-metrics-binary=PATH       Use an existing system-metrics binary.
  --system-metrics-cli-binary=PATH   Use an existing system-metrics-cli binary.
  --web-gateway-binary=PATH          Use an existing web-gateway binary.
  --output-dir=PATH                  Directory where the package and build root will be written.
  -h, --help                         Show this help.
MSG
}

args.parse "$@"
if ! args.assertKnown version auth-service-binary system-metrics-binary system-metrics-cli-binary web-gateway-binary output-dir help h; then
	log.error "Unknown option: --$(args.unknownKeys version auth-service-binary system-metrics-binary system-metrics-cli-binary web-gateway-binary output-dir help h | head -n 1)"
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
if args.has auth-service-binary && ! auth_service_binary_path="$(args.require_arg auth-service-binary)"; then
	log.error "Missing value for --auth-service-binary"
	usage >&2
	exit 64
fi
if args.has system-metrics-binary && ! system_metrics_binary_path="$(args.require_arg system-metrics-binary)"; then
	log.error "Missing value for --system-metrics-binary"
	usage >&2
	exit 64
fi
if args.has system-metrics-cli-binary && ! system_metrics_cli_binary_path="$(args.require_arg system-metrics-cli-binary)"; then
	log.error "Missing value for --system-metrics-cli-binary"
	usage >&2
	exit 64
fi
if args.has web-gateway-binary && ! web_gateway_binary_path="$(args.require_arg web-gateway-binary)"; then
	log.error "Missing value for --web-gateway-binary"
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

if [ -z "$auth_service_binary_path" ]; then
	auth_service_binary_path="$output_dir/${package_name}-${package_arch}/auth-service"
	"$LITE_NAS_REPO_ROOT/scripts/build-auth-service-binary.sh" \
		"--output=${auth_service_binary_path}"
fi

if [ -z "$system_metrics_binary_path" ]; then
	system_metrics_binary_path="$output_dir/${package_name}-${package_arch}/system-metrics"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-metrics-binary.sh" \
		"--output=${system_metrics_binary_path}"
fi

if [ -z "$system_metrics_cli_binary_path" ]; then
	system_metrics_cli_binary_path="$output_dir/${package_name}-${package_arch}/system-metrics-cli"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-metrics-cli-binary.sh" \
		"--output=${system_metrics_cli_binary_path}"
fi

if [ -z "$web_gateway_binary_path" ]; then
	web_gateway_binary_path="$output_dir/${package_name}-${package_arch}/web-gateway"
	"$LITE_NAS_REPO_ROOT/scripts/build-web-gateway-binary.sh" \
		"--output=${web_gateway_binary_path}"
fi

if [ ! -f "$auth_service_binary_path" ]; then
	log.error "Missing auth-service binary: $auth_service_binary_path"
	exit 1
fi

if [ ! -f "$system_metrics_binary_path" ]; then
	log.error "Missing system-metrics binary: $system_metrics_binary_path"
	exit 1
fi

if [ ! -f "$system_metrics_cli_binary_path" ]; then
	log.error "Missing system-metrics-cli binary: $system_metrics_cli_binary_path"
	exit 1
fi

if [ ! -f "$web_gateway_binary_path" ]; then
	log.error "Missing web-gateway binary: $web_gateway_binary_path"
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
cp "$package_template_dir/DEBIAN/templates" "$package_root/DEBIAN/templates"
for maintainer_script in config postinst prerm postrm; do
	package.installMaintainerScript "$package_template_dir" "$package_root" "$maintainer_script"
done

package.copyTree "$package_template_dir/usr/share" "$package_root/usr/share"
gzip -n -9 "$package_root/usr/share/doc/$package_name/changelog.Debian"
mv "$package_root/usr/share/doc/$package_name/changelog.Debian.gz" \
	"$package_root/usr/share/doc/$package_name/changelog.gz"

package.copyTree "$LITE_NAS_REPO_ROOT/configs" "$package_root/usr/libexec/lite-nas/configs"
package.copyTree "$LITE_NAS_REPO_ROOT/scripts/helpers" "$package_root/usr/libexec/lite-nas/scripts/helpers"
package.copyTree "$LITE_NAS_REPO_ROOT/scripts/deploy" "$package_root/usr/libexec/lite-nas/scripts/deploy"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-configs.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-configs.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-lite-nas.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-lite-nas.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-auth-service.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-auth-service.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-system-metrics.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-system-metrics-cli.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics-cli.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-web-gateway.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-web-gateway.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/install-runtime-dependencies.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/install-runtime-dependencies.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-nats-certificates.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/rotate-nginx-certificates.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-nginx-certificates.sh"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/rotate-auth-token-certificates.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-auth-token-certificates.sh"
install -D -m 0755 "$auth_service_binary_path" \
	"$package_root/usr/libexec/lite-nas/auth-service"
install -D -m 0755 "$system_metrics_binary_path" \
	"$package_root/usr/libexec/lite-nas/system-metrics"
install -D -m 0755 "$system_metrics_cli_binary_path" \
	"$package_root/usr/libexec/lite-nas/system-metrics-cli"
install -D -m 0755 "$web_gateway_binary_path" \
	"$package_root/usr/libexec/lite-nas/web-gateway"
package.copyTree "$LITE_NAS_REPO_ROOT/services/web-gateway/assets" \
	"$package_root/usr/libexec/lite-nas/services/web-gateway/assets"

find "$package_root/usr/libexec/lite-nas/scripts" -type f -name "*.sh" -exec chmod 0755 {} +

find "$package_root/usr" -type d -exec chmod 0755 {} +
find "$package_root/usr" -type f -exec chmod 0644 {} +
chmod 0755 \
	"$package_root/usr/libexec/lite-nas/auth-service" \
	"$package_root/usr/libexec/lite-nas/system-metrics" \
	"$package_root/usr/libexec/lite-nas/system-metrics-cli" \
	"$package_root/usr/libexec/lite-nas/web-gateway" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-auth-service.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-configs.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-lite-nas.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-system-metrics-cli.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-web-gateway.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/install-runtime-dependencies.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-auth-token-certificates.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-nats-certificates.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/rotate-nginx-certificates.sh"
find "$package_root/usr/libexec/lite-nas/scripts" -type f -name "*.sh" -exec chmod 0755 {} +

package.writeMd5sums "$package_root"
log.popTask

log.pushTask "Building Debian package ${deb_path}"
dpkg-deb --root-owner-group --build "$package_root" "$deb_path"
log.popTask

log.info "Built Debian package: $deb_path"
