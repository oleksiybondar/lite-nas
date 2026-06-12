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

deploy.resolveUserGroupOwner() {
	local user_name="$1"
	local group_name="$2"
	local fallback_owner="$3"

	if id -u "$user_name" >/dev/null 2>&1 && getent group "$group_name" >/dev/null 2>&1; then
		printf '%s:%s\n' "$user_name" "$group_name"
		return 0
	fi

	if getent group "$group_name" >/dev/null 2>&1; then
		printf 'root:%s\n' "$group_name"
		return 0
	fi

	printf '%s\n' "$fallback_owner"
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
	deploy.normalizePath 0711 "$litenas_config_dir" "$litenas_config_owner"
	deploy.normalizePath 0711 "$litenas_certificates_dir" "$litenas_config_owner"

	if [ -d "$litenas_config_dir" ]; then
		while IFS= read -r -d '' config_file; do
			deploy.normalizePath 0640 "$config_file" "$litenas_config_owner"
		done < <(find "$litenas_config_dir" -maxdepth 1 -type f -name '*.conf' -print0)
	fi
	deploy.normalizePath 0640 "$litenas_auth_config_file" "$litenas_config_owner"
	deploy.normalizePath 0640 "$litenas_rbac_config_file" "$(deploy.resolveUserGroupOwner "$litenas_rbac_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_network_metrics_config_file" "$(deploy.resolveUserGroupOwner "$litenas_network_metrics_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_system_metrics_config_file" "$(deploy.resolveUserGroupOwner "$litenas_system_metrics_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_zfs_metrics_config_file" "$(deploy.resolveUserGroupOwner "$litenas_zfs_metrics_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_resources_monitor_config_file" "$(deploy.resolveUserGroupOwner "$litenas_resources_monitor_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_system_logging_manager_config_file" "$(deploy.resolveUserGroupOwner "$litenas_system_logging_manager_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_security_logging_manager_config_file" "$(deploy.resolveUserGroupOwner "$litenas_security_logging_manager_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_system_email_notifier_config_file" "$(deploy.resolveUserGroupOwner "$litenas_system_email_notifier_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_security_email_notifier_config_file" "$(deploy.resolveUserGroupOwner "$litenas_security_email_notifier_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_web_gateway_config_file" "$(deploy.resolveUserGroupOwner "$litenas_web_gateway_runtime_user" "$litenas_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_cli_config_file" "$(deploy.resolveUserGroupOwner "$litenas_cli_runtime_user" "$litenas_cli_access_group" "$owner")"
	deploy.normalizePath 0640 "$litenas_network_cli_config_file" "$(deploy.resolveUserGroupOwner "$litenas_network_cli_runtime_user" "$litenas_network_cli_access_group" "$owner")"
	deploy.normalizePath 0640 "$litenas_zfs_cli_config_file" "$(deploy.resolveUserGroupOwner "$litenas_zfs_cli_runtime_user" "$litenas_zfs_cli_access_group" "$owner")"
	deploy.normalizePath 0640 "$litenas_system_logging_manager_cli_config_file" "$(deploy.resolveUserGroupOwner "$litenas_system_logging_manager_cli_runtime_user" "$litenas_system_logging_manager_cli_access_group" "$litenas_config_owner")"
	deploy.normalizePath 0640 "$litenas_security_logging_manager_cli_config_file" "$(deploy.resolveUserGroupOwner "$litenas_security_logging_manager_cli_runtime_user" "$litenas_security_logging_manager_cli_access_group" "$litenas_config_owner")"

	if [ -d "$litenas_system_email_notifier_templates_dir" ]; then
		deploy.normalizePath 0750 "$litenas_system_email_notifier_templates_dir" "$(deploy.resolveUserGroupOwner "$litenas_system_email_notifier_runtime_user" "$litenas_group" "$litenas_config_owner")"
		while IFS= read -r -d '' template_file; do
			deploy.normalizePath 0640 "$template_file" "$(deploy.resolveUserGroupOwner "$litenas_system_email_notifier_runtime_user" "$litenas_group" "$litenas_config_owner")"
		done < <(find "$litenas_system_email_notifier_templates_dir" -maxdepth 1 -type f -name '*.html' -print0)
	fi

	if [ -d "$litenas_security_email_notifier_templates_dir" ]; then
		deploy.normalizePath 0750 "$litenas_security_email_notifier_templates_dir" "$(deploy.resolveUserGroupOwner "$litenas_security_email_notifier_runtime_user" "$litenas_group" "$litenas_config_owner")"
		while IFS= read -r -d '' template_file; do
			deploy.normalizePath 0640 "$template_file" "$(deploy.resolveUserGroupOwner "$litenas_security_email_notifier_runtime_user" "$litenas_group" "$litenas_config_owner")"
		done < <(find "$litenas_security_email_notifier_templates_dir" -maxdepth 1 -type f -name '*.html' -print0)
	fi

	if [ -d "$litenas_transport_certificates_dir" ]; then
		deploy.normalizePath 0711 "$litenas_transport_certificates_dir" "$litenas_config_owner"
		deploy.normalizePath 0644 "$litenas_transport_ca_cert" "$owner"

		while IFS= read -r -d '' identity_dir; do
			identity_group="$(basename "$identity_dir")"
			if [ "$identity_group" = "$litenas_cli_certificate_user" ] && getent group "$litenas_cli_access_group" >/dev/null 2>&1; then
				deploy.normalizePath 0755 "$identity_dir" "$owner"
				credential_owner="$owner"
			elif [ "$identity_group" = "$litenas_network_cli_certificate_user" ] && getent group "$litenas_network_cli_access_group" >/dev/null 2>&1; then
				deploy.normalizePath 0755 "$identity_dir" "$owner"
				credential_owner="$owner"
			elif [ "$identity_group" = "$litenas_zfs_cli_certificate_user" ] && getent group "$litenas_zfs_cli_access_group" >/dev/null 2>&1; then
				deploy.normalizePath 0755 "$identity_dir" "$owner"
				credential_owner="$owner"
			elif [ "$identity_group" = "$litenas_system_logging_manager_cli_certificate_user" ] && getent group "$litenas_system_logging_manager_cli_access_group" >/dev/null 2>&1; then
				deploy.normalizePath 0750 "$identity_dir" "root:${litenas_system_logging_manager_cli_access_group}"
				credential_owner="root:${litenas_system_logging_manager_cli_access_group}"
			elif [ "$identity_group" = "$litenas_security_logging_manager_cli_certificate_user" ] && getent group "$litenas_security_logging_manager_cli_access_group" >/dev/null 2>&1; then
				deploy.normalizePath 0750 "$identity_dir" "root:${litenas_security_logging_manager_cli_access_group}"
				credential_owner="root:${litenas_security_logging_manager_cli_access_group}"
			elif getent group "$identity_group" >/dev/null 2>&1; then
				deploy.normalizePath 0750 "$identity_dir" "root:${identity_group}"
				credential_owner="root:${identity_group}"
			else
				deploy.normalizePath 0700 "$identity_dir" "root:root"
				credential_owner="root:root"
			fi
			while IFS= read -r -d '' credential_file; do
				if [ "${credential_file##*.}" = "csr" ]; then
					deploy.normalizePath 0600 "$credential_file" "$credential_owner"
				elif [ "$identity_group" = "$litenas_cli_certificate_user" ] || [ "$identity_group" = "$litenas_network_cli_certificate_user" ] || [ "$identity_group" = "$litenas_zfs_cli_certificate_user" ]; then
					deploy.normalizePath 0644 "$credential_file" "$credential_owner"
				elif [ "$identity_group" = "$litenas_system_logging_manager_cli_certificate_user" ] || [ "$identity_group" = "$litenas_security_logging_manager_cli_certificate_user" ]; then
					deploy.normalizePath 0640 "$credential_file" "$credential_owner"
				else
					deploy.normalizePath 0640 "$credential_file" "$credential_owner"
				fi
			done < <(find "$identity_dir" -maxdepth 1 -type f \( -name '*.crt' -o -name '*.key' -o -name '*.csr' \) -print0)
		done < <(find "$litenas_transport_certificates_dir" -mindepth 1 -maxdepth 1 -type d -print0)
	fi

	if [ -d "$litenas_nginx_certificates_dir" ]; then
		deploy.normalizePath 0700 "$litenas_nginx_certificates_dir" "$owner"
		while IFS= read -r -d '' certificate_file; do
			deploy.normalizePath 0640 "$certificate_file" "$owner"
		done < <(find "$litenas_nginx_certificates_dir" -maxdepth 1 -type f \( -name '*.crt' -o -name '*.key' \) -print0)
	fi

	if [ -d "$litenas_auth_certificates_dir" ]; then
		deploy.normalizePath 0750 "$litenas_auth_certificates_dir" "$litenas_config_owner"
		while IFS= read -r -d '' certificate_file; do
			deploy.normalizePath 0640 "$certificate_file" "$litenas_config_owner"
		done < <(find "$litenas_auth_certificates_dir" -maxdepth 1 -type f -name '*.crt' -print0)
		while IFS= read -r -d '' key_file; do
			deploy.normalizePath 0600 "$key_file" "$owner"
		done < <(find "$litenas_auth_certificates_dir" -maxdepth 1 -type f -name '*.key' -print0)
	fi

	if [ -d "$litenas_identities_certificates_dir" ]; then
		deploy.normalizePath 0750 "$litenas_identities_certificates_dir" "$litenas_config_owner"
		while IFS= read -r -d '' identity_file; do
			case "${identity_file##*.}" in
			key | srl)
				deploy.normalizePath 0600 "$identity_file" "$owner"
				;;
			crt)
				deploy.normalizePath 0640 "$identity_file" "$litenas_config_owner"
				;;
			esac
		done < <(find "$litenas_identities_certificates_dir" -maxdepth 1 -type f \( -name '*.crt' -o -name '*.key' -o -name '*.srl' \) -print0)

		while IFS= read -r -d '' identity_leaf_dir; do
			identity_group="$(basename "$identity_leaf_dir")"
			if getent group "$identity_group" >/dev/null 2>&1; then
				deploy.normalizePath 0750 "$identity_leaf_dir" "root:${identity_group}"
				credential_owner="root:${identity_group}"
			else
				deploy.normalizePath 0750 "$identity_leaf_dir" "$litenas_config_owner"
				credential_owner="$litenas_config_owner"
			fi

			while IFS= read -r -d '' credential_file; do
				case "${credential_file##*.}" in
				key)
					deploy.normalizePath 0600 "$credential_file" "$credential_owner"
					;;
				crt)
					deploy.normalizePath 0640 "$credential_file" "$credential_owner"
					;;
				csr)
					deploy.normalizePath 0600 "$credential_file" "$credential_owner"
					;;
				esac
			done < <(find "$identity_leaf_dir" -maxdepth 1 -type f \( -name '*.crt' -o -name '*.key' -o -name '*.csr' \) -print0)
		done < <(find "$litenas_identities_certificates_dir" -mindepth 1 -maxdepth 1 -type d -print0)
	fi

	log.popTask
}

