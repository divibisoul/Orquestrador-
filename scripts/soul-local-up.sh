#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
LOCK="$ROOT/config/soul-repos.lock"
REPO_ROOT="$ROOT/.soul/repos"
SECRETS="$ROOT/.env.soul.local"

command -v git >/dev/null || { echo "BLOCKED: git is required"; exit 2; }
command -v docker >/dev/null || { echo "BLOCKED: docker is required"; exit 2; }
docker compose version >/dev/null || { echo "BLOCKED: docker compose is required"; exit 2; }
mkdir -p "$REPO_ROOT"

while read -r nucleus repo sha; do
  if [[ -z "$nucleus" || "$nucleus" == #* ]]; then continue; fi
  dir="$REPO_ROOT/$(basename "$repo")"
  if [[ ! -d "$dir/.git" ]]; then git clone "https://github.com/$repo.git" "$dir"; fi
  git -C "$dir" fetch --all --tags --prune
  git -C "$dir" checkout --detach "$sha"
done < "$LOCK"

# SOUL-25 external providers are canonical submodules of N07.
# Materialize them before adapter/runtime validation.
if [[ -f "$ROOT/.gitmodules" ]]; then
  git -C "$ROOT" submodule sync --recursive
  git -C "$ROOT" submodule update --init --recursive
  node "$ROOT/scripts/validate-soul-external-submodules.mjs"
fi

if [[ ! -f "$SECRETS" ]]; then
  cat > "$SECRETS" <<'EOF'
SOUL_MESH_SECRET=local-development-secret-change-me
SOUL_MESH_TOKEN=local-development-token-change-me
SARA_API_TOKEN=local-development-sara-token-change-me
N07_APP_TOKEN=local-development-n07-token-change-me
EOF
  chmod 600 "$SECRETS"
fi

set -a
source "$SECRETS"
set +a

node "$ROOT/scripts/validate-soul-blueprint.mjs"
docker compose --env-file "$SECRETS" -f "$ROOT/infra/local/docker-compose.yml" up -d --build
echo "Infrastructure started. Full commissioning requires live E2E capability tests."
