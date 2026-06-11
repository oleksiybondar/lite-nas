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
identity_certificate_dir="${LITE_NAS_AUTH_IDENTITY_CERTIFICATE_DIR:-$litenas_config_dir/certificates/identities}"
certificate_days="${LITE_NAS_AUTH_TOKEN_CERTIFICATE_DAYS:-825}"
identity_leaf_days="${LITE_NAS_AUTH_IDENTITY_CERTIFICATE_DAYS:-825}"
identity_root_ca_days="${LITE_NAS_AUTH_IDENTITY_ROOT_CA_DAYS:-3650}"
certificate_common_name="${LITE_NAS_AUTH_TOKEN_CERTIFICATE_COMMON_NAME:-lite-nas-auth-token-signing}"
identity_ca_common_name="${LITE_NAS_AUTH_IDENTITY_CA_COMMON_NAME:-lite-nas-auth-service-identity-root-ca}"
auth_identity_owner_user="${LITE_NAS_AUTH_IDENTITY_OWNER_USER:-lite-nas-auth-service}"
resources_monitor_identity_user="${LITE_NAS_RESOURCES_MONITOR_IDENTITY_USER:-lite-nas-resources-monitor}"
system_logging_manager_cli_identity_user="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_USER:-lite-nas-sys-log-mgr-cli}"
security_logging_manager_cli_identity_user="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_USER:-lite-nas-sec-log-mgr-cli}"
rotate_only_if_missing=0

