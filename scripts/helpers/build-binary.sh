#!/usr/bin/env bash

if [ -n "${LITE_NAS_BUILD_BINARY_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_BUILD_BINARY_LOADED=1

build.runGoBinaryScript() {
	local script_name="$1"
	local binary_name="$2"
	local module_dir="$3"
	local default_dir="$4"
	local cgo_enabled="$5"
	local require_gcc="$6"

	shift 6

	local output_path=""
	local target_arch=""

	usage() {
		cat <<MSG
Usage: ${script_name} [options]

Options:
  --output PATH       Output binary path. Defaults to .build/${default_dir}/linux-<arch>/${binary_name}
  -h, --help          Show this help.
MSG
	}

	args.parse "$@"
	if ! args.assertKnown output help h; then
		log.error "Unknown option: --$(args.unknownKeys output help h | head -n 1)"
		usage >&2
		exit 64
	fi
	if args.has h || args.has help; then
		usage
		exit 0
	fi
	if args.has output && ! output_path="$(args.require_arg output)"; then
		log.error "Missing value for --output"
		usage >&2
		exit 64
	fi

	log.requireCommand "go" "Install Go and retry."
	if [ "$require_gcc" = "1" ]; then
		log.requireCommand "gcc" "Install a C compiler and retry."
	fi

	if [ "$cgo_enabled" = "1" ]; then
		target_arch="$(go env GOARCH)"
	else
		target_arch="$(build.resolveTargetArch)"
	fi

	output_path="$(build.prepareOutputPath "$output_path" "$default_dir" "$binary_name")"
	mkdir -p "$(dirname "$output_path")"

	build.prepareGoCache
	build.goBinary "$binary_name" "$module_dir" "$output_path" "$cgo_enabled" "$target_arch"
}