deploy.normalizeUFW() {
	log.pushTask "Normalizing UFW config permissions"
	deploy.normalizePath 0755 "$default_dir" "$owner"
	deploy.normalizePath 0755 "$ufw_config_dir" "$owner"
	deploy.normalizePath 0644 "$ufw_default_config" "$owner"
	deploy.normalizePath 0644 "$ufw_runtime_config" "$owner"
	log.popTask
}

deploy.normalizeNginx() {
	log.pushTask "Normalizing nginx config permissions"
	deploy.normalizePath 0755 "$nginx_config_dir" "$owner"
	deploy.normalizePath 0755 "$nginx_sites_available_dir" "$owner"
	deploy.normalizePath 0755 "$nginx_sites_enabled_dir" "$owner"

	if [ -d "$nginx_sites_available_dir" ]; then
		while IFS= read -r -d '' config_file; do
			deploy.normalizePath 0644 "$config_file" "$owner"
		done < <(find "$nginx_sites_available_dir" -maxdepth 1 -type f -print0)
	fi

	log.popTask
}

deploy.normalizePostfix() {
	log.pushTask "Normalizing Postfix config permissions"
	deploy.normalizePath 0755 "$postfix_config_dir" "$owner"
	deploy.normalizePath 0755 "$postfix_config_dir/postfix.d" "$owner"

	if [ -d "$postfix_config_dir" ]; then
		while IFS= read -r -d '' postfix_file; do
			deploy.normalizePath 0644 "$postfix_file" "$owner"
		done < <(find "$postfix_config_dir" -maxdepth 2 -type f -print0)
	fi

	deploy.normalizePath 0600 "$postfix_config_dir/postfix.d/authentication.conf" "$owner"
	deploy.normalizePath 0600 "$postfix_config_dir/sasl_passwd" "$owner"
	deploy.normalizePath 0600 "$postfix_config_dir/sasl_passwd.db" "$owner"

	log.popTask
}

