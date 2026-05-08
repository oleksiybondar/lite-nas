#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_UFW_DEFAULT_SOURCE="${LITE_NAS_UFW_DEFAULT_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/default/ufw}"
readonly LITE_NAS_UFW_DEFAULT_TARGET="${LITE_NAS_UFW_DEFAULT_TARGET:-/etc/default/ufw}"
readonly LITE_NAS_UFW_CONF_SOURCE="${LITE_NAS_UFW_CONF_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/ufw/ufw.conf}"
readonly LITE_NAS_UFW_CONF_TARGET="${LITE_NAS_UFW_CONF_TARGET:-/etc/ufw/ufw.conf}"

deploy.ufw.requireTools() {
	local tool
	local tools=(
		awk
		grep
		install
		sed
		sort
		yes
	)

	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required firewall tooling and retry."
	done
}

deploy.ufw.installConfig() {
	if [ ! -f "$LITE_NAS_UFW_DEFAULT_SOURCE" ]; then
		log.error "Missing UFW default config source: $LITE_NAS_UFW_DEFAULT_SOURCE"
		exit 1
	fi

	if [ ! -f "$LITE_NAS_UFW_CONF_SOURCE" ]; then
		log.error "Missing UFW runtime config source: $LITE_NAS_UFW_CONF_SOURCE"
		exit 1
	fi

	install -d -m 0755 /etc/default /etc/ufw
	install -m 0644 "$LITE_NAS_UFW_DEFAULT_SOURCE" "$LITE_NAS_UFW_DEFAULT_TARGET"
	install -m 0644 "$LITE_NAS_UFW_CONF_SOURCE" "$LITE_NAS_UFW_CONF_TARGET"
}

deploy.ufw.deleteNumberedRule() {
	local rule_number="$1"

	yes | ufw delete "$rule_number" >/dev/null
}

deploy.ufw.removeHTTPAllowRules() {
	local status_output
	local rule_numbers
	local rule_number

	status_output="$(ufw status numbered || true)"
	rule_numbers="$(printf '%s\n' "$status_output" |
		awk '/80\/tcp/ && /ALLOW IN/ {
			if (match($0, /^\[[[:space:]]*[0-9]+]/)) {
				rule = substr($0, RSTART + 1, RLENGTH - 2)
				gsub(/[[:space:]]/, "", rule)
				print rule
			}
		}' |
		sort -rn)"

	if [ -z "$rule_numbers" ]; then
		return 0
	fi

	log.pushTask "Removing UFW allow rules for 80/tcp"
	while IFS= read -r rule_number; do
		if [ -z "$rule_number" ]; then
			continue
		fi
		deploy.ufw.deleteNumberedRule "$rule_number"
	done <<<"$rule_numbers"
	log.popTask
}

deploy.ufw.applyPolicy() {
	log.requireCommand "ufw" "Install UFW and retry."
	log.pushTask "Applying LiteNAS UFW policy"
	ufw default deny incoming
	ufw default allow outgoing
	deploy.ufw.removeHTTPAllowRules
	ufw allow 443/tcp
	ufw deny 80/tcp
	ufw --force enable
	ufw reload
	log.popTask
}

deploy.ufw.deploy() {
	deploy.ufw.installConfig
	deploy.ufw.applyPolicy
}
