#!/usr/bin/env bash

deploy.normalizePath() {
	local mode="$1"
	local path="$2"
	local owner="$3"

	if [ ! -e "$path" ]; then
		return 0
	fi

	chown "$owner" "$path"
	chmod "$mode" "$path"
}

deploy.normalizeNATS() {
	log.pushTask "Normalizing LiteNAS etc permissions"
	deploy.normalizePath 0644 "$nats_main_config" "$owner"
	deploy.normalizePath 0755 "$nats_config_dir" "$owner"
	deploy.normalizePath 0750 "$nats_certificate_dir" "$nats_certificate_owner"

	if [ -d "$nats_config_dir" ]; then
		while IFS= read -r -d '' config_file; do
			deploy.normalizePath 0644 "$config_file" "$owner"
		done < <(find "$nats_config_dir" -maxdepth 1 -type f -name '*.conf' -print0)
	fi

	if [ -d "$nats_certificate_dir" ]; then
		while IFS= read -r -d '' certificate_file; do
			deploy.normalizePath 0640 "$certificate_file" "$nats_certificate_owner"
		done < <(find "$nats_certificate_dir" -maxdepth 1 -type f \( -name '*.crt' -o -name '*.key' -o -name '*.srl' \) -print0)
	fi
	log.popTask
}

deploy.normalizeLiteNAS() {
	log.pushTask "Normalizing LiteNAS service config permissions"
	deploy.normalizePath 0750 "$litenas_config_dir" "$litenas_config_owner"
	deploy.normalizePath 0750 "$litenas_certificates_dir" "$litenas_config_owner"
	deploy.normalizePath 0640 "$litenas_ca_cert" "$litenas_config_owner"

	if [ -d "$litenas_config_dir" ]; then
		while IFS= read -r -d '' config_file; do
			deploy.normalizePath 0640 "$config_file" "$litenas_config_owner"
		done < <(find "$litenas_config_dir" -maxdepth 1 -type f -name '*.conf' -print0)
	fi

	if [ -d "$litenas_certificates_dir" ]; then
		while IFS= read -r -d '' identity_dir; do
			if [ "$identity_dir" = "$litenas_certificates_dir/root-ca.crt" ]; then
				continue
			fi
			identity_group="$(basename "$identity_dir")"
			if getent group "$identity_group" >/dev/null 2>&1; then
				deploy.normalizePath 0750 "$identity_dir" "root:${identity_group}"
				credential_owner="root:${identity_group}"
			else
				deploy.normalizePath 0700 "$identity_dir" "root:root"
				credential_owner="root:root"
			fi
			while IFS= read -r -d '' credential_file; do
				deploy.normalizePath 0640 "$credential_file" "$credential_owner"
			done < <(find "$identity_dir" -maxdepth 1 -type f \( -name '*.crt' -o -name '*.key' -o -name '*.csr' \) -print0)
		done < <(find "$litenas_certificates_dir" -mindepth 1 -maxdepth 1 -type d -print0)
	fi

	log.popTask
}

deploy.normalizeEtcPermissions() {
	local target_dir="${1:-${LITE_NAS_ETC_TARGET:-/etc}}"
	local owner="${LITE_NAS_ETC_OWNER:-root:root}"
	local litenas_group="${LITE_NAS_GROUP:-lite-nas}"
	local litenas_config_owner="root:${litenas_group}"
	local nats_certificate_owner="root:root"
	local nats_main_config="$target_dir/nats-server.conf"
	local nats_config_dir="$target_dir/nats-server"
	local nats_certificate_dir="$nats_config_dir/certificates"
	local litenas_config_dir="$target_dir/liteNAS"
	local litenas_certificates_dir="$litenas_config_dir/certificates"
	local litenas_ca_cert="$litenas_certificates_dir/root-ca.crt"
	local config_file
	local certificate_file
	local identity_dir
	local identity_group
	local credential_owner
	local credential_file

	if [ ! -d "$target_dir" ]; then
		log.error "Missing target etc directory: $target_dir"
		exit 1
	fi

	if ! getent group "$litenas_group" >/dev/null 2>&1; then
		litenas_config_owner="root:root"
	fi

	if getent group nats >/dev/null 2>&1; then
		nats_certificate_owner="root:nats"
	fi

	deploy.normalizeNATS
	deploy.normalizeLiteNAS
}
