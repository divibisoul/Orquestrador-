#!/usr/bin/env bash
set -euo pipefail

# Materialize only the SOUL top-level gitlinks.
# The exact gitlink SHA is authoritative; avoid shallow-fetch assumptions that may
# fail when the pinned revision is not the upstream default-branch tip.
# Nested repositories remain governed by their own manifests; malformed nested
# metadata is preserved as BLOCKED rather than silently discarded or promoted.
if [[ ! -f .gitmodules ]]; then
  echo 'TOP_LEVEL_SUBMODULES: NONE'
  exit 0
fi

mapfile -t paths < <(
  git config -f .gitmodules --get-regexp '^submodule\\..*\\.path$' |
    awk '{print $2}'
)

for path in "${paths[@]}"; do
  echo "MATERIALIZE_TOP_LEVEL: ${path}"
  git submodule update --init -- "${path}"
done

# Autogenesis has a verified nested gitlink for the public HLE dataset revision
# 5a81a4c7271a2a2a312b9a690f0c2fde837e4c29 but no .gitmodules metadata in that
# upstream commit. Repair only the CI-local URL configuration; do not edit or
# vendor the upstream repository and do not invent runtime success.
if [[ -d integrations/external/autogenesis/.git ]]; then
  if git -C integrations/external/autogenesis ls-tree -r --full-tree HEAD |
      awk '$1 == "160000" {print $4}' |
      grep -qx 'datasets/hle'; then
    if ! git -C integrations/external/autogenesis config --get submodule.datasets/hle.url >/dev/null 2>&1; then
      git -C integrations/external/autogenesis config submodule.datasets/hle.url \
        'https://huggingface.co/datasets/cais/hle'
      echo 'NESTED_SUBMODULE_METADATA: PROJECTED/BLOCKED'
      echo 'NESTED_SUBMODULE_REASON: Autogenesis pins HLE but its own .gitmodules is absent.'
      echo 'NESTED_SUBMODULE_PROVENANCE: HLE revision 5a81a4c7271a2a2a312b9a690f0c2fde837e4c29'
      echo 'NESTED_SUBMODULE_ACTION: CI-local URL repair only; upstream remains unchanged.'
    else
      echo 'NESTED_SUBMODULE_METADATA: CONFIGURED'
    fi
  fi
fi

echo "TOP_LEVEL_SUBMODULES: MATERIALIZED"
