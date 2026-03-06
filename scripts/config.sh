#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
INPUT="${1:-config.json}"
OUTPUT="${2:-config.complete.json}"

cd "$ROOT"
echo "validating and exporting $INPUT -> $OUTPUT"
cue export ./schema/ "$INPUT" --out json > "$OUTPUT"
echo "done"
