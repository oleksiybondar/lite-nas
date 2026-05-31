#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_POSTFIX_SERVICE_NAME="${LITE_NAS_POSTFIX_SERVICE_NAME:-postfix}"
readonly LITE_NAS_POSTFIX_CONFIG_SOURCE_DIR="${LITE_NAS_POSTFIX_CONFIG_SOURCE_DIR:-$LITE_NAS_REPO_ROOT/configs/etc/postfix}"
readonly LITE_NAS_POSTFIX_CONFIG_TARGET_DIR="${LITE_NAS_POSTFIX_CONFIG_TARGET_DIR:-/etc/postfix}"
readonly LITE_NAS_POSTFIX_MAIN_CF_SOURCE="${LITE_NAS_POSTFIX_MAIN_CF_SOURCE:-$LITE_NAS_POSTFIX_CONFIG_SOURCE_DIR/main.cf}"
readonly LITE_NAS_POSTFIX_MASTER_CF_SOURCE="${LITE_NAS_POSTFIX_MASTER_CF_SOURCE:-$LITE_NAS_POSTFIX_CONFIG_SOURCE_DIR/master.cf}"
readonly LITE_NAS_POSTFIX_MAIN_CF_TARGET="${LITE_NAS_POSTFIX_MAIN_CF_TARGET:-$LITE_NAS_POSTFIX_CONFIG_TARGET_DIR/main.cf}"
readonly LITE_NAS_POSTFIX_MASTER_CF_TARGET="${LITE_NAS_POSTFIX_MASTER_CF_TARGET:-$LITE_NAS_POSTFIX_CONFIG_TARGET_DIR/master.cf}"
readonly LITE_NAS_POSTFIX_AUTH_CONFIG_SOURCE="${LITE_NAS_POSTFIX_AUTH_CONFIG_SOURCE:-$LITE_NAS_POSTFIX_CONFIG_SOURCE_DIR/postfix.d/authentication.conf.example}"
readonly LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET_DIR="${LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET_DIR:-$LITE_NAS_POSTFIX_CONFIG_TARGET_DIR/postfix.d}"
readonly LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET="${LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET:-$LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET_DIR/authentication.conf}"
readonly LITE_NAS_POSTFIX_SASL_PASSWD_TARGET="${LITE_NAS_POSTFIX_SASL_PASSWD_TARGET:-$LITE_NAS_POSTFIX_CONFIG_TARGET_DIR/sasl_passwd}"

deploy.postfix.requireTools() {
	local tool
	local tools=(install mkdir postconf postfix postmap rm systemctl)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required Postfix/AppArmor tooling and retry."
	done
}

deploy.postfix.installConfig() {
	if [ ! -f "$LITE_NAS_POSTFIX_MAIN_CF_SOURCE" ]; then
		log.error "Missing Postfix main.cf source: $LITE_NAS_POSTFIX_MAIN_CF_SOURCE"
		exit 1
	fi
	if [ ! -f "$LITE_NAS_POSTFIX_MASTER_CF_SOURCE" ]; then
		log.error "Missing Postfix master.cf source: $LITE_NAS_POSTFIX_MASTER_CF_SOURCE"
		exit 1
	fi

	install -d -m 0755 "$LITE_NAS_POSTFIX_CONFIG_TARGET_DIR"
	install -d -m 0755 "$LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET_DIR"
	install -m 0644 "$LITE_NAS_POSTFIX_MAIN_CF_SOURCE" "$LITE_NAS_POSTFIX_MAIN_CF_TARGET"
	install -m 0644 "$LITE_NAS_POSTFIX_MASTER_CF_SOURCE" "$LITE_NAS_POSTFIX_MASTER_CF_TARGET"

	if [ ! -f "$LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET" ]; then
		if [ ! -f "$LITE_NAS_POSTFIX_AUTH_CONFIG_SOURCE" ]; then
			log.error "Missing Postfix authentication template: $LITE_NAS_POSTFIX_AUTH_CONFIG_SOURCE"
			exit 1
		fi

		install -m 0600 "$LITE_NAS_POSTFIX_AUTH_CONFIG_SOURCE" "$LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET"
	fi
}

deploy.postfix.applyAuthenticationConfig() {
	local postfix_relay_host=""
	local postfix_relay_port=""
	local postfix_relay_username=""
	local postfix_relay_password=""
	local postfix_relay_tls_level="encrypt"
	local sasl_passwd_db=""

	if [ -f "$LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET" ]; then
		# shellcheck disable=SC1090
		source "$LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET"
	fi

	sasl_passwd_db="${LITE_NAS_POSTFIX_SASL_PASSWD_TARGET}.db"

	if [ -z "$postfix_relay_host" ] && [ -z "$postfix_relay_port" ] && [ -z "$postfix_relay_username" ] && [ -z "$postfix_relay_password" ]; then
		postconf -X relayhost >/dev/null 2>&1 || true
		postconf -X smtp_sasl_auth_enable >/dev/null 2>&1 || true
		postconf -X smtp_sasl_password_maps >/dev/null 2>&1 || true
		postconf -X smtp_sasl_security_options >/dev/null 2>&1 || true
		postconf -X smtp_sasl_tls_security_options >/dev/null 2>&1 || true
		rm -f "$LITE_NAS_POSTFIX_SASL_PASSWD_TARGET" "$sasl_passwd_db"
		return 0
	fi

	if [ -z "$postfix_relay_host" ] || [ -z "$postfix_relay_port" ] || [ -z "$postfix_relay_username" ] || [ -z "$postfix_relay_password" ]; then
		log.error "Incomplete Postfix relay configuration in $LITE_NAS_POSTFIX_AUTH_CONFIG_TARGET"
		exit 1
	fi

	printf '[%s]:%s %s:%s\n' \
		"$postfix_relay_host" \
		"$postfix_relay_port" \
		"$postfix_relay_username" \
		"$postfix_relay_password" >"$LITE_NAS_POSTFIX_SASL_PASSWD_TARGET"
	chown root:root "$LITE_NAS_POSTFIX_SASL_PASSWD_TARGET"
	chmod 0600 "$LITE_NAS_POSTFIX_SASL_PASSWD_TARGET"
	postmap "hash:$LITE_NAS_POSTFIX_SASL_PASSWD_TARGET"
	chown root:root "$sasl_passwd_db"
	chmod 0600 "$sasl_passwd_db"

	postconf -e "relayhost = [${postfix_relay_host}]:${postfix_relay_port}"
	postconf -e 'smtp_sasl_auth_enable = yes'
	postconf -e "smtp_sasl_password_maps = hash:${LITE_NAS_POSTFIX_SASL_PASSWD_TARGET}"
	postconf -e 'smtp_sasl_security_options = noanonymous'
	postconf -e 'smtp_sasl_tls_security_options = noanonymous'
	postconf -e "smtp_tls_security_level = ${postfix_relay_tls_level}"
}

deploy.postfix.validateConfig() {
	postconf -n >/dev/null
	postconf inet_interfaces >/dev/null
}

deploy.postfix.enableAndStart() {
	systemctl daemon-reload
	systemctl enable "$LITE_NAS_POSTFIX_SERVICE_NAME.service"
	systemctl restart "$LITE_NAS_POSTFIX_SERVICE_NAME.service"
}

deploy.postfix.deploy() {
	local should_start="${1:-1}"
	deploy.postfix.installConfig
	deploy.postfix.applyAuthenticationConfig
	deploy.postfix.validateConfig
	if [ "$should_start" = "1" ]; then
		deploy.postfix.enableAndStart
		return 0
	fi

	log.info "Skipping Postfix enable/start."
}
