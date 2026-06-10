#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/admin-panel-assets.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../config/version.conf"

cd "$LITE_NAS_REPO_ROOT"

package_name="lite-nas"
package_version="${LITE_NAS_PACKAGE_VERSION:-$LITE_NAS_BASE_VERSION}"
auth_service_binary_path=""
rbac_service_binary_path=""
system_logging_manager_binary_path=""
security_logging_manager_binary_path=""
system_email_notifier_binary_path=""
security_email_notifier_binary_path=""
system_metrics_binary_path=""
zfs_metrics_binary_path=""
system_logging_manager_cli_binary_path=""
security_logging_manager_cli_binary_path=""
system_metrics_cli_binary_path=""
zfs_metrics_cli_binary_path=""
web_gateway_binary_path=""
resources_monitor_binary_path=""
admin_panel_assets_path=""
output_dir="$LITE_NAS_REPO_ROOT/.build/packages"
package_template_dir="$LITE_NAS_REPO_ROOT/packaging/debian/$package_name"
package_root=""

usage() {
	cat <<'MSG'
Usage: scripts/package/build-lite-nas-deb.sh [options]

Options:
  --version=VERSION                  Debian package version. Defaults to LITE_NAS_PACKAGE_VERSION or LITE_NAS_BASE_VERSION from scripts/config/version.conf.
  --auth-service-binary=PATH         Use an existing auth-service binary.
  --rbac-service-binary=PATH         Use an existing rbac-service binary.
  --system-logging-manager-binary=PATH
                                     Use an existing system-logging-manager binary.
  --security-logging-manager-binary=PATH
                                     Use an existing security-logging-manager binary.
  --system-email-notifier-binary=PATH
                                     Use an existing system-email-notifier binary.
  --security-email-notifier-binary=PATH
                                     Use an existing security-email-notifier binary.
  --system-metrics-binary=PATH       Use an existing system-metrics binary.
  --zfs-metrics-binary=PATH          Use an existing zfs-metrics binary.
  --system-logging-manager-cli-binary=PATH
                                     Use an existing system-logging-manager-cli binary.
  --security-logging-manager-cli-binary=PATH
                                     Use an existing security-logging-manager-cli binary.
  --system-metrics-cli-binary=PATH   Use an existing system-metrics-cli binary.
  --zfs-metrics-cli-binary=PATH      Use an existing zfs-metrics-cli binary.
  --web-gateway-binary=PATH          Use an existing web-gateway binary.
  --resources-monitor-binary=PATH    Use an existing resources-monitor binary.
  --admin-panel-assets=PATH          Use an existing admin-panel Vite build output directory.
  --output-dir=PATH                  Directory where the package and build root will be written.
  -h, --help                         Show this help.
MSG
}

args.parse "$@"
if ! args.assertKnown version auth-service-binary rbac-service-binary system-logging-manager-binary security-logging-manager-binary system-email-notifier-binary security-email-notifier-binary system-metrics-binary zfs-metrics-binary system-logging-manager-cli-binary security-logging-manager-cli-binary system-metrics-cli-binary zfs-metrics-cli-binary web-gateway-binary resources-monitor-binary admin-panel-assets output-dir help h; then
	log.error "Unknown option: --$(args.unknownKeys version auth-service-binary rbac-service-binary system-logging-manager-binary security-logging-manager-binary system-email-notifier-binary security-email-notifier-binary system-metrics-binary zfs-metrics-binary system-logging-manager-cli-binary security-logging-manager-cli-binary system-metrics-cli-binary zfs-metrics-cli-binary web-gateway-binary resources-monitor-binary admin-panel-assets output-dir help h | head -n 1)"
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
if args.has rbac-service-binary && ! rbac_service_binary_path="$(args.require_arg rbac-service-binary)"; then
	log.error "Missing value for --rbac-service-binary"
	usage >&2
	exit 64
fi
if args.has system-logging-manager-binary && ! system_logging_manager_binary_path="$(args.require_arg system-logging-manager-binary)"; then
	log.error "Missing value for --system-logging-manager-binary"
	usage >&2
	exit 64
fi
if args.has security-logging-manager-binary && ! security_logging_manager_binary_path="$(args.require_arg security-logging-manager-binary)"; then
	log.error "Missing value for --security-logging-manager-binary"
	usage >&2
	exit 64
fi
if args.has system-email-notifier-binary && ! system_email_notifier_binary_path="$(args.require_arg system-email-notifier-binary)"; then
	log.error "Missing value for --system-email-notifier-binary"
	usage >&2
	exit 64
