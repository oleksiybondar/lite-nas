#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/normalize-etc-permissions.sh"

litenas_config_dir="${LITE_NAS_CONFIG_DIR:-/etc/lite-nas}"
litenas_group="${LITE_NAS_GROUP:-lite-nas}"
certificate_dir="${LITE_NAS_AUTH_TOKEN_CERTIFICATE_DIR:-$litenas_config_dir/certificates/auth}"
certificate_days="${LITE_NAS_AUTH_TOKEN_CERTIFICATE_DAYS:-825}"
certificate_common_name="${LITE_NAS_AUTH_TOKEN_CERTIFICATE_COMMON_NAME:-lite-nas-auth-token-signing}"
rotate_only_if_missing=0

usage() {
	cat <<'MSG'
Usage: scripts/rotate-auth-token-certificates.sh [options]

Options:
  --if-missing Rotate only when token-signing.key and token-signing.crt are missing.
  -h, --help   Show this help.

Environment:
  LITE_NAS_AUTH_TOKEN_CERTIFICATE_DIR
                                   Directory for JWT signing certificate material.
                                   Default: /etc/lite-nas/certificates/auth.
  LITE_NAS_AUTH_TOKEN_CERTIFICATE_DAYS
                                   Certificate validity days. Default: 825.
  LITE_NAS_AUTH_TOKEN_CERTIFICATE_COMMON_NAME
                                   Certificate common name. Default: lite-nas-auth-token-signing.
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

sudo.guard.requireRoot "scripts/rotate-auth-token-certificates.sh"

for tool in getent groupadd install mktemp openssl; do
	log.requireCommand "$tool" "Install required runtime tooling and retry."
done

ensure_litenas_group() {
	if getent group "$litenas_group" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating system group: $litenas_group"
	groupadd --system "$litenas_group"
}

ensure_directories() {
	install -d -m 0750 -o root -g "$litenas_group" "$litenas_config_dir"
	install -d -m 0750 -o root -g "$litenas_group" "$certificate_dir"
}

required_certificates_exist() {
	[ -f "$certificate_dir/token-signing.key" ] && [ -f "$certificate_dir/token-signing.crt" ]
}

rotate_certificate() {
	local config_file

	config_file="$(mktemp)"
	cat >"$config_file" <<EOF
[req]
prompt = no
distinguished_name = req_distinguished_name
x509_extensions = v3_req

[req_distinguished_name]
CN = $certificate_common_name

[v3_req]
basicConstraints = CA:FALSE
keyUsage = digitalSignature
EOF

	log.info "Rotating LiteNAS auth token signing certificate."
	openssl genpkey \
		-algorithm Ed25519 \
		-out "$certificate_dir/token-signing.key"

	openssl req \
		-new \
		-x509 \
		-days "$certificate_days" \
		-key "$certificate_dir/token-signing.key" \
		-out "$certificate_dir/token-signing.crt" \
		-config "$config_file"

	rm -f "$config_file"
}

normalize_certificate_permissions() {
	chown root:root "$certificate_dir/token-signing.key"
	chmod 0600 "$certificate_dir/token-signing.key"
	chown "root:$litenas_group" "$certificate_dir/token-signing.crt"
	chmod 0640 "$certificate_dir/token-signing.crt"
	chown "root:$litenas_group" "$certificate_dir"
	chmod 0750 "$certificate_dir"
}

log.pushTask "Preparing LiteNAS auth token certificate directory"
ensure_litenas_group
ensure_directories
log.popTask

if [ "$rotate_only_if_missing" -eq 1 ] && required_certificates_exist; then
	log.info "LiteNAS auth token signing certificate already exists; skipping rotation."
	log.pushTask "Normalizing LiteNAS auth token certificate permissions"
	normalize_certificate_permissions
	deploy.normalizeEtcPermissions /etc
	log.popTask
	exit 0
fi

log.pushTask "Rotating LiteNAS auth token signing certificate"
rotate_certificate
log.popTask

log.pushTask "Normalizing LiteNAS auth token certificate permissions"
normalize_certificate_permissions
deploy.normalizeEtcPermissions /etc
log.popTask

log.info "LiteNAS auth token signing certificate rotated."
