#!/usr/bin/env bash
# Atomic OpenAPI snapshot fetch: download JSON+YAML to temps, then mv both.
# Usage: openapi_fetch.sh <json_url> <yaml_url> <json_out> <yaml_out>
set -euo pipefail

if [[ $# -ne 4 ]]; then
  echo "usage: $0 <json_url> <yaml_url> <json_out> <yaml_out>" >&2
  exit 2
fi

JSON_URL=$1
YAML_URL=$2
JSON_OUT=$3
YAML_OUT=$4

OUT_DIR=$(dirname "$JSON_OUT")
mkdir -p "$OUT_DIR"

JSON_TMP=$(mktemp "${OUT_DIR}/.openapi.json.XXXXXX")
YAML_TMP=$(mktemp "${OUT_DIR}/.openapi.yaml.XXXXXX")

cleanup() {
  [[ -n "${JSON_TMP:-}" && -e "$JSON_TMP" ]] && rm -f "$JSON_TMP"
  [[ -n "${YAML_TMP:-}" && -e "$YAML_TMP" ]] && rm -f "$YAML_TMP"
}
trap cleanup EXIT

curl -fsSL "$JSON_URL" -o "$JSON_TMP"
curl -fsSL "$YAML_URL" -o "$YAML_TMP"

mv "$JSON_TMP" "$JSON_OUT"
JSON_TMP=
mv "$YAML_TMP" "$YAML_OUT"
YAML_TMP=
trap - EXIT