fi
if args.has security-email-notifier-binary && ! security_email_notifier_binary_path="$(args.require_arg security-email-notifier-binary)"; then
	log.error "Missing value for --security-email-notifier-binary"
	usage >&2
	exit 64
fi
if args.has system-metrics-binary && ! system_metrics_binary_path="$(args.require_arg system-metrics-binary)"; then
	log.error "Missing value for --system-metrics-binary"
	usage >&2
	exit 64
fi
if args.has zfs-metrics-binary && ! zfs_metrics_binary_path="$(args.require_arg zfs-metrics-binary)"; then
	log.error "Missing value for --zfs-metrics-binary"
	usage >&2
	exit 64
fi
if args.has system-logging-manager-cli-binary && ! system_logging_manager_cli_binary_path="$(args.require_arg system-logging-manager-cli-binary)"; then
	log.error "Missing value for --system-logging-manager-cli-binary"
	usage >&2
	exit 64
fi
if args.has security-logging-manager-cli-binary && ! security_logging_manager_cli_binary_path="$(args.require_arg security-logging-manager-cli-binary)"; then
	log.error "Missing value for --security-logging-manager-cli-binary"
	usage >&2
	exit 64
fi
if args.has system-metrics-cli-binary && ! system_metrics_cli_binary_path="$(args.require_arg system-metrics-cli-binary)"; then
	log.error "Missing value for --system-metrics-cli-binary"
	usage >&2
	exit 64
fi
if args.has zfs-metrics-cli-binary && ! zfs_metrics_cli_binary_path="$(args.require_arg zfs-metrics-cli-binary)"; then
	log.error "Missing value for --zfs-metrics-cli-binary"
	usage >&2
	exit 64
fi
if args.has web-gateway-binary && ! web_gateway_binary_path="$(args.require_arg web-gateway-binary)"; then
	log.error "Missing value for --web-gateway-binary"
	usage >&2
	exit 64
fi
if args.has resources-monitor-binary && ! resources_monitor_binary_path="$(args.require_arg resources-monitor-binary)"; then
	log.error "Missing value for --resources-monitor-binary"
	usage >&2
	exit 64
fi
if args.has admin-panel-assets && ! admin_panel_assets_path="$(args.require_arg admin-panel-assets)"; then
	log.error "Missing value for --admin-panel-assets"
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
if [ -z "$rbac_service_binary_path" ]; then
	rbac_service_binary_path="$output_dir/${package_name}-${package_arch}/rbac-service"
	"$LITE_NAS_REPO_ROOT/scripts/build-rbac-service-binary.sh" \
		"--output=${rbac_service_binary_path}"
fi

if [ -z "$system_logging_manager_binary_path" ]; then
	system_logging_manager_binary_path="$output_dir/${package_name}-${package_arch}/system-logging-manager"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-logging-manager-binary.sh" \
		"--output=${system_logging_manager_binary_path}"
fi

if [ -z "$security_logging_manager_binary_path" ]; then
	security_logging_manager_binary_path="$output_dir/${package_name}-${package_arch}/security-logging-manager"
	"$LITE_NAS_REPO_ROOT/scripts/build-security-logging-manager-binary.sh" \
		"--output=${security_logging_manager_binary_path}"
fi

if [ -z "$system_email_notifier_binary_path" ]; then
	system_email_notifier_binary_path="$output_dir/${package_name}-${package_arch}/system-email-notifier"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-email-notifier-binary.sh" \
		"--output=${system_email_notifier_binary_path}"
fi

if [ -z "$security_email_notifier_binary_path" ]; then
	security_email_notifier_binary_path="$output_dir/${package_name}-${package_arch}/security-email-notifier"
	"$LITE_NAS_REPO_ROOT/scripts/build-security-email-notifier-binary.sh" \
		"--output=${security_email_notifier_binary_path}"
fi

if [ -z "$system_metrics_binary_path" ]; then
	system_metrics_binary_path="$output_dir/${package_name}-${package_arch}/system-metrics"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-metrics-binary.sh" \
		"--output=${system_metrics_binary_path}"
fi
if [ -z "$zfs_metrics_binary_path" ]; then
	zfs_metrics_binary_path="$output_dir/${package_name}-${package_arch}/zfs-metrics"
	"$LITE_NAS_REPO_ROOT/scripts/build-zfs-metrics-binary.sh" \
		"--output=${zfs_metrics_binary_path}"
fi

if [ -z "$system_logging_manager_cli_binary_path" ]; then
	system_logging_manager_cli_binary_path="$output_dir/${package_name}-${package_arch}/system-logging-manager-cli"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-logging-manager-cli-binary.sh" \
		"--output=${system_logging_manager_cli_binary_path}"
