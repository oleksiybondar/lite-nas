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

deploy.normalizeEtcPermissions() {
	local target_dir="${1:-${LITE_NAS_ETC_TARGET:-/etc}}"
	local owner="${LITE_NAS_ETC_OWNER:-root:root}"
	local nats_certificate_owner="root:root"
	local nats_main_config="$target_dir/nats-server.conf"
	local nats_config_dir="$target_dir/nats-server"
	local nats_certificate_dir="$nats_config_dir/certificates"
	local config_file
	local certificate_file

	if [ ! -d "$target_dir" ]; then
		log.error "Missing target etc directory: $target_dir"
		exit 1
	fi

	if getent group nats >/dev/null 2>&1; then
		nats_certificate_owner="root:nats"
	fi

	deploy.normalizeNATS
}
