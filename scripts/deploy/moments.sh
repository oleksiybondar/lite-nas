#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/nginx.sh"

# Installs the loopback Moments API and its independently served browser assets.
deploy.moments.deploy() {
	local binary_path="$1"
	local assets_source="$2"
	local should_start="${3:-1}"
	binary_path="$(realpath -- "$binary_path")"
	assets_source="$(realpath -- "$assets_source")"

	if [ ! -f "$binary_path" ] || [ ! -f "$assets_source/index.html" ] || [ ! -d "$assets_source/assets" ]; then
		log.error "Missing Moments binary or frontend assets."
		exit 1
	fi

	deploy.ensureSystemGroup lite-nas-moments "Moments service group"
	if ! id lite-nas-moments >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin --gid lite-nas-moments --groups lite-nas lite-nas-moments
	fi

	install -D -m 0755 "$binary_path" /usr/libexec/lite-nas/moments
	deploy.installMergedConfig "$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/moments.conf" /etc/lite-nas/moments.conf \
		lite-nas-moments lite-nas
	install -d -m 0750 -o lite-nas-moments -g lite-nas-moments /var/lib/lite-nas/moments
	install -d -m 0750 -o lite-nas-moments -g lite-nas-moments /var/lib/lite-nas/moments/thumbnails
	install -d -m 0751 /var/log/lite-nas
	if [ ! -f /var/log/lite-nas/moments.log ]; then
		install -D -m 0640 -o lite-nas-moments -g lite-nas-moments /dev/null /var/log/lite-nas/moments.log
	fi
	install -D -m 0644 "$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-moments.service" /etc/systemd/system/lite-nas-moments.service

	install -d -m 0755 /usr/share/lite-nas/moments /usr/share/lite-nas/nginx
	cp -R "$assets_source/." /usr/share/lite-nas/moments/
	local place_data_source="$LITE_NAS_REPO_ROOT/services/moments/data/cities1000.zip"
	if [ -f "$place_data_source" ]; then
		install -m 0644 "$place_data_source" /usr/share/lite-nas/moments/cities1000.zip
		install -m 0644 "$LITE_NAS_REPO_ROOT/services/moments/data/README.md" /usr/share/lite-nas/moments/GEONAMES-ATTRIBUTION.md
	elif [ ! -f /usr/share/lite-nas/moments/cities1000.zip ]; then
		log.error "Moments offline city data is missing"
		return 1
	fi
	find /usr/share/lite-nas/moments -type d -exec chmod 0755 {} +
	find /usr/share/lite-nas/moments -type f -exec chmod 0644 {} +
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nginx-certificates.sh" --if-missing
	deploy.nginx.deploy 0
	if [ "$should_start" = "1" ]; then
		deploy.nginx.ensureCertificateAssets
		deploy.nginx.validateConfig
		deploy.moments.enableAndStart
		deploy.nginx.enableAndStart
	fi
}

# Enables the Moments service after all system dependencies are ready.
deploy.moments.enableAndStart() {
	deploy.enableAndRefreshService lite-nas-moments.service
}