fi

if [ -z "$security_logging_manager_cli_binary_path" ]; then
	security_logging_manager_cli_binary_path="$output_dir/${package_name}-${package_arch}/security-logging-manager-cli"
	"$LITE_NAS_REPO_ROOT/scripts/build-security-logging-manager-cli-binary.sh" \
		"--output=${security_logging_manager_cli_binary_path}"
fi

if [ -z "$system_metrics_cli_binary_path" ]; then
	system_metrics_cli_binary_path="$output_dir/${package_name}-${package_arch}/system-metrics-cli"
	"$LITE_NAS_REPO_ROOT/scripts/build-system-metrics-cli-binary.sh" \
		"--output=${system_metrics_cli_binary_path}"
fi
if [ -z "$zfs_metrics_cli_binary_path" ]; then
	zfs_metrics_cli_binary_path="$output_dir/${package_name}-${package_arch}/zfs-metrics-cli"
	"$LITE_NAS_REPO_ROOT/scripts/build-zfs-metrics-cli-binary.sh" \
		"--output=${zfs_metrics_cli_binary_path}"
fi

if [ -z "$web_gateway_binary_path" ]; then
	web_gateway_binary_path="$output_dir/${package_name}-${package_arch}/web-gateway"
	"$LITE_NAS_REPO_ROOT/scripts/build-web-gateway-binary.sh" \
		"--output=${web_gateway_binary_path}"
fi

if [ -z "$resources_monitor_binary_path" ]; then
	resources_monitor_binary_path="$output_dir/${package_name}-${package_arch}/resources-monitor"
	"$LITE_NAS_REPO_ROOT/scripts/build-resources-monitor-binary.sh" \
		"--output=${resources_monitor_binary_path}"
fi

if [ -z "$admin_panel_assets_path" ]; then
	admin_panel_assets_path="$output_dir/${package_name}-${package_arch}/admin-panel-assets"
	"$LITE_NAS_REPO_ROOT/scripts/build-admin-panel.sh" \
		"--output-dir=${admin_panel_assets_path}"
fi

if [ ! -f "$auth_service_binary_path" ]; then
	log.error "Missing auth-service binary: $auth_service_binary_path"
	exit 1
fi
if [ ! -f "$rbac_service_binary_path" ]; then
	log.error "Missing rbac-service binary: $rbac_service_binary_path"
	exit 1
fi

if [ ! -f "$system_logging_manager_binary_path" ]; then
	log.error "Missing system-logging-manager binary: $system_logging_manager_binary_path"
	exit 1
fi

if [ ! -f "$security_logging_manager_binary_path" ]; then
	log.error "Missing security-logging-manager binary: $security_logging_manager_binary_path"
	exit 1
fi

if [ ! -f "$system_email_notifier_binary_path" ]; then
	log.error "Missing system-email-notifier binary: $system_email_notifier_binary_path"
	exit 1
fi

if [ ! -f "$security_email_notifier_binary_path" ]; then
	log.error "Missing security-email-notifier binary: $security_email_notifier_binary_path"
	exit 1
fi

if [ ! -f "$system_metrics_binary_path" ]; then
	log.error "Missing system-metrics binary: $system_metrics_binary_path"
	exit 1
fi
if [ ! -f "$zfs_metrics_binary_path" ]; then
	log.error "Missing zfs-metrics binary: $zfs_metrics_binary_path"
	exit 1
fi

if [ ! -f "$system_logging_manager_cli_binary_path" ]; then
	log.error "Missing system-logging-manager-cli binary: $system_logging_manager_cli_binary_path"
	exit 1
fi

if [ ! -f "$security_logging_manager_cli_binary_path" ]; then
	log.error "Missing security-logging-manager-cli binary: $security_logging_manager_cli_binary_path"
	exit 1
fi

if [ ! -f "$system_metrics_cli_binary_path" ]; then
	log.error "Missing system-metrics-cli binary: $system_metrics_cli_binary_path"
	exit 1
fi
if [ ! -f "$zfs_metrics_cli_binary_path" ]; then
	log.error "Missing zfs-metrics-cli binary: $zfs_metrics_cli_binary_path"
	exit 1
fi

if [ ! -f "$web_gateway_binary_path" ]; then
	log.error "Missing web-gateway binary: $web_gateway_binary_path"
	exit 1
fi

if [ ! -f "$resources_monitor_binary_path" ]; then
	log.error "Missing resources-monitor binary: $resources_monitor_binary_path"
	exit 1
fi

