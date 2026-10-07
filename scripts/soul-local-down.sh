#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git rev-parse --show-toplevel)"
docker compose --env-file "$ROOT/.env.soul.local" -f "$ROOT/infra/local/docker-compose.yml" down
