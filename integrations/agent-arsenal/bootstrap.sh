#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(dirname "$0")"
node --check "$ROOT_DIR/gateway.mjs"

printf '%s\n' '{"service":"soul-agent-arsenal","state":"READY_FOR_CONFIG","executionDefault":"disabled"}'