adminPanelAssets.validateBuildOutput "$admin_panel_assets_path"

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
package.copyTree "$LITE_NAS_REPO_ROOT/scripts/runtime" "$package_root/usr/libexec/lite-nas/scripts/runtime"
install -D -m 0755 "$LITE_NAS_REPO_ROOT/scripts/deploy-configs.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-configs.sh"
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
install -D -m 0755 "$rbac_service_binary_path" \
	"$package_root/usr/libexec/lite-nas/rbac-service"
install -D -m 0755 "$system_logging_manager_binary_path" \
	"$package_root/usr/libexec/lite-nas/system-logging-manager"
install -D -m 0755 "$security_logging_manager_binary_path" \
	"$package_root/usr/libexec/lite-nas/security-logging-manager"
install -D -m 0755 "$system_email_notifier_binary_path" \
	"$package_root/usr/libexec/lite-nas/system-email-notifier"
install -D -m 0755 "$security_email_notifier_binary_path" \
	"$package_root/usr/libexec/lite-nas/security-email-notifier"
install -D -m 0755 "$system_metrics_binary_path" \
	"$package_root/usr/libexec/lite-nas/system-metrics"
install -D -m 0755 "$zfs_metrics_binary_path" \
	"$package_root/usr/libexec/lite-nas/zfs-metrics"
install -D -m 0755 "$system_logging_manager_cli_binary_path" \
	"$package_root/usr/libexec/lite-nas/system-logging-manager-cli"
install -D -m 0755 "$security_logging_manager_cli_binary_path" \
	"$package_root/usr/libexec/lite-nas/security-logging-manager-cli"
install -D -m 0755 "$system_metrics_cli_binary_path" \
	"$package_root/usr/libexec/lite-nas/system-metrics-cli"
install -D -m 0755 "$zfs_metrics_cli_binary_path" \
	"$package_root/usr/libexec/lite-nas/zfs-metrics-cli"
install -d -m 0755 "$package_root/usr/bin"
ln -sfn /usr/libexec/lite-nas/system-logging-manager-cli \
	"$package_root/usr/bin/system-logging-manager-cli"
ln -sfn /usr/libexec/lite-nas/security-logging-manager-cli \
	"$package_root/usr/bin/security-logging-manager-cli"
ln -sfn /usr/libexec/lite-nas/system-metrics-cli \
	"$package_root/usr/bin/system-metrics-cli"
ln -sfn /usr/libexec/lite-nas/zfs-metrics-cli \
	"$package_root/usr/bin/zfs-metrics-cli"
install -D -m 0755 "$web_gateway_binary_path" \
	"$package_root/usr/libexec/lite-nas/web-gateway"
install -D -m 0755 "$resources_monitor_binary_path" \
	"$package_root/usr/libexec/lite-nas/resources-monitor"
package.copyTree "$admin_panel_assets_path" \
	"$package_root/usr/libexec/lite-nas/admin-panel-assets"
adminPanelAssets.installFlat \
	"$admin_panel_assets_path" \
	"$package_root/usr/share/lite-nas/web-gateway/assets"

find "$package_root/usr/libexec/lite-nas/scripts" -type f -name "*.sh" -exec chmod 0755 {} +

find "$package_root/usr" -type d -exec chmod 0755 {} +
find "$package_root/usr" -type f -exec chmod 0644 {} +
# Keep executable mode on real binaries and scripts only; the usr/bin entries are symlinks.
chmod 0755 \
	"$package_root/usr/libexec/lite-nas/auth-service" \
	"$package_root/usr/libexec/lite-nas/rbac-service" \
	"$package_root/usr/libexec/lite-nas/system-logging-manager" \
	"$package_root/usr/libexec/lite-nas/security-logging-manager" \
	"$package_root/usr/libexec/lite-nas/system-email-notifier" \
	"$package_root/usr/libexec/lite-nas/security-email-notifier" \
	"$package_root/usr/libexec/lite-nas/system-metrics" \
	"$package_root/usr/libexec/lite-nas/zfs-metrics" \
	"$package_root/usr/libexec/lite-nas/system-logging-manager-cli" \
	"$package_root/usr/libexec/lite-nas/security-logging-manager-cli" \
	"$package_root/usr/libexec/lite-nas/system-metrics-cli" \
	"$package_root/usr/libexec/lite-nas/zfs-metrics-cli" \
	"$package_root/usr/libexec/lite-nas/web-gateway" \
	"$package_root/usr/libexec/lite-nas/scripts/deploy-configs.sh" \
	"$package_root/usr/libexec/lite-nas/scripts/runtime/deploy-package-runtime.sh" \
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
