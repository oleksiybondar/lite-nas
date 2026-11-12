#!/usr/bin/env bash

if [ -n "${LITE_NAS_PACKAGE_VERSION_HELPERS_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_PACKAGE_VERSION_HELPERS_LOADED=1

# shellcheck disable=SC1091
source "$COMMON_DIR/../config/version.conf"

readonly LITE_NAS_LOCAL_BUILD_COUNTER_FILE="${LITE_NAS_REPO_ROOT}/scripts/config/package-build-counter.txt"

packageVersion.requireBaseVersion() {
	if [ -z "${LITE_NAS_BASE_VERSION:-}" ]; then
		log.error "LITE_NAS_BASE_VERSION is not configured."
		exit 1
	fi
}

packageVersion.localCounterFile() {
	printf '%s\n' "$LITE_NAS_LOCAL_BUILD_COUNTER_FILE"
}

packageVersion.readLocalBuildCounter() {
	local counter_file=""
	local counter_value=""

	packageVersion.requireBaseVersion
	counter_file="$(packageVersion.localCounterFile)"

	if [ ! -f "$counter_file" ]; then
		printf '1\n'
		return 0
	fi

	counter_value="$(tr -d '[:space:]' <"$counter_file")"
	if [ -z "$counter_value" ]; then
		log.error "Local package build counter is empty: $counter_file"
		exit 1
	fi
	if ! [[ "$counter_value" =~ ^[0-9]+$ ]]; then
		log.error "Local package build counter must be a non-negative integer: $counter_file"
		exit 1
	fi

	printf '%s\n' "$counter_value"
}

packageVersion.bumpLocalBuildCounter() {
	local counter_file=""
	local next_counter=""

	counter_file="$(packageVersion.localCounterFile)"
	next_counter=$(($(packageVersion.readLocalBuildCounter) + 1))
	printf '%s\n' "$next_counter" >"$counter_file"
}

packageVersion.localAlpha() {
	local build_number="${1:-}"

	if [ -z "$build_number" ]; then
		build_number="$(packageVersion.readLocalBuildCounter)"
	fi

	printf '%s~alpha.%s\n' "$LITE_NAS_BASE_VERSION" "$build_number"
}

packageVersion.prBeta() {
	local run_number="$1"
	local git_sha="$2"

	printf '%s~beta.%s.%s\n' \
		"$LITE_NAS_BASE_VERSION" \
		"$run_number" \
		"${git_sha:0:7}"
}

packageVersion.releaseRel() {
	local run_number="$1"
	local git_sha="$2"

	printf '%s+rel.%s.%s\n' \
		"$LITE_NAS_BASE_VERSION" \
		"$run_number" \
		"${git_sha:0:7}"
}