usage() {
	cat <<'MSG'
Usage: scripts/rotate-auth-token-certificates.sh [options]

Options:
  --if-missing Rotate only when expected auth-token and auth-identity files are missing.
  -h, --help   Show this help.

Environment:
  LITE_NAS_AUTH_TOKEN_CERTIFICATE_DIR
                                   Directory for JWT signing certificate material.
                                   Default: /etc/lite-nas/certificates/auth.
  LITE_NAS_AUTH_TOKEN_CERTIFICATE_DAYS
                                   Certificate validity days. Default: 825.
  LITE_NAS_AUTH_TOKEN_CERTIFICATE_COMMON_NAME
                                   Certificate common name. Default: lite-nas-auth-token-signing.
  LITE_NAS_AUTH_IDENTITY_CERTIFICATE_DIR
                                   Directory for auth identity CA and leaf certificates.
                                   Default: /etc/lite-nas/certificates/identities.
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

for tool in getent groupadd id install mktemp openssl; do
	log.requireCommand "$tool" "Install required runtime tooling and retry."
done

ensure_litenas_group() {
	if getent group "$litenas_group" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating system group: $litenas_group"
	groupadd --system "$litenas_group"
}

ensure_identity_groups() {
	local identity_group
	for identity_group in \
		"$auth_identity_owner_user" \
		"$resources_monitor_identity_user" \
		"$system_logging_manager_cli_identity_user" \
		"$security_logging_manager_cli_identity_user"; do
		if getent group "$identity_group" >/dev/null 2>&1; then
			continue
		fi
		log.info "Creating identity group: $identity_group"
		groupadd --system "$identity_group"
	done
}

ensure_directories() {
	install -d -m 0750 -o root -g "$litenas_group" "$litenas_config_dir"
	install -d -m 0750 -o root -g "$litenas_group" "$certificate_dir"
	install -d -m 0750 -o root -g "$litenas_group" "$identity_certificate_dir"
}

required_certificates_exist() {
	[ -f "$certificate_dir/token-signing.key" ] &&
		[ -f "$certificate_dir/token-signing.crt" ] &&
		[ -f "$identity_certificate_dir/root-ca.key" ] &&
		[ -f "$identity_certificate_dir/root-ca.crt" ] &&
		[ -f "$identity_certificate_dir/$auth_identity_owner_user/client.key" ] &&
		[ -f "$identity_certificate_dir/$auth_identity_owner_user/client.crt" ] &&
		[ -f "$identity_certificate_dir/$resources_monitor_identity_user/client.key" ] &&
		[ -f "$identity_certificate_dir/$resources_monitor_identity_user/client.crt" ] &&
		[ -f "$identity_certificate_dir/$system_logging_manager_cli_identity_user/client.key" ] &&
		[ -f "$identity_certificate_dir/$system_logging_manager_cli_identity_user/client.crt" ] &&
		[ -f "$identity_certificate_dir/$security_logging_manager_cli_identity_user/client.key" ] &&
		[ -f "$identity_certificate_dir/$security_logging_manager_cli_identity_user/client.crt" ]
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

ensure_identity_leaf_directory() {
	local identity_user="$1"
	install -d -m 0750 -o root -g "$litenas_group" "$identity_certificate_dir/$identity_user"
}

create_identity_root_ca_if_missing() {
	if [ -f "$identity_certificate_dir/root-ca.key" ] && [ -f "$identity_certificate_dir/root-ca.crt" ]; then
		return 0
	fi

	log.info "Creating auth identity root CA."
	openssl req \
		-x509 \
		-newkey rsa:4096 \
		-sha256 \
		-days "$identity_root_ca_days" \
		-nodes \
		-keyout "$identity_certificate_dir/root-ca.key" \
		-out "$identity_certificate_dir/root-ca.crt" \
		-subj "/CN=$identity_ca_common_name"
}

rotate_identity_leaf_certificate() {
	local identity_user="$1"
	local leaf_dir="$identity_certificate_dir/$identity_user"
	local key_file="$leaf_dir/client.key"
	local csr_file="$leaf_dir/client.csr"
	local crt_file="$leaf_dir/client.crt"
	local extension_file

	extension_file="$(mktemp)"
	cat >"$extension_file" <<EOF
basicConstraints = CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
subjectAltName = DNS:$identity_user
EOF

	log.info "Rotating auth identity certificate: $identity_user"
	openssl req \
		-newkey rsa:4096 \
		-nodes \
		-keyout "$key_file" \
		-out "$csr_file" \
		-subj "/CN=$identity_user"

	openssl x509 \
		-req \
		-in "$csr_file" \
		-CA "$identity_certificate_dir/root-ca.crt" \
		-CAkey "$identity_certificate_dir/root-ca.key" \
		-CAcreateserial \
		-out "$crt_file" \
		-days "$identity_leaf_days" \
		-sha256 \
		-extfile "$extension_file"

	rm -f "$csr_file" "$extension_file"
}

rotate_identity_certificates() {
	local identity_user
	create_identity_root_ca_if_missing

	for identity_user in \
		"$auth_identity_owner_user" \
		"$resources_monitor_identity_user" \
		"$system_logging_manager_cli_identity_user" \
		"$security_logging_manager_cli_identity_user"; do
		ensure_identity_leaf_directory "$identity_user"
		rotate_identity_leaf_certificate "$identity_user"
	done
}

owner_for_identity() {
	local identity_user="$1"
	if id -u "$identity_user" >/dev/null 2>&1; then
		printf '%s:%s' "$identity_user" "$identity_user"
		return 0
	fi

	if getent group "$identity_user" >/dev/null 2>&1; then
		printf 'root:%s' "$identity_user"
		return 0
	fi

	printf 'root:%s' "$litenas_group"
}

normalize_certificate_permissions() {
	local identity_user
	local identity_owner

	chown root:root "$certificate_dir/token-signing.key"
	chmod 0600 "$certificate_dir/token-signing.key"
	chown "root:$litenas_group" "$certificate_dir/token-signing.crt"
	chmod 0640 "$certificate_dir/token-signing.crt"
	chown "root:$litenas_group" "$certificate_dir"
	chmod 0750 "$certificate_dir"

	chown "root:$litenas_group" "$identity_certificate_dir"
	chmod 0750 "$identity_certificate_dir"

	identity_owner="$(owner_for_identity "$auth_identity_owner_user")"
	chown "$identity_owner" "$identity_certificate_dir/root-ca.key"
	chmod 0600 "$identity_certificate_dir/root-ca.key"
	chown "$identity_owner" "$identity_certificate_dir/root-ca.crt"
	chmod 0640 "$identity_certificate_dir/root-ca.crt"
	if [ -f "$identity_certificate_dir/root-ca.srl" ]; then
		chown "$identity_owner" "$identity_certificate_dir/root-ca.srl"
		chmod 0600 "$identity_certificate_dir/root-ca.srl"
	fi

	for identity_user in \
		"$auth_identity_owner_user" \
		"$resources_monitor_identity_user" \
		"$system_logging_manager_cli_identity_user" \
		"$security_logging_manager_cli_identity_user"; do
		identity_owner="$(owner_for_identity "$identity_user")"
		chown "$identity_owner" "$identity_certificate_dir/$identity_user"
		chmod 0750 "$identity_certificate_dir/$identity_user"
		chown "$identity_owner" "$identity_certificate_dir/$identity_user/client.key"
		chmod 0600 "$identity_certificate_dir/$identity_user/client.key"
		chown "$identity_owner" "$identity_certificate_dir/$identity_user/client.crt"
		chmod 0640 "$identity_certificate_dir/$identity_user/client.crt"
	done
}

log.pushTask "Preparing LiteNAS auth token certificate directory"
ensure_litenas_group
ensure_identity_groups
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

log.pushTask "Rotating LiteNAS auth identity certificates"
rotate_identity_certificates
log.popTask

log.pushTask "Normalizing LiteNAS auth token certificate permissions"
normalize_certificate_permissions
deploy.normalizeEtcPermissions /etc
log.popTask

log.info "LiteNAS auth token and identity certificates rotated."
