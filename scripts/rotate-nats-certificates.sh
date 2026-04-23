#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/sudo-guard.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/normalize-etc-permissions.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/restart-affected-services.sh"

nats_certificate_dir="${LITE_NAS_NATS_CERTIFICATE_DIR:-/etc/nats-server/certificates}"
litenas_config_dir="${LITE_NAS_CONFIG_DIR:-/etc/liteNAS}"
litenas_certificate_dir="${LITE_NAS_CERTIFICATE_DIR:-$litenas_config_dir/certificates}"
litenas_user="${LITE_NAS_USER:-litenas}"
litenas_group="${LITE_NAS_GROUP:-litenas}"
certificate_days="${LITE_NAS_CERTIFICATE_DAYS:-825}"
root_ca_days="${LITE_NAS_ROOT_CA_DAYS:-3650}"
server_common_name="${LITE_NAS_NATS_SERVER_COMMON_NAME:-lite-nas-nats-server}"
server_alt_names="${LITE_NAS_NATS_SERVER_ALT_NAMES:-DNS:localhost,DNS:lite-nas,DNS:lite-nas.local,IP:127.0.0.1}"

read -r -a certificate_users <<<"${LITE_NAS_NATS_CERT_USERS:-system-metrics-svc system-metrics-cli}"

usage() {
	cat <<'MSG'
Usage: scripts/rotate-nats-certificates.sh [options]

Options:
  --user NAME  Add a NATS client certificate user. May be repeated.
  -h, --help   Show this help.

Environment:
  LITE_NAS_NATS_CERT_USERS         Space-separated users when --user is not set.
  LITE_NAS_USER                    Local system user for /etc/liteNAS. Default: litenas.
  LITE_NAS_GROUP                   Local system group for /etc/liteNAS. Default: litenas.
  LITE_NAS_CERTIFICATE_DAYS        Leaf certificate validity days. Default: 825.
  LITE_NAS_ROOT_CA_DAYS            Root CA validity days. Default: 3650.
  LITE_NAS_NATS_SERVER_ALT_NAMES   OpenSSL SAN list for the NATS server cert.
MSG
}

cli_certificate_users=()
while [ "$#" -gt 0 ]; do
	case "$1" in
	--user)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --user"
			usage >&2
			exit 2
		fi
		cli_certificate_users+=("$2")
		shift 2
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

if [ "${#cli_certificate_users[@]}" -gt 0 ]; then
	certificate_users=("${cli_certificate_users[@]}")
fi

sudo.guard.requireRoot "scripts/rotate-nats-certificates.sh"

for tool in openssl groupadd useradd chown chmod find; do
	log.requireCommand "$tool" "Install required runtime tooling and retry."
done

ensure_litenas_identity() {
	if ! getent group "$litenas_group" >/dev/null 2>&1; then
		log.info "Creating system group: $litenas_group"
		groupadd --system "$litenas_group"
	fi

	if ! id "$litenas_user" >/dev/null 2>&1; then
		log.info "Creating system user: $litenas_user"
		useradd \
			--system \
			--no-create-home \
			--home-dir /nonexistent \
			--shell /usr/sbin/nologin \
			--gid "$litenas_group" \
			"$litenas_user"
	fi
}

ensure_directories() {
	install -d -m 0750 "$nats_certificate_dir"
	install -d -m 0750 "$litenas_certificate_dir"
}

create_root_ca_if_missing() {
	local ca_key="$nats_certificate_dir/root-ca.key"
	local ca_certificate="$nats_certificate_dir/root-ca.crt"

	if [ -f "$ca_key" ] && [ -f "$ca_certificate" ]; then
		log.info "Root CA already exists; keeping existing CA."
		return
	fi

	log.info "Creating NATS root CA certificate."
	openssl req \
		-x509 \
		-newkey rsa:4096 \
		-sha256 \
		-days "$root_ca_days" \
		-nodes \
		-keyout "$ca_key" \
		-out "$ca_certificate" \
		-subj "/CN=LiteNAS NATS Root CA"
}

publish_client_ca_certificate() {
	local source_ca="$nats_certificate_dir/root-ca.crt"
	local target_ca="$litenas_certificate_dir/root-ca.crt"

	install -m 0640 "$source_ca" "$target_ca"
}

create_signed_certificate() {
	local common_name="$1"
	local certificate_dir="$2"
	local basename="$3"
	local extension_file="$4"
	local ca_key="$nats_certificate_dir/root-ca.key"
	local ca_certificate="$nats_certificate_dir/root-ca.crt"
	local key_file="$certificate_dir/$basename.key"
	local csr_file="$certificate_dir/$basename.csr"
	local certificate_file="$certificate_dir/$basename.crt"

	openssl req \
		-newkey rsa:4096 \
		-nodes \
		-keyout "$key_file" \
		-out "$csr_file" \
		-subj "/CN=$common_name"

	openssl x509 \
		-req \
		-in "$csr_file" \
		-CA "$ca_certificate" \
		-CAkey "$ca_key" \
		-CAcreateserial \
		-out "$certificate_file" \
		-days "$certificate_days" \
		-sha256 \
		-extfile "$extension_file"

	rm -f "$csr_file"
}

rotate_server_certificate() {
	local extension_file
	extension_file="$(mktemp)"

	cat >"$extension_file" <<EOF
basicConstraints = CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = $server_alt_names
EOF

	log.info "Rotating NATS server certificate."
	create_signed_certificate "$server_common_name" "$nats_certificate_dir" "server" "$extension_file"
	rm -f "$extension_file"
}

rotate_client_certificates() {
	local extension_file
	local certificate_user

	extension_file="$(mktemp)"

	cat >"$extension_file" <<'EOF'
basicConstraints = CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF

	for certificate_user in "${certificate_users[@]}"; do
		log.info "Rotating NATS client certificate: $certificate_user"
		create_signed_certificate "$certificate_user" "$litenas_certificate_dir" "$certificate_user" "$extension_file"
	done
	rm -f "$extension_file"
}

normalize_certificate_permissions() {
	local nats_group="root"

	if getent group nats >/dev/null 2>&1; then
		nats_group="nats"
	fi

	chown -R "root:$nats_group" "$nats_certificate_dir"
	chmod 0750 "$nats_certificate_dir"
	find "$nats_certificate_dir" -type f -name '*.crt' -exec chmod 0640 {} +
	find "$nats_certificate_dir" -type f -name '*.key' -exec chmod 0640 {} +
	find "$nats_certificate_dir" -type f -name '*.srl' -exec chmod 0640 {} +

	chown -R "root:$litenas_group" "$litenas_config_dir"
	chmod 0750 "$litenas_config_dir"
	chmod 0750 "$litenas_certificate_dir"
	find "$litenas_certificate_dir" -type f -name '*.crt' -exec chmod 0640 {} +
	find "$litenas_certificate_dir" -type f -name '*.key' -exec chmod 0640 {} +
	find "$litenas_certificate_dir" -type f -name '*.srl' -exec chmod 0640 {} +
}

log.pushTask "Preparing LiteNAS certificate identity"
ensure_litenas_identity
ensure_directories
log.popTask

log.pushTask "Rotating NATS certificates"
create_root_ca_if_missing
publish_client_ca_certificate
rotate_server_certificate
rotate_client_certificates
log.popTask

log.pushTask "Normalizing certificate permissions"
normalize_certificate_permissions
deploy.normalizeEtcPermissions /etc
log.popTask

deploy.restartAffectedServices

log.info "NATS certificates rotated."
