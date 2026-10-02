#!/usr/bin/env bash
set -euo pipefail

ROOT="$1"
mkdir -p "$ROOT"

for name in superpowers ecc ruflo; do
  case "$name" in
    superpowers)
      url="https://github.com/obra/superpowers.git"
      ref="8ca22dba9a94f28898bbce59f2537ff4d87c747d"
      ;;
    ecc)
      url="https://github.com/affaan-m/ECC.git"
      ref="c05b2d6614f62f6db0047669aa4eefb223d478f9"
      ;;
    ruflo)
      url="https://github.com/ruvnet/ruflo.git"
      ref="27982983ea6cdc4767c0b6614a4ad9a9d9497cce"
      ;;
  esac
  dir="$ROOT/$name"
  rm -rf "$dir"
  git clone --filter=blob:none --no-checkout "$url" "$dir"
  git -C "$dir" fetch --depth 1 origin "$ref"
  git -C "$dir" checkout --detach "$ref"
  printf '%s\n' "$ref" > "$dir/.soul-agent-arsenal-pin"
  test "$(git -C "$dir" rev-parse HEAD)" = "$ref"
  printf '%-12s %s\n' "$name" "$(git -C "$dir" rev-parse HEAD)"
done