deploy.normalizeAppArmor() {
	log.pushTask "Normalizing AppArmor config permissions"
	deploy.normalizePath 0755 "$apparmor_config_dir" "$owner"

	if [ -d "$apparmor_config_dir" ]; then
		while IFS= read -r -d '' apparmor_file; do
			deploy.normalizePath 0644 "$apparmor_file" "$owner"
		done < <(find "$apparmor_config_dir" -maxdepth 2 -type f -print0)
	fi

	log.popTask
}

deploy.normalizeSystemd() {
	log.pushTask "Normalizing systemd unit permissions"
	deploy.normalizePath 0755 "$systemd_dir" "$owner"
	deploy.normalizePath 0755 "$systemd_system_dir" "$owner"

	if [ -d "$systemd_system_dir" ]; then
		while IFS= read -r -d '' unit_file; do
			deploy.normalizePath 0644 "$unit_file" "$owner"
		done < <(find "$systemd_system_dir" -maxdepth 1 -type f -name 'lite-nas-*.service' -print0)
	fi

	log.popTask
}

deploy.normalizeSudoers() {
	local sudoers_dir="$target_dir/sudoers.d"
	local sudoers_file

	log.pushTask "Normalizing sudoers drop-in permissions"
	deploy.normalizePath 0750 "$sudoers_dir" "$owner"

	if [ -d "$sudoers_dir" ]; then
		while IFS= read -r -d '' sudoers_file; do
			deploy.normalizePath 0440 "$sudoers_file" "$owner"
		done < <(find "$sudoers_dir" -maxdepth 1 -type f -name 'lite-nas-*' -print0)
	fi

	log.popTask
}

