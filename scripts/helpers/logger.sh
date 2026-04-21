#!/usr/bin/env bash

if [ -n "${LITE_NAS_LOGGER_LOADED:-}" ]; then
  return 0
fi
readonly LITE_NAS_LOGGER_LOADED=1

export LOG_TASK_DEPTH="${LOG_TASK_DEPTH:-0}"

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  readonly LOG_COLOR_INFO=$'\033[0;36m'
  readonly LOG_COLOR_DEBUG=$'\033[0;90m'
  readonly LOG_COLOR_WARN=$'\033[0;33m'
  readonly LOG_COLOR_ERROR=$'\033[0;31m'
  readonly LOG_COLOR_RESET=$'\033[0m'
else
  readonly LOG_COLOR_INFO=""
  readonly LOG_COLOR_DEBUG=""
  readonly LOG_COLOR_WARN=""
  readonly LOG_COLOR_ERROR=""
  readonly LOG_COLOR_RESET=""
fi

log.indent() {
  local depth="$LOG_TASK_DEPTH"
  local indent=""

  while [ "$depth" -gt 0 ]; do
    indent="${indent}  "
    depth=$((depth - 1))
  done

  printf '%s' "$indent"
}

log.write() {
  local color="$1"
  local level="$2"
  local message="$3"
  local stream="${4:-1}"
  local indent

  indent="$(log.indent)"
  printf '%s%s[%s]%s %s\n' "$indent" "$color" "$level" "$LOG_COLOR_RESET" "$message" >&"$stream"
}

log.info() {
  log.write "$LOG_COLOR_INFO" "INFO" "$*" 1
}

log.warn() {
  log.write "$LOG_COLOR_WARN" "WARN" "$*" 1
}

log.error() {
  log.write "$LOG_COLOR_ERROR" "ERROR" "$*" 2
}

log.debug() {
  if [ "${LOG_DEBUG:-0}" = "1" ]; then
    log.write "$LOG_COLOR_DEBUG" "DEBUG" "$*" 1
  fi
}

log.pushTask() {
  local indent

  indent="$(log.indent)"
  printf '%s> %s\n' "$indent" "$*"
  LOG_TASK_DEPTH=$((LOG_TASK_DEPTH + 1))
  export LOG_TASK_DEPTH
}

log.popTask() {
  if [ "$LOG_TASK_DEPTH" -gt 0 ]; then
    LOG_TASK_DEPTH=$((LOG_TASK_DEPTH - 1))
    export LOG_TASK_DEPTH
  fi
}
