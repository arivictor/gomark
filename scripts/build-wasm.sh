#!/usr/bin/env bash
# Builds the client-side Go runner to WebAssembly and stages the artifacts the
# site serves: a gzip-compressed wasm module and the matching wasm_exec.js glue.
#
# The wasm is committed gzipped (~8 MB vs ~39 MB raw) and served with
# Content-Encoding: gzip, so the repo, the embedded binary, and the client
# download all stay small while the browser decompresses transparently.
#
# Usage: scripts/build-wasm.sh
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

out_dir="public"
mkdir -p "$out_dir"

echo "==> building cmd/wasm -> $out_dir/runner.wasm.gz"
tmp_wasm="$(mktemp)"
trap 'rm -f "$tmp_wasm"' EXIT
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o "$tmp_wasm" ./cmd/wasm
# -n omits the filename and timestamp so the committed archive is deterministic.
gzip -9 -n -c "$tmp_wasm" > "$out_dir/runner.wasm.gz"

echo "==> copying wasm_exec.js from $(go env GOROOT)"
goroot="$(go env GOROOT)"
if [ -f "$goroot/lib/wasm/wasm_exec.js" ]; then
  cp "$goroot/lib/wasm/wasm_exec.js" "$out_dir/wasm_exec.js"
elif [ -f "$goroot/misc/wasm/wasm_exec.js" ]; then
  cp "$goroot/misc/wasm/wasm_exec.js" "$out_dir/wasm_exec.js"
else
  echo "error: wasm_exec.js not found under $goroot" >&2
  exit 1
fi

raw_size="$(stat -c%s "$tmp_wasm" 2>/dev/null || stat -f%z "$tmp_wasm")"
gz_size="$(stat -c%s "$out_dir/runner.wasm.gz" 2>/dev/null || stat -f%z "$out_dir/runner.wasm.gz")"
# awk keeps the script dependency-free (bc is absent on many minimal images).
awk -v raw="$raw_size" -v gz="$gz_size" \
  'BEGIN { printf "==> done: raw %.1f MB, gzipped %.1f MB\n", raw/1048576, gz/1048576 }'
