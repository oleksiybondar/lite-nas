#!/usr/bin/env bash

deploy.entrypoint.usage() {
	local name="$1"
	local include_no_start="$2"
	echo "Usage: ${name} [options]"
	echo
	echo "Options:"
	echo "  --binary PATH       Install an existing binary instead of building one."
	if [ "$include_no_start" = "1" ]; then
		echo "  --no-start          Install files but do not enable or start the service."
	fi
	echo "  --skip-bootstrap    Install files without running LiteNAS bootstrap first."
	echo "  -h, --help          Show this help."
}

deploy.entrypoint.run() {
	local script_name="$1"
	local include_no_start="$2"
	local artifact_name="$3"
	local build_script="$4"
	local require_tools_fn="$5"
	local deploy_fn="$6"
	local deploy_task_label="$7"
	local completion_message="$8"

	shift 8

	local binary_path=""
	local should_start=1
	local should_bootstrap=1

	while [ "$#" -gt 0 ]; do
		case "$1" in
		--binary)
			if [ -z "${2:-}" ]; then
				log.error "Missing value for --binary"
				deploy.entrypoint.usage "$script_name" "$include_no_start" >&2
				exit 2
			fi
			binary_path="$2"
			shift 2
			;;
		--no-start)
			if [ "$include_no_start" != "1" ]; then
				log.error "Unknown option: $1"
				deploy.entrypoint.usage "$script_name" "$include_no_start" >&2
				exit 2
			fi
			should_start=0
			shift
			;;
		--skip-bootstrap)
			should_bootstrap=0
			shift
			;;
		-h | --help)
			deploy.entrypoint.usage "$script_name" "$include_no_start"
			exit 0
			;;
		*)
			log.error "Unknown option: $1"
			deploy.entrypoint.usage "$script_name" "$include_no_start" >&2
			exit 2
			;;
		esac
	done

	sudo.guard.requireRoot "$script_name"
	deploy.liteNAS.requireTools
	"$require_tools_fn"

	if [ -z "$binary_path" ]; then
		tmp_dir="$(mktemp -d)"
		trap 'rm -rf "$tmp_dir"' EXIT
		binary_path="$tmp_dir/$artifact_name"
		log.pushTask "Building ${artifact_name} deploy artifact"
		"$ENTRYPOINT_DIR/$build_script" "--output=$binary_path"
		log.popTask
	fi

	if [ "$should_bootstrap" -eq 1 ]; then
		log.pushTask "Bootstrapping LiteNAS prerequisites"
		deploy.liteNAS.bootstrap 1
		log.popTask
	else
		log.info "Skipping LiteNAS bootstrap."
	fi

	log.pushTask "$deploy_task_label"
	if [ "$include_no_start" = "1" ]; then
		"$deploy_fn" "$binary_path" "$should_start"
	else
		"$deploy_fn" "$binary_path"
	fi
	log.popTask

	log.info "$completion_message"
}
