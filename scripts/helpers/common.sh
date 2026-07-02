#!/usr/bin/env bash

if [ -n "${LITE_NAS_COMMON_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_COMMON_LOADED=1

COMMON_SOURCE="${BASH_SOURCE[0]}"
if command -v realpath >/dev/null 2>&1; then
	COMMON_SOURCE="$(realpath "$COMMON_SOURCE")"
fi

COMMON_DIR="$(cd "$(dirname "$COMMON_SOURCE")" && pwd)"
export LITE_NAS_REPO_ROOT="${LITE_NAS_REPO_ROOT:-$(cd "$COMMON_DIR/../.." && pwd)}"

# shellcheck disable=SC1091
source "$COMMON_DIR/logger.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/sudo-guard.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/tool-paths.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/go-modules.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/args.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/build.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/packaging.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/package-version.sh"

deploy.hasUsableSystemd() {
	command -v systemctl >/dev/null 2>&1 &&
		[ -d /run/systemd/system ] &&
		systemctl daemon-reload >/dev/null 2>&1
}

deploy.hasServiceCommand() {
	command -v service >/dev/null 2>&1
}

deploy.ensureSystemGroup() {
	local group_name="$1"
	local group_description="${2:-system group}"

	if getent group "$group_name" >/dev/null 2>&1; then
		return 0
	fi

	log.info "Creating ${group_description}: $group_name"
	groupadd --system "$group_name"
}

deploy.validateSudoersDropIn() {
	local sudoers_file="$1"

	visudo -c -f "$sudoers_file" >/dev/null
}

deploy.installSudoersTemplate() {
	local source_template="$1"
	local target_file="$2"

	install -d -m 0750 -o root -g root "$(dirname "$target_file")"
	install -m 0440 -o root -g root "$source_template" "$target_file"
	deploy.validateSudoersDropIn "$target_file"
}

# Runs a command with stderr suppressed on success and replayed on failure.
command.quietStderrUnlessFailure() {
	local stderr_file
	local exit_code=0

	stderr_file="$(mktemp)"
	if "$@" 2>"$stderr_file"; then
		rm -f "$stderr_file"
		return 0
	else
		exit_code=$?
	fi

	if [ -s "$stderr_file" ]; then
		cat "$stderr_file" >&2
	fi
	rm -f "$stderr_file"
	return "$exit_code"
}

deploy.enableAndRefreshService() {
	local service="$1"

	if deploy.hasUsableSystemd; then
		systemctl enable "$service"

		if systemctl is-active --quiet "$service"; then
			systemctl restart "$service"
			return 0
		fi

		systemctl start "$service"
		return 0
	fi

	if deploy.hasServiceCommand; then
		if service "$service" restart >/dev/null 2>&1; then
			return 0
		fi

		log.warn "Service manager is present but cannot restart $service in this environment; skipping."
		return 0
	fi

	log.warn "No usable service manager is available for $service; skipping enable/start."
}
