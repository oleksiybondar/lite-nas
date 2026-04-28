#!/usr/bin/env bash

if [ -n "${LITE_NAS_ARGS_HELPERS_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_ARGS_HELPERS_LOADED=1

readonly LITE_NAS_ARGS_FLAG_VALUE="__LITE_NAS_ARGS_FLAG__"
readonly LITE_NAS_ARGS_EMPTY_VALUE="__LITE_NAS_ARGS_EMPTY__"

args.reset() {
	unset LITE_NAS_ARGS_KEYS
	unset LITE_NAS_ARGS_POSITIONAL
	declare -ga LITE_NAS_ARGS_KEYS=()
	declare -ga LITE_NAS_ARGS_POSITIONAL=()
}

args.keyToVar() {
	local key="$1"
	local sanitized="${key//-/_}"
	sanitized="${sanitized//[^a-zA-Z0-9_]/_}"
	printf '__arg_%s\n' "$sanitized"
}

args.setRaw() {
	local key="$1"
	local value="$2"
	local var_name

	var_name="$(args.keyToVar "$key")"
	printf -v "$var_name" '%s' "$value"
	LITE_NAS_ARGS_KEYS+=("$key")
}

args.raw() {
	local key="$1"
	local var_name

	var_name="$(args.keyToVar "$key")"
	if [ -z "${!var_name+x}" ]; then
		return 1
	fi

	printf '%s\n' "${!var_name}"
}

args.has() {
	local key="$1"
	local var_name

	var_name="$(args.keyToVar "$key")"
	[ -n "${!var_name+x}" ]
}

args.parse() {
	args.reset

	local pending_key=""
	local token=""
	local key=""
	local value=""

	while [ "$#" -gt 0 ]; do
		token="$1"
		shift

		if [ "$token" = "--" ]; then
			if [ -n "$pending_key" ]; then
				args.setRaw "$pending_key" "$LITE_NAS_ARGS_FLAG_VALUE"
				pending_key=""
			fi

			while [ "$#" -gt 0 ]; do
				LITE_NAS_ARGS_POSITIONAL+=("$1")
				shift
			done
			break
		fi

		if [ -n "$pending_key" ]; then
			if [[ "$token" == --* ]]; then
				args.setRaw "$pending_key" "$LITE_NAS_ARGS_FLAG_VALUE"
				pending_key=""
			else
				args.setRaw "$pending_key" "$token"
				pending_key=""
				continue
			fi
		fi

		if [[ "$token" == --*=* ]]; then
			key="${token%%=*}"
			key="${key#--}"
			value="${token#*=}"
			if [ -z "$value" ]; then
				value="$LITE_NAS_ARGS_EMPTY_VALUE"
			fi
			args.setRaw "$key" "$value"
			continue
		fi

		if [[ "$token" == --* ]]; then
			pending_key="${token#--}"
			continue
		fi

		LITE_NAS_ARGS_POSITIONAL+=("$token")
	done

	if [ -n "$pending_key" ]; then
		args.setRaw "$pending_key" "$LITE_NAS_ARGS_FLAG_VALUE"
	fi
}

args.assertKnown() {
	local known_key
	local seen_key

	for seen_key in "${LITE_NAS_ARGS_KEYS[@]}"; do
		for known_key in "$@"; do
			if [ "$seen_key" = "$known_key" ]; then
				continue 2
			fi
		done

		return 1
	done

	return 0
}

args.unknownKeys() {
	local known_key
	local seen_key

	for seen_key in "${LITE_NAS_ARGS_KEYS[@]}"; do
		for known_key in "$@"; do
			if [ "$seen_key" = "$known_key" ]; then
				continue 2
			fi
		done

		printf '%s\n' "$seen_key"
	done
}

args.parse_arg() {
	local key="$1"
	local default_value="${2:-}"
	local value=""

	if ! value="$(args.raw "$key")"; then
		printf '%s\n' "$default_value"
		return 0
	fi

	if [ "$value" = "$LITE_NAS_ARGS_FLAG_VALUE" ] || [ "$value" = "$LITE_NAS_ARGS_EMPTY_VALUE" ]; then
		printf '%s\n' "$default_value"
		return 0
	fi

	printf '%s\n' "$value"
}

args.require_arg() {
	local key="$1"
	local value=""

	if ! value="$(args.raw "$key")"; then
		return 1
	fi

	if [ "$value" = "$LITE_NAS_ARGS_FLAG_VALUE" ] || [ "$value" = "$LITE_NAS_ARGS_EMPTY_VALUE" ]; then
		return 1
	fi

	printf '%s\n' "$value"
}

args.parse_bool_arg() {
	local key="$1"
	local default_value="${2:-false}"
	local value=""

	if ! value="$(args.raw "$key")"; then
		value="$default_value"
	fi

	case "$value" in
	"$LITE_NAS_ARGS_FLAG_VALUE" | true | 1)
		printf 'true\n'
		;;
	false | 0)
		printf 'false\n'
		;;
	"$LITE_NAS_ARGS_EMPTY_VALUE" | "")
		if [ "$default_value" = "true" ] || [ "$default_value" = "1" ]; then
			printf 'true\n'
		else
			printf 'false\n'
		fi
		;;
	*)
		if [ "$default_value" = "true" ] || [ "$default_value" = "1" ]; then
			printf 'true\n'
		else
			printf 'false\n'
		fi
		;;
	esac
}