deploy.normalizeEtcPermissions() {
	local target_dir="${1:-${LITE_NAS_ETC_TARGET:-/etc}}"
	local owner="${LITE_NAS_ETC_OWNER:-root:root}"
	local litenas_group="${LITE_NAS_GROUP:-lite-nas}"
	local litenas_config_owner="root:${litenas_group}"
	local nats_certificate_owner="root:root"
	local litenas_cli_certificate_user="${LITE_NAS_SYSTEM_METRICS_CLI_CERT_USER:-lite-nas-system-metrics-cli}"
	local litenas_cli_access_group="${LITE_NAS_SYSTEM_METRICS_CLI_ACCESS_GROUP:-users}"
	local litenas_cli_runtime_user="${LITE_NAS_SYSTEM_METRICS_CLI_USER:-lite-nas-system-metrics-cli}"
	local litenas_network_cli_certificate_user="${LITE_NAS_NETWORK_METRICS_CLI_CERT_USER:-lite-nas-network-metrics-cli}"
	local litenas_network_cli_access_group="${LITE_NAS_NETWORK_METRICS_CLI_ACCESS_GROUP:-users}"
	local litenas_network_cli_runtime_user="${LITE_NAS_NETWORK_METRICS_CLI_USER:-lite-nas-network-metrics-cli}"
	local litenas_zfs_cli_certificate_user="${LITE_NAS_ZFS_METRICS_CLI_CERT_USER:-lite-nas-zfs-metrics-cli}"
	local litenas_zfs_cli_access_group="${LITE_NAS_ZFS_METRICS_CLI_ACCESS_GROUP:-users}"
	local litenas_zfs_cli_runtime_user="${LITE_NAS_ZFS_METRICS_CLI_USER:-lite-nas-zfs-metrics-cli}"
	local litenas_system_logging_manager_cli_certificate_user="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_CERT_USER:-lite-nas-sys-log-mgr-cli}"
	local litenas_system_logging_manager_cli_access_group="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_ACCESS_GROUP:-lite-nas-operator}"
	local litenas_system_logging_manager_cli_runtime_user="${LITE_NAS_SYSTEM_LOGGING_MANAGER_CLI_IDENTITY_USER:-lite-nas-sys-log-mgr-cli}"
	local litenas_security_logging_manager_cli_certificate_user="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_CERT_USER:-lite-nas-sec-log-mgr-cli}"
	local litenas_security_logging_manager_cli_access_group="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_ACCESS_GROUP:-lite-nas-security}"
	local litenas_security_logging_manager_cli_runtime_user="${LITE_NAS_SECURITY_LOGGING_MANAGER_CLI_IDENTITY_USER:-lite-nas-sec-log-mgr-cli}"
	local litenas_rbac_runtime_user="${LITE_NAS_RBAC_RUNTIME_USER:-lite-nas-rbac}"
	local litenas_network_metrics_runtime_user="${LITE_NAS_NETWORK_METRICS_RUNTIME_USER:-lite-nas-network-metrics}"
	local litenas_system_metrics_runtime_user="${LITE_NAS_SYSTEM_METRICS_RUNTIME_USER:-lite-nas-system-metrics}"
	local litenas_zfs_metrics_runtime_user="${LITE_NAS_ZFS_METRICS_RUNTIME_USER:-lite-nas-zfs-metrics}"
	local litenas_resources_monitor_runtime_user="${LITE_NAS_RESOURCES_MONITOR_RUNTIME_USER:-lite-nas-resources-monitor}"
	local litenas_system_logging_manager_runtime_user="${LITE_NAS_SYSTEM_LOGGING_MANAGER_RUNTIME_USER:-lite-nas-sys-log-mgr}"
	local litenas_security_logging_manager_runtime_user="${LITE_NAS_SECURITY_LOGGING_MANAGER_RUNTIME_USER:-lite-nas-sec-log-mgr}"
	local litenas_system_email_notifier_runtime_user="${LITE_NAS_SYSTEM_EMAIL_NOTIFIER_RUNTIME_USER:-lite-nas-sys-email-notifier}"
	local litenas_security_email_notifier_runtime_user="${LITE_NAS_SECURITY_EMAIL_NOTIFIER_RUNTIME_USER:-lite-nas-sec-email-notifier}"
	local litenas_web_gateway_runtime_user="${LITE_NAS_WEB_GATEWAY_RUNTIME_USER:-lite-nas-web-gateway}"
	local nats_main_config="$target_dir/nats-server.conf"
	local nats_config_dir="$target_dir/nats-server"
	local nats_certificate_dir="$nats_config_dir/certificates"
	local litenas_config_dir="$target_dir/lite-nas"
	local litenas_auth_config_file="$litenas_config_dir/auth.conf"
	local litenas_rbac_config_file="$litenas_config_dir/rbac-service.conf"
	local litenas_network_metrics_config_file="$litenas_config_dir/network-metrics.conf"
	local litenas_system_metrics_config_file="$litenas_config_dir/system-metrics.conf"
	local litenas_zfs_metrics_config_file="$litenas_config_dir/zfs-metrics.conf"
	local litenas_resources_monitor_config_file="$litenas_config_dir/resources-monitor.conf"
	local litenas_system_logging_manager_config_file="$litenas_config_dir/system-logging-manager.conf"
	local litenas_security_logging_manager_config_file="$litenas_config_dir/security-logging-manager.conf"
	local litenas_system_email_notifier_config_file="$litenas_config_dir/system-email-notifier.conf"
	local litenas_security_email_notifier_config_file="$litenas_config_dir/security-email-notifier.conf"
	local litenas_system_email_notifier_templates_dir="$litenas_config_dir/system-email-notifier"
	local litenas_security_email_notifier_templates_dir="$litenas_config_dir/security-email-notifier"
	local litenas_web_gateway_config_file="$litenas_config_dir/web-gateway.conf"
	local litenas_cli_config_file="$litenas_config_dir/system-metrics-cli.conf"
	local litenas_network_cli_config_file="$litenas_config_dir/network-metrics-cli.conf"
	local litenas_zfs_cli_config_file="$litenas_config_dir/zfs-metrics-cli.conf"
	local litenas_system_logging_manager_cli_config_file="$litenas_config_dir/system-logging-manager-cli.conf"
	local litenas_security_logging_manager_cli_config_file="$litenas_config_dir/security-logging-manager-cli.conf"
	local litenas_certificates_dir="$litenas_config_dir/certificates"
	local litenas_transport_certificates_dir="$litenas_certificates_dir/transport"
	local litenas_transport_ca_cert="$litenas_transport_certificates_dir/root-ca.crt"
	local litenas_nginx_certificates_dir="$litenas_certificates_dir/nginx"
	local litenas_auth_certificates_dir="$litenas_certificates_dir/auth"
	local litenas_identities_certificates_dir="$litenas_certificates_dir/identities"
	local default_dir="$target_dir/default"
	local apparmor_config_dir="$target_dir/apparmor.d"
	local postfix_config_dir="$target_dir/postfix"
	local ufw_config_dir="$target_dir/ufw"
	local ufw_default_config="$default_dir/ufw"
	local ufw_runtime_config="$ufw_config_dir/ufw.conf"
	local nginx_config_dir="$target_dir/nginx"
	local nginx_sites_available_dir="$nginx_config_dir/sites-available"
	local nginx_sites_enabled_dir="$nginx_config_dir/sites-enabled"
	local systemd_dir="$target_dir/systemd"
	local systemd_system_dir="$systemd_dir/system"
	local config_file
	local certificate_file
	local identity_dir
	local identity_group
	local credential_owner
	local credential_file
	local identity_leaf_dir
	local template_file

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

	deploy.normalizePath 0755 "$target_dir" "$owner"
	deploy.normalizeNATS
	deploy.normalizeLiteNAS
	deploy.normalizeUFW
	deploy.normalizeNginx
	deploy.normalizePostfix
	deploy.normalizeAppArmor
	deploy.normalizeSystemd
	deploy.normalizeSudoers
}
