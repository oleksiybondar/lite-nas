#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/normalize-etc-permissions.sh"

litenas_config_dir="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
certificate_dir="${LITE_NAS_NGINX_CERTIFICATE_DIR:-$litenas_config_dir/certificates/nginx}"
certificate_days="${LITE_NAS_CERTIFICATE_DAYS:-825}"
certificate_common_name="${LITE_NAS_NGINX_CERTIFICATE_COMMON_NAME:-lite-nas-web-gateway}"
certificate_alt_names="${LITE_NAS_NGINX_CERTIFICATE_ALT_NAMES:-DNS:localhost,DNS:lite-nas,DNS:lite-nas.local,IP:127.0.0.1}"
moments_certificate_alt_names="${LITE_NAS_MOMENTS_CERTIFICATE_ALT_NAMES:-DNS:moments.localhost}"
rotate_only_if_missing=0

usage() {
	cat <<'MSG'
Usage: scripts/rotate-nginx-certificates.sh [options]

Options:
  --if-missing Generate only missing certificate pairs. Preserve existing trusted certificates.
  -h, --help   Show this help.

Environment:
  LITE_NAS_CERTIFICATE_DAYS               Self-signed certificate validity days. Default: 825.
  LITE_NAS_NGINX_CERTIFICATE_COMMON_NAME  Certificate common name. Default: lite-nas-web-gateway.
  LITE_NAS_NGINX_CERTIFICATE_ALT_NAMES    Admin host OpenSSL SAN list.
  LITE_NAS_MOMENTS_CERTIFICATE_ALT_NAMES  Moments host OpenSSL SAN list. Default: moments.localhost.
MSG
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--if-missing)
		rotate_only_if_missing=1
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

sudo.guard.requireRoot "scripts/rotate-nginx-certificates.sh"

for tool in cmp getent groupadd install mktemp openssl; do
	log.requireCommand "$tool" "Install required runtime tooling and retry."
done

ensure_litenas_group() {
	local litenas_group="${LITE_NAS_GROUP:-lite-nas}"

	if getent group "$litenas_group" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating system group: $litenas_group"
	groupadd --system "$litenas_group"
}

ensure_directories() {
	local litenas_group="${LITE_NAS_GROUP:-lite-nas}"

	install -d -m 0750 -o root -g "$litenas_group" "$litenas_config_dir"
	install -d -m 0700 -o root -g root "$certificate_dir"
}

certificate_pair_exists() {
	local name="$1"

	[ -f "$certificate_dir/$name.crt" ] && [ -f "$certificate_dir/$name.key" ] || return 1
	openssl x509 -in "$certificate_dir/$name.crt" -noout >/dev/null 2>&1 || return 1
	openssl pkey -in "$certificate_dir/$name.key" -noout >/dev/null 2>&1 || return 1
	cmp -s \
		<(openssl x509 -in "$certificate_dir/$name.crt" -pubkey -noout) \
		<(openssl pkey -in "$certificate_dir/$name.key" -pubout)
}

rotate_certificate() {
	local name="$1"
	local common_name="$2"
	local alt_names="$3"
	local config_file

	config_file="$(mktemp)"
	cat >"$config_file" <<EOF
[req]
prompt = no
distinguished_name = req_distinguished_name
x509_extensions = v3_req

[req_distinguished_name]
CN = $common_name

[v3_req]
basicConstraints = CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = $alt_names
EOF

	log.info "Rotating LiteNAS nginx $name certificate."
	command.quietStderrUnlessFailure openssl req \
		-x509 \
		-newkey rsa:4096 \
		-sha256 \
		-days "$certificate_days" \
		-nodes \
		-keyout "$certificate_dir/$name.key" \
		-out "$certificate_dir/$name.crt" \
		-config "$config_file"

	rm -f "$config_file"
}

normalize_certificate_permissions() {
	chown -R root:root "$certificate_dir"
	chmod 0700 "$certificate_dir"
	find "$certificate_dir" -type f -name '*.crt' -exec chmod 0640 {} +
	find "$certificate_dir" -type f -name '*.key' -exec chmod 0640 {} +
}

log.pushTask "Preparing LiteNAS nginx certificate directory"
ensure_litenas_group
ensure_directories
log.popTask

if [ "$rotate_only_if_missing" -eq 1 ] && certificate_pair_exists server && certificate_pair_exists moments; then
	log.info "Both LiteNAS nginx certificates already exist; preserving browser trust."
	exit 0
fi

log.pushTask "Rotating LiteNAS nginx certificate"
if [ "$rotate_only_if_missing" -eq 0 ] || ! certificate_pair_exists server; then
	rotate_certificate server "$certificate_common_name" "$certificate_alt_names"
else
	log.info "Preserving the existing admin certificate and browser trust."
fi
if [ "$rotate_only_if_missing" -eq 0 ] || ! certificate_pair_exists moments; then
	rotate_certificate moments moments.localhost "$moments_certificate_alt_names"
else
	log.info "Preserving the existing Moments certificate."
fi
log.popTask

log.pushTask "Normalizing LiteNAS nginx certificate permissions"
normalize_certificate_permissions
deploy.normalizeEtcPermissions /etc
log.popTask

log.info "LiteNAS nginx certificate rotated."
