#!/usr/bin/env bash

if [ -n "${LITE_NAS_BUILD_HELPERS_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_BUILD_HELPERS_LOADED=1

build.resolveTargetArch() {
	go env GOARCH
}

build.prepareOutputPath() {
	local output_path="$1"
	local default_dir="$2"
	local binary_name="$3"

	if [ -n "$output_path" ]; then
		printf '%s\n' "$output_path"
		return 0
	fi

	local target_arch
	target_arch="$(build.resolveTargetArch)"
	printf '%s/.build/%s/linux-%s/%s\n' \
		"$LITE_NAS_REPO_ROOT" \
		"$default_dir" \
		"$target_arch" \
		"$binary_name"
}

build.prepareGoCache() {
	export GOCACHE="${GOCACHE:-${TMPDIR:-/tmp}/lite-nas-go-build}"
	mkdir -p "$GOCACHE"
}

build.goBinary() {
	local task_name="$1"
	local module_dir="$2"
	local output_path="$3"
	local cgo_enabled="$4"
	local target_arch="$5"

	log.pushTask "Building ${task_name} binary for linux/${target_arch}"
	(
		cd "$module_dir" || exit 1
		CGO_ENABLED="$cgo_enabled" GOOS=linux GOARCH="$target_arch" go build \
			-ldflags="-s -w" \
			-o "$output_path" .
	)
	log.popTask

	log.info "Built binary: $output_path"
}
