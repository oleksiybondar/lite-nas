#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

GRAMMAR_FILE="$LITE_NAS_REPO_ROOT/shared/go/parsers/acl/getfacl/grammar/Getfacl.g4"
OUTPUT_DIR="$LITE_NAS_REPO_ROOT/shared/go/parsers/generated/acl/getfacl"
PACKAGE_NAME="getfacl"
ANTLR_VERSION="4.13.2"
ANTLR_JAR_DIR="$LITE_NAS_REPO_ROOT/.bin/antlr4"
ANTLR_JAR_FILE="$ANTLR_JAR_DIR/antlr-$ANTLR_VERSION-complete.jar"
ANTLR_JAR_URL="https://www.antlr.org/download/antlr-$ANTLR_VERSION-complete.jar"

log.requireCommand "java" "Run ./scripts/install-dev-dependencies.sh to install Java runtime."
log.requireCommand "curl" "Install curl and retry."

if [ ! -f "$GRAMMAR_FILE" ]; then
	log.error "Missing grammar file: $GRAMMAR_FILE"
	exit 1
fi

log.pushTask "Generating ANTLR4 parser for getfacl"
mkdir -p "$OUTPUT_DIR"
mkdir -p "$ANTLR_JAR_DIR"

if [ ! -f "$ANTLR_JAR_FILE" ]; then
	log.info "Downloading ANTLR $ANTLR_VERSION tool jar"
	curl -fsSL "$ANTLR_JAR_URL" -o "$ANTLR_JAR_FILE"
fi

rm -f "$OUTPUT_DIR"/*.go "$OUTPUT_DIR"/*.tokens "$OUTPUT_DIR"/*.interp

java -jar "$ANTLR_JAR_FILE" \
	-Xexact-output-dir \
	-Dlanguage=Go \
	-package "$PACKAGE_NAME" \
	-o "$OUTPUT_DIR" \
	"$GRAMMAR_FILE"
log.popTask

log.info "Generated parser files in $OUTPUT_DIR"
