#!/usr/bin/env bash
# Contract tests for atomic openapi fetch (F03 US2).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPT="$ROOT/scripts/openapi_fetch.sh"
WORKDIR=$(mktemp -d)
trap 'kill ${HTTP_PID:-} 2>/dev/null || true; wait ${HTTP_PID:-} 2>/dev/null || true; rm -rf "$WORKDIR"' EXIT

fail() { echo "FAIL: $*" >&2; exit 1; }
pass() { echo "PASS: $*"; }

chmod +x "$SCRIPT"

# --- T041: Makefile must not curl -o directly onto finals ---
if grep -E 'curl[[:space:]].*-o[[:space:]]+docs/api/openapi\.(json|yaml)' "$ROOT/Makefile" >/dev/null; then
  fail "Makefile still curls directly onto docs/api/openapi.{json,yaml}"
fi
if ! grep -E 'scripts/openapi_fetch\.sh' "$ROOT/Makefile" >/dev/null; then
  fail "Makefile openapi-fetch should invoke scripts/openapi_fetch.sh"
fi
pass "Makefile does not curl -o onto finals; uses helper"

# --- Local fixture HTTP server ---
FIX="$WORKDIR/fix"
mkdir -p "$FIX"
echo '{"openapi":"3.0.0","paths":{}}' >"$FIX/good.json"
echo 'openapi: "3.0.0"' >"$FIX/good.yaml"
echo 'STALE_JSON' >"$FIX/stale.json"
echo 'STALE_YAML' >"$FIX/stale.yaml"

PORT_FILE="$WORKDIR/port"
python3 - "$FIX" "$PORT_FILE" <<'PY' &
import http.server, socketserver, sys, os
os.chdir(sys.argv[1])
httpd = socketserver.TCPServer(("127.0.0.1", 0), http.server.SimpleHTTPRequestHandler)
open(sys.argv[2], "w").write(str(httpd.server_address[1]))
httpd.serve_forever()
PY
HTTP_PID=$!
for _ in $(seq 1 100); do
  [[ -s "$PORT_FILE" ]] && break
  sleep 0.05
done
[[ -s "$PORT_FILE" ]] || fail "HTTP fixture server did not start"
PORT=$(cat "$PORT_FILE")
BASE="http://127.0.0.1:${PORT}"

JSON_OUT="$WORKDIR/out/openapi.json"
YAML_OUT="$WORKDIR/out/openapi.yaml"
mkdir -p "$WORKDIR/out"

# Success path
cp "$FIX/stale.json" "$JSON_OUT"
cp "$FIX/stale.yaml" "$YAML_OUT"
"$SCRIPT" "$BASE/good.json" "$BASE/good.yaml" "$JSON_OUT" "$YAML_OUT"
grep -q openapi "$JSON_OUT" || fail "JSON not updated on success"
grep -q openapi "$YAML_OUT" || fail "YAML not updated on success"
pass "both finals updated when both downloads succeed"

# Second download fails → finals unchanged
cp "$FIX/stale.json" "$JSON_OUT"
cp "$FIX/stale.yaml" "$YAML_OUT"
set +e
"$SCRIPT" "$BASE/good.json" "$BASE/missing.yaml" "$JSON_OUT" "$YAML_OUT"
rc=$?
set -e
[[ $rc -ne 0 ]] || fail "expected non-zero when second download fails"
[[ "$(cat "$JSON_OUT")" == "STALE_JSON" ]] || fail "JSON should stay unchanged on second-download failure"
[[ "$(cat "$YAML_OUT")" == "STALE_YAML" ]] || fail "YAML should stay unchanged on second-download failure"
pass "finals unchanged when second download fails"

echo "All openapi_fetch atomic tests passed."
