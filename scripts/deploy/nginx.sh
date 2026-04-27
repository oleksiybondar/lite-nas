#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_NGINX_SERVICE_NAME="${LITE_NAS_NGINX_SERVICE_NAME:-nginx}"
readonly LITE_NAS_NGINX_CONFIG_ROOT="${LITE_NAS_NGINX_CONFIG_ROOT:-/etc/nginx}"
readonly LITE_NAS_NGINX_SITE_NAME="${LITE_NAS_NGINX_SITE_NAME:-lite-nas-web-gateway.conf}"
readonly LITE_NAS_NGINX_SITE_SOURCE="${LITE_NAS_NGINX_SITE_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/nginx/sites-available/$LITE_NAS_NGINX_SITE_NAME}"
readonly LITE_NAS_NGINX_SITES_AVAILABLE_DIR="${LITE_NAS_NGINX_SITES_AVAILABLE_DIR:-$LITE_NAS_NGINX_CONFIG_ROOT/sites-available}"
readonly LITE_NAS_NGINX_SITES_ENABLED_DIR="${LITE_NAS_NGINX_SITES_ENABLED_DIR:-$LITE_NAS_NGINX_CONFIG_ROOT/sites-enabled}"
readonly LITE_NAS_NGINX_SITE_TARGET="${LITE_NAS_NGINX_SITE_TARGET:-$LITE_NAS_NGINX_SITES_AVAILABLE_DIR/$LITE_NAS_NGINX_SITE_NAME}"
readonly LITE_NAS_NGINX_SITE_SYMLINK="${LITE_NAS_NGINX_SITE_SYMLINK:-$LITE_NAS_NGINX_SITES_ENABLED_DIR/$LITE_NAS_NGINX_SITE_NAME}"
readonly LITE_NAS_NGINX_CERTIFICATE_DIR="${LITE_NAS_NGINX_CERTIFICATE_DIR:-/etc/lite-nas/certificates/nginx}"
readonly LITE_NAS_NGINX_CERTIFICATE_FILE="${LITE_NAS_NGINX_CERTIFICATE_FILE:-$LITE_NAS_NGINX_CERTIFICATE_DIR/server.crt}"
readonly LITE_NAS_NGINX_CERTIFICATE_KEY_FILE="${LITE_NAS_NGINX_CERTIFICATE_KEY_FILE:-$LITE_NAS_NGINX_CERTIFICATE_DIR/server.key}"

deploy.nginx.requireTools() {
	local tool
	local tools=(
		install
		ln
		readlink
		rm
		systemctl
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required nginx tooling and retry."
	done
}

deploy.nginx.installSiteConfig() {
	if [ ! -f "$LITE_NAS_NGINX_SITE_SOURCE" ]; then
		log.error "Missing LiteNAS nginx site config source: $LITE_NAS_NGINX_SITE_SOURCE"
		exit 1
	fi

	install -d -m 0755 "$LITE_NAS_NGINX_SITES_AVAILABLE_DIR" "$LITE_NAS_NGINX_SITES_ENABLED_DIR"
	install -m 0644 "$LITE_NAS_NGINX_SITE_SOURCE" "$LITE_NAS_NGINX_SITE_TARGET"
}

deploy.nginx.ensureCertificateAssets() {
	local required_file
	local required_files=(
		"$LITE_NAS_NGINX_CERTIFICATE_FILE"
		"$LITE_NAS_NGINX_CERTIFICATE_KEY_FILE"
	)

	for required_file in "${required_files[@]}"; do
		if [ -f "$required_file" ]; then
			continue
		fi

		log.error "Missing nginx certificate asset: $required_file"
		log.error "Run scripts/rotate-nginx-certificates.sh before deploying nginx."
		exit 1
	done
}

deploy.nginx.enableSite() {
	ln -sfn "$LITE_NAS_NGINX_SITE_TARGET" "$LITE_NAS_NGINX_SITE_SYMLINK"
}

deploy.nginx.validateConfig() {
	log.requireCommand "nginx" "Install nginx and retry."
	log.pushTask "Validating nginx configuration"
	nginx -t
	log.popTask
}

deploy.nginx.enableAndStart() {
	systemctl enable "$LITE_NAS_NGINX_SERVICE_NAME.service"

	if systemctl is-active --quiet "$LITE_NAS_NGINX_SERVICE_NAME.service"; then
		systemctl reload "$LITE_NAS_NGINX_SERVICE_NAME.service"
		return 0
	fi

	systemctl start "$LITE_NAS_NGINX_SERVICE_NAME.service"
}

deploy.nginx.deploy() {
	local should_start="${1:-1}"

	deploy.nginx.installSiteConfig
	deploy.nginx.enableSite

	if [ "$should_start" = "1" ]; then
		deploy.nginx.ensureCertificateAssets
		deploy.nginx.validateConfig
		deploy.nginx.enableAndStart
		return 0
	fi

	log.info "Skipping nginx enable/start."
}
