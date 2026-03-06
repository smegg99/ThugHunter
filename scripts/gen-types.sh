#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SCHEMA_DIR="$ROOT/schema"
OUTPUT_DIR="$ROOT/common/config"
GEN_FILE="cue_types_schema_gen.go"
TARGET_FILE="cue_types_gen.go"

echo "generating Go types from CUE schema..."
cd "$SCHEMA_DIR"
cue exp gengotypes .

if [[ ! -f "$GEN_FILE" ]]; then
  echo "error: expected $SCHEMA_DIR/$GEN_FILE not found" >&2
  exit 1
fi

mv "$GEN_FILE" "$OUTPUT_DIR/$TARGET_FILE"
echo "done: $OUTPUT_DIR/$TARGET_FILE"
